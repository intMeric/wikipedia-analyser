// internal/analyzer/user.go
package analyzer

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/intMeric/wikipedia-analyser/internal/client"
	"github.com/intMeric/wikipedia-analyser/internal/models"
)

// UserAnalyzer analyzes Wikipedia user data
type UserAnalyzer struct {
	client *client.WikipediaClient
}

// RevokedAnalysisConfig configuration for revoked contributions analysis
type RevokedAnalysisConfig struct {
	MaxPagesToAnalyze   int  `json:"max_pages_to_analyze"`
	MaxRevisionsPerPage int  `json:"max_revisions_per_page"`
	EnableDeepAnalysis  bool `json:"enable_deep_analysis"`
	RecentDaysOnly      int  `json:"recent_days_only"`
}

// QuickRevertResult result from quick revert analysis
type QuickRevertResult struct {
	HasReverts    bool
	RevertCount   int
	LastRevertAge int // days since last revert
}

// NewUserAnalyzer creates a new user analyzer
func NewUserAnalyzer(client *client.WikipediaClient) *UserAnalyzer {
	return &UserAnalyzer{
		client: client,
	}
}

// GetUserProfile retrieves and analyzes a complete user profile
func (ua *UserAnalyzer) GetUserProfile(username string) (*models.UserProfile, error) {
	// 1. Get basic information
	userInfo, err := ua.client.GetUserInfo(username)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve user info: %w", err)
	}

	// 2. Get recent contributions with tags
	contributions, err := ua.client.GetUserContributionsWithTags(username, 100)
	if err != nil {
		// Fallback to standard contributions if tags are not available
		contributions, err = ua.client.GetUserContributions(username, 100)
		if err != nil {
			return nil, fmt.Errorf("unable to retrieve contributions: %w", err)
		}
	}

	// 3. Create basic profile
	profile := &models.UserProfile{
		Username:       userInfo.Name,
		UserID:         userInfo.UserID,
		EditCount:      userInfo.EditCount,
		Groups:         userInfo.Groups,
		ImplicitGroups: userInfo.ImplicitGroups,
		RightsInfo:     userInfo.Rights,
		Language:       ua.client.Language(),
		RetrievedAt:    time.Now(),
	}

	// 4. Parse registration date
	if userInfo.Registration != "" {
		regDate, err := time.Parse("2006-01-02T15:04:05Z", userInfo.Registration)
		if err == nil {
			profile.RegistrationDate = &regDate
		}
	}

	// 5. Analyze block information
	profile.BlockInfo = ua.analyzeBlockInfo(userInfo)

	// 6. Convert and analyze contributions
	profile.RecentContribs = ua.convertContributions(contributions)
	profile.TopPages = ua.analyzeTopPages(contributions)
	profile.ActivityStats = ua.analyzeActivity(contributions, profile.RegistrationDate)

	// 7. Analyze revoked contributions (NEW STEP)
	fmt.Printf("🔍 Analyzing revoked contributions for %s...\n", username)

	// Use default configuration for revoked analysis
	config := GetDefaultRevokedAnalysisConfig()
	revokedContribs, err := ua.analyzeRevokedContributions(username, contributions, config)
	if err != nil {
		fmt.Printf("⚠️ Failed to analyze revoked contributions: %v\n", err)
		// Continue without this analysis rather than failing completely
		revokedContribs = []models.RevokedContribution{}
	}

	// Calculate revoked contribution statistics
	profile.RevokedContribs = revokedContribs
	profile.RevokedCount = len(revokedContribs)

	if len(contributions) > 0 {
		profile.RevokedRatio = float64(profile.RevokedCount) / float64(len(contributions))
	}

	// Analyze who reverts this user most often
	revertedByUsers := make(map[string]int)
	for _, revoked := range revokedContribs {
		revertedByUsers[revoked.RevokedBy]++
	}
	profile.RevertedByUsers = revertedByUsers

	// Mark revoked contributions in the recent contributions list
	ua.markRevokedContributions(profile)

	// 8. Calculate suspicion score (now with revocation data)
	profile.SuspicionScore, profile.SuspicionFlags = ua.calculateSuspicionScore(profile)

	return profile, nil
}

// analyzeBlockInfo analyzes block information
func (ua *UserAnalyzer) analyzeBlockInfo(userInfo *models.WikiUserInfo) *models.BlockInfo {
	blockInfo := &models.BlockInfo{
		Blocked: false,
	}

	if userInfo.BlockExpiry != "" && userInfo.BlockExpiry != "infinity" {
		blockInfo.Blocked = true
		blockInfo.BlockedBy = userInfo.BlockedBy
		blockInfo.Reason = userInfo.BlockReason

		// Parse block dates if possible
		if blockExpiry, err := time.Parse("2006-01-02T15:04:05Z", userInfo.BlockExpiry); err == nil {
			blockInfo.BlockEnd = blockExpiry
		}
	}

	return blockInfo
}

// convertContributions converts API contributions to internal model
func (ua *UserAnalyzer) convertContributions(wikiContribs []models.WikiContribution) []models.Contribution {
	contributions := make([]models.Contribution, 0, len(wikiContribs))

	for _, wc := range wikiContribs {
		timestamp, _ := time.Parse("2006-01-02T15:04:05Z", wc.Timestamp)

		contribution := models.Contribution{
			RevID:     wc.RevID,
			PageTitle: wc.Title,
			Namespace: wc.NS,
			Timestamp: timestamp,
			Comment:   wc.Comment,
			SizeDiff:  wc.SizeDiff,
			IsMinor:   wc.Minor == "true",
			IsTop:     wc.Top == "true",
			PageID:    wc.PageID,
		}

		contributions = append(contributions, contribution)
	}

	return contributions
}

// analyzeTopPages analyzes most edited pages
func (ua *UserAnalyzer) analyzeTopPages(contributions []models.WikiContribution) []models.PageEditSummary {
	pageStats := make(map[string]*models.PageEditSummary)

	for _, contrib := range contributions {
		timestamp, _ := time.Parse("2006-01-02T15:04:05Z", contrib.Timestamp)

		key := fmt.Sprintf("%d:%s", contrib.PageID, contrib.Title)

		if summary, exists := pageStats[key]; exists {
			summary.EditCount++
			summary.TotalSizeDiff += contrib.SizeDiff

			if timestamp.After(summary.LastEdit) {
				summary.LastEdit = timestamp
			}
			if timestamp.Before(summary.FirstEdit) {
				summary.FirstEdit = timestamp
			}
		} else {
			pageStats[key] = &models.PageEditSummary{
				PageTitle:     contrib.Title,
				PageID:        contrib.PageID,
				Namespace:     contrib.NS,
				EditCount:     1,
				FirstEdit:     timestamp,
				LastEdit:      timestamp,
				TotalSizeDiff: contrib.SizeDiff,
			}
		}
	}

	// Convert to slice and sort by edit count
	var topPages []models.PageEditSummary
	for _, summary := range pageStats {
		topPages = append(topPages, *summary)
	}

	sort.Slice(topPages, func(i, j int) bool {
		return topPages[i].EditCount > topPages[j].EditCount
	})

	// Return top 15
	lTopPages := 15
	if len(topPages) > lTopPages {
		topPages = topPages[:lTopPages]
	}

	return topPages
}

// analyzeActivity analyzes activity patterns
func (ua *UserAnalyzer) analyzeActivity(contributions []models.WikiContribution, regDate *time.Time) models.ActivityStats {
	stats := models.ActivityStats{
		NamespaceDistrib: make(map[string]int),
		RecentActivity:   make([]models.DailyActivity, 0),
	}

	if len(contributions) == 0 {
		return stats
	}

	// Analyze namespaces
	namespaceNames := map[int]string{
		0:   "Main",
		1:   "Talk",
		2:   "User",
		3:   "User talk",
		4:   "Wikipedia",
		6:   "File",
		10:  "Template",
		14:  "Category",
		100: "Portal",
	}

	hourStats := make(map[int]int)
	dayStats := make(map[string]int)
	dailyActivity := make(map[string]int)

	for _, contrib := range contributions {
		timestamp, _ := time.Parse("2006-01-02T15:04:05Z", contrib.Timestamp)

		// Namespace stats
		nsName := namespaceNames[contrib.NS]
		if nsName == "" {
			nsName = fmt.Sprintf("NS_%d", contrib.NS)
		}
		stats.NamespaceDistrib[nsName]++

		// Hour stats
		hourStats[timestamp.Hour()]++

		// Day stats
		dayName := timestamp.Weekday().String()
		dayStats[dayName]++

		// Daily activity
		dateKey := timestamp.Format("2006-01-02")
		dailyActivity[dateKey]++
	}

	// Find most active hour
	maxHourCount := 0
	for hour, count := range hourStats {
		if count > maxHourCount {
			maxHourCount = count
			stats.MostActiveHour = hour
		}
	}

	// Find most active day
	maxDayCount := 0
	for day, count := range dayStats {
		if count > maxDayCount {
			maxDayCount = count
			stats.MostActiveDay = day
		}
	}

	// Calculate averages
	if regDate != nil {
		daysSinceReg := int(time.Since(*regDate).Hours() / 24)
		if daysSinceReg > 0 {
			stats.DaysActive = daysSinceReg
			stats.AverageEditsPerDay = float64(len(contributions)) / float64(daysSinceReg)
		}
	}

	// Convert daily activity to sorted slice
	var dates []string
	for date := range dailyActivity {
		dates = append(dates, date)
	}
	sort.Strings(dates)

	for _, date := range dates {
		if len(stats.RecentActivity) >= 30 { // Limit to last 30 days
			break
		}

		parsedDate, _ := time.Parse("2006-01-02", date)
		stats.RecentActivity = append(stats.RecentActivity, models.DailyActivity{
			Date:      parsedDate,
			EditCount: dailyActivity[date],
		})
	}

	return stats
}

// analyzeRevokedContributions analyzes revoked contributions of a user
func (ua *UserAnalyzer) analyzeRevokedContributions(username string, contributions []models.WikiContribution, config RevokedAnalysisConfig) ([]models.RevokedContribution, error) {
	var revokedContribs []models.RevokedContribution

	// First, check if contributions already have reverted tags/information
	directRevocations := ua.detectDirectRevocations(contributions)
	revokedContribs = append(revokedContribs, directRevocations...)

	// Limit to recent pages if configured
	sortedContribs := contributions
	if config.RecentDaysOnly > 0 {
		cutoffDate := time.Now().AddDate(0, 0, -config.RecentDaysOnly)
		var filteredContribs []models.WikiContribution

		for _, contrib := range contributions {
			contribTime, _ := time.Parse("2006-01-02T15:04:05Z", contrib.Timestamp)
			if contribTime.After(cutoffDate) {
				filteredContribs = append(filteredContribs, contrib)
			}
		}
		sortedContribs = filteredContribs
	}

	// Limit the number of pages to analyze
	pagesSeen := make(map[string]bool)
	pagesAnalyzed := 0

	for _, contrib := range sortedContribs {
		if pagesAnalyzed >= config.MaxPagesToAnalyze {
			break
		}

		if pagesSeen[contrib.Title] {
			continue
		}

		pagesSeen[contrib.Title] = true
		pagesAnalyzed++

		// Light analysis first
		lightAnalysis := ua.quickRevertCheck(username, contrib.Title)
		if lightAnalysis.HasReverts {
			// Deep analysis only if necessary and enabled
			if config.EnableDeepAnalysis {
				pageReverts, err := ua.deepRevertAnalysis(username, contrib.Title, config.MaxRevisionsPerPage)
				if err == nil {
					revokedContribs = append(revokedContribs, pageReverts...)
				}
			} else {
				// Add just basic information
				basicRevert := models.RevokedContribution{
					PageTitle:     contrib.Title,
					RevertType:    "detected_light",
					RevokedBy:     "detected",
					RevertComment: fmt.Sprintf("Approximately %d reverts detected", lightAnalysis.RevertCount),
					RevokedAt:     time.Now().AddDate(0, 0, -lightAnalysis.LastRevertAge),
				}
				revokedContribs = append(revokedContribs, basicRevert)
			}
		}
	}

	return revokedContribs, nil
}

// detectDirectRevocations tries to detect revocations from contribution metadata
func (ua *UserAnalyzer) detectDirectRevocations(contributions []models.WikiContribution) []models.RevokedContribution {
	var revocations []models.RevokedContribution

	// First, check for revocation tags (most reliable method)
	for _, contrib := range contributions {
		if len(contrib.Tags) > 0 {
			isRevoked := ua.isRevokedByTags(contrib.Tags)
			if isRevoked {
				contribTime, _ := time.Parse("2006-01-02T15:04:05Z", contrib.Timestamp)

				revocation := models.RevokedContribution{
					OriginalContrib: models.Contribution{
						RevID:     contrib.RevID,
						PageTitle: contrib.Title,
						Namespace: contrib.NS,
						Timestamp: contribTime,
						Comment:   contrib.Comment,
						SizeDiff:  contrib.SizeDiff,
					},
					RevokedBy:     "system_detected", // We don't know who specifically reverted from tags
					RevokedAt:     contribTime,       // We don't have exact revert time from tags
					RevertComment: "Detected from revision tags",
					PageTitle:     contrib.Title,
					RevertType:    ua.classifyRevertTypeFromTags(contrib.Tags),
				}

				revocations = append(revocations, revocation)
				continue // Skip temporal analysis for this contribution
			}
		}
	}

	// Then, use temporal and size-based detection as fallback
	for i, contrib := range contributions {
		// Skip if already detected by tags
		alreadyDetected := false
		for _, existing := range revocations {
			if existing.OriginalContrib.RevID == contrib.RevID {
				alreadyDetected = true
				break
			}
		}
		if alreadyDetected {
			continue
		}

		// Look for potential revocations by checking subsequent edits
		if i < len(contributions)-1 {
			nextContrib := contributions[i+1]

			// Check if next contribution might be undoing this one
			contribTime, _ := time.Parse("2006-01-02T15:04:05Z", contrib.Timestamp)
			nextTime, _ := time.Parse("2006-01-02T15:04:05Z", nextContrib.Timestamp)

			// If on same page and within reasonable time frame
			if contrib.Title == nextContrib.Title &&
				nextTime.Sub(contribTime) <= 24*time.Hour &&
				contrib.User != nextContrib.User {

				// If size changes are opposite (indicating undo)
				if (contrib.SizeDiff > 0 && nextContrib.SizeDiff < 0) ||
					(contrib.SizeDiff < 0 && nextContrib.SizeDiff > 0) {

					// Check if sizes are similar (indicating revert)
					sizeDiff := contrib.SizeDiff + nextContrib.SizeDiff
					if sizeDiff < 100 && sizeDiff > -100 { // Within 100 bytes

						revocation := models.RevokedContribution{
							OriginalContrib: models.Contribution{
								RevID:     contrib.RevID,
								PageTitle: contrib.Title,
								Namespace: contrib.NS,
								Timestamp: contribTime,
								Comment:   contrib.Comment,
								SizeDiff:  contrib.SizeDiff,
							},
							RevokedBy:     nextContrib.User,
							RevokedAt:     nextTime,
							RevertComment: nextContrib.Comment,
							PageTitle:     contrib.Title,
							RevertType:    ua.classifyRevertType(nextContrib.Comment),
						}

						revocations = append(revocations, revocation)
					}
				}
			}
		}
	}

	return revocations
}

// isRevokedByTags checks if a contribution was revoked based on its tags
func (ua *UserAnalyzer) isRevokedByTags(tags []string) bool {
	revokedTags := []string{
		"mw-reverted", // English
		"mw-rollback", // Rollback
		"reverted",    // Generic
	}

	// Language-specific tags
	switch ua.client.Language() {
	case "fr":
		revokedTags = append(revokedTags, "révoqué", "annulé")
	case "de":
		revokedTags = append(revokedTags, "rückgängig")
	case "es":
		revokedTags = append(revokedTags, "revertido")
	}

	for _, tag := range tags {
		tagLower := strings.ToLower(tag)
		for _, revokedTag := range revokedTags {
			if strings.Contains(tagLower, revokedTag) {
				return true
			}
		}
	}

	return false
}

// classifyRevertTypeFromTags classifies revert type from tags
func (ua *UserAnalyzer) classifyRevertTypeFromTags(tags []string) string {
	for _, tag := range tags {
		tagLower := strings.ToLower(tag)
		if strings.Contains(tagLower, "vandal") {
			return "vandalism_revert"
		}
		if strings.Contains(tagLower, "rollback") {
			return "rollback"
		}
		if strings.Contains(tagLower, "reverted") || strings.Contains(tagLower, "révoqué") {
			return "generic_revert"
		}
	}
	return "system_detected"
}

// quickRevertCheck performs a light check to detect if there are reverts
func (ua *UserAnalyzer) quickRevertCheck(username string, pageTitle string) QuickRevertResult {
	// Get only the last 20 revisions for a quick check
	revisions, err := ua.client.GetPageRevisions(pageTitle, 20)
	if err != nil {
		return QuickRevertResult{HasReverts: false}
	}

	revertCount := 0
	var lastRevertTime time.Time

	for _, rev := range revisions {
		comment := strings.ToLower(rev.Comment)
		userMentioned := strings.Contains(comment, strings.ToLower(username))
		isRevert := ua.detectRevert(rev.Comment)

		if isRevert && userMentioned {
			revertCount++
			revTime, _ := time.Parse("2006-01-02T15:04:05Z", rev.Timestamp)
			if lastRevertTime.IsZero() || revTime.After(lastRevertTime) {
				lastRevertTime = revTime
			}
		}
	}

	result := QuickRevertResult{
		HasReverts:  revertCount > 0,
		RevertCount: revertCount,
	}

	if !lastRevertTime.IsZero() {
		result.LastRevertAge = int(time.Since(lastRevertTime).Hours() / 24)
	}

	return result
}

// deepRevertAnalysis performs detailed analysis of reverts for a specific page
func (ua *UserAnalyzer) deepRevertAnalysis(username string, pageTitle string, maxRevisions int) ([]models.RevokedContribution, error) {
	// Get page revision history
	pageHistory, err := ua.client.GetPageRevisions(pageTitle, maxRevisions)
	if err != nil {
		return nil, fmt.Errorf("could not get history for %s: %w", pageTitle, err)
	}

	// Find reverts of user's contributions
	userReverts := ua.findUserReverts(username, pageHistory, pageTitle)

	return userReverts, nil
}

// findUserReverts finds reverts of a specific user's contributions
func (ua *UserAnalyzer) findUserReverts(username string, pageHistory []models.WikiRevision, pageTitle string) []models.RevokedContribution {
	var reverts []models.RevokedContribution

	// Create a map of revisions by user
	userRevisions := make(map[int]models.WikiRevision) // revID -> revision

	for _, rev := range pageHistory {
		if rev.User == username {
			userRevisions[rev.RevID] = rev
		}
	}

	// Look for reverts in the history
	for _, rev := range pageHistory {
		if rev.User == username {
			continue // Skip user's own revisions
		}

		// Check if this revision reverts a user's contribution
		revertInfo := ua.detectUserRevert(rev, userRevisions)
		if revertInfo != nil {
			timestamp, _ := time.Parse("2006-01-02T15:04:05Z", rev.Timestamp)
			originalTimestamp, _ := time.Parse("2006-01-02T15:04:05Z", revertInfo.Timestamp)

			revert := models.RevokedContribution{
				OriginalContrib: models.Contribution{
					RevID:     revertInfo.RevID,
					PageTitle: pageTitle,
					Namespace: 0, // Could be retrieved if necessary
					Timestamp: originalTimestamp,
					Comment:   revertInfo.Comment,
					SizeDiff:  revertInfo.Size,
				},
				RevokedBy:     rev.User,
				RevokedAt:     timestamp,
				RevertComment: rev.Comment,
				PageTitle:     pageTitle,
				RevertType:    ua.classifyRevertType(rev.Comment),
			}

			reverts = append(reverts, revert)
		}
	}

	return reverts
}

// detectUserRevert detects if a revision reverts a user's contribution
func (ua *UserAnalyzer) detectUserRevert(revision models.WikiRevision, userRevisions map[int]models.WikiRevision) *models.WikiRevision {
	comment := strings.ToLower(revision.Comment)

	// Keywords indicating a revert by language
	var revertKeywords []string

	switch ua.client.Language() {
	case "fr":
		revertKeywords = []string{"révoqué", "révocation", "annulé", "annulation", "rv", "rvt", "restauré", "rollback", "revert", "undo", "vandalisé", "vandalisme"}
	case "de":
		revertKeywords = []string{"rückgängig", "revert", "undo", "rv", "zurückgesetzt", "vandalismus", "restore", "rollback"}
	case "es":
		revertKeywords = []string{"revertir", "deshacer", "rv", "vandalismo", "restaurar", "revert", "undo", "rollback"}
	default:
		revertKeywords = []string{"revert", "undo", "undid", "rv", "reverted", "restore", "restored", "rollback", "rolled back"}
	}

	isRevert := false
	for _, keyword := range revertKeywords {
		if strings.Contains(comment, keyword) {
			isRevert = true
			break
		}
	}

	if !isRevert {
		return nil
	}

	// Detection methods:

	// 1. Search for revision ID in comment
	for revID := range userRevisions {
		revIDStr := strconv.Itoa(revID)
		if strings.Contains(comment, revIDStr) {
			userRev := userRevisions[revID]
			return &userRev
		}
	}

	// 2. Search for username in revert comment (improved for French)
	if len(userRevisions) > 0 {
		// Get first user revision to extract username
		var firstUserRev models.WikiRevision
		for _, rev := range userRevisions {
			firstUserRev = rev
			break
		}

		username := strings.ToLower(firstUserRev.User)

		// Check if username is mentioned in revert comment
		if strings.Contains(comment, username) {
			// Find the most recent user revision before this revert
			revertTime, _ := time.Parse("2006-01-02T15:04:05Z", revision.Timestamp)

			var closestRev *models.WikiRevision
			var closestTime time.Duration = time.Hour * 24 * 365 // 1 year

			for _, userRev := range userRevisions {
				userRevTime, _ := time.Parse("2006-01-02T15:04:05Z", userRev.Timestamp)
				if userRevTime.Before(revertTime) {
					timeDiff := revertTime.Sub(userRevTime)
					if timeDiff < closestTime {
						closestTime = timeDiff
						closestRev = &userRev
					}
				}
			}

			return closestRev
		}
	}

	// 3. Detection by parentID (if revision returns to a previous version)
	if revision.ParentID > 0 {
		for _, userRev := range userRevisions {
			if userRev.RevID == revision.ParentID {
				return &userRev
			}
		}
	}

	// 4. Temporal detection - if it's a revert and happens shortly after user's edit
	if len(userRevisions) > 0 {
		revertTime, _ := time.Parse("2006-01-02T15:04:05Z", revision.Timestamp)

		// Find user revisions within the last 24 hours
		for _, userRev := range userRevisions {
			userRevTime, _ := time.Parse("2006-01-02T15:04:05Z", userRev.Timestamp)
			if userRevTime.Before(revertTime) {
				timeDiff := revertTime.Sub(userRevTime)
				// If the revert happened within 24 hours and it's clearly a revert
				if timeDiff <= 24*time.Hour && isRevert {
					return &userRev
				}
			}
		}
	}

	return nil
}

// classifyRevertType classifies the type of revert
func (ua *UserAnalyzer) classifyRevertType(comment string) string {
	comment = strings.ToLower(comment)

	switch ua.client.Language() {
	case "fr":
		if strings.Contains(comment, "vandalisme") || strings.Contains(comment, "vandalisé") {
			return "vandalism_revert"
		}
		if strings.Contains(comment, "rollback") {
			return "rollback"
		}
		if strings.Contains(comment, "annulation") || strings.Contains(comment, "annulé") {
			return "undo"
		}
		if strings.Contains(comment, "restauré") || strings.Contains(comment, "restoration") {
			return "restore"
		}
		if strings.Contains(comment, "rv") || strings.Contains(comment, "rvt") {
			return "manual_revert"
		}

	case "de":
		if strings.Contains(comment, "vandalismus") {
			return "vandalism_revert"
		}
		if strings.Contains(comment, "rollback") {
			return "rollback"
		}
		if strings.Contains(comment, "rückgängig") {
			return "undo"
		}

	case "es":
		if strings.Contains(comment, "vandalismo") {
			return "vandalism_revert"
		}
		if strings.Contains(comment, "rollback") {
			return "rollback"
		}
		if strings.Contains(comment, "deshacer") {
			return "undo"
		}

	default: // English
		if strings.Contains(comment, "vandalism") || strings.Contains(comment, "vandal") {
			return "vandalism_revert"
		}
		if strings.Contains(comment, "rollback") {
			return "rollback"
		}
		if strings.Contains(comment, "undo") {
			return "undo"
		}
		if strings.Contains(comment, "restore") {
			return "restore"
		}
		if strings.Contains(comment, "rv") {
			return "manual_revert"
		}
	}

	return "generic_revert"
}

// markRevokedContributions marks contributions that have been revoked
func (ua *UserAnalyzer) markRevokedContributions(profile *models.UserProfile) {
	// Create a map of revoked contributions by RevID
	revokedRevIDs := make(map[int]bool)
	revokedDetails := make(map[int]models.RevokedContribution)

	for _, revoked := range profile.RevokedContribs {
		revokedRevIDs[revoked.OriginalContrib.RevID] = true
		revokedDetails[revoked.OriginalContrib.RevID] = revoked
	}

	// Mark revoked contributions in the recent contributions list
	for i := range profile.RecentContribs {
		if revokedRevIDs[profile.RecentContribs[i].RevID] {
			profile.RecentContribs[i].IsRevoked = true
			details := revokedDetails[profile.RecentContribs[i].RevID]
			profile.RecentContribs[i].RevokedBy = details.RevokedBy
			profile.RecentContribs[i].RevokedAt = details.RevokedAt
			profile.RecentContribs[i].RevertReason = details.RevertComment
		}
	}
}

// calculateSuspicionScore calculates a suspicion score including revoked contributions
func (ua *UserAnalyzer) calculateSuspicionScore(profile *models.UserProfile) (int, []string) {
	score := 0
	flags := []string{}

	// 1. Recent account with high activity
	if profile.RegistrationDate != nil {
		daysSinceReg := int(time.Since(*profile.RegistrationDate).Hours() / 24)
		if daysSinceReg < 30 && profile.EditCount > 100 {
			score += 20
			flags = append(flags, "RECENT_ACCOUNT_HIGH_ACTIVITY")
		}
	}

	// 2. Blocked user
	if profile.BlockInfo != nil && profile.BlockInfo.Blocked {
		score += 30
		flags = append(flags, "USER_BLOCKED")
	}

	// 3. Focus on small number of pages
	if len(profile.TopPages) > 0 && profile.TopPages[0].EditCount > profile.EditCount/2 {
		score += 15
		flags = append(flags, "SINGLE_PAGE_FOCUS")
	}

	// 4. No special groups (unconfirmed user)
	hasSpecialGroups := false
	for _, group := range profile.Groups {
		if group != "*" && group != "user" {
			hasSpecialGroups = true
			break
		}
	}
	if !hasSpecialGroups && profile.EditCount > 50 {
		score += 10
		flags = append(flags, "NO_SPECIAL_GROUPS")
	}

	// 5. Activity only in sensitive namespaces
	sensitiveNamespaces := []string{"Main", "Wikipedia", "Portal"}
	totalSensitive := 0
	totalEdits := 0
	for ns, count := range profile.ActivityStats.NamespaceDistrib {
		totalEdits += count
		for _, sensitive := range sensitiveNamespaces {
			if ns == sensitive {
				totalSensitive += count
			}
		}
	}
	if totalEdits > 0 && float64(totalSensitive)/float64(totalEdits) > 0.9 {
		score += 15
		flags = append(flags, "SENSITIVE_NAMESPACE_FOCUS")
	}

	// 6. Empty or repetitive edit comments
	emptyComments := 0
	for _, contrib := range profile.RecentContribs {
		if strings.TrimSpace(contrib.Comment) == "" {
			emptyComments++
		}
	}
	if len(profile.RecentContribs) > 0 && float64(emptyComments)/float64(len(profile.RecentContribs)) > 0.7 {
		score += 10
		flags = append(flags, "FREQUENT_EMPTY_COMMENTS")
	}

	// 7. High ratio of revoked contributions
	if profile.RevokedRatio > 0.5 { // More than 50% revoked
		score += 30
		flags = append(flags, "VERY_HIGH_REVOKED_RATIO")
	} else if profile.RevokedRatio > 0.3 { // More than 30%
		score += 20
		flags = append(flags, "HIGH_REVOKED_RATIO")
	} else if profile.RevokedRatio > 0.2 { // More than 20%
		score += 10
		flags = append(flags, "MODERATE_REVOKED_RATIO")
	}

	// 8. Many revoked contributions in absolute value
	if profile.RevokedCount > 50 {
		score += 15
		flags = append(flags, "MANY_REVOKED_CONTRIBUTIONS")
	} else if profile.RevokedCount > 20 {
		score += 10
		flags = append(flags, "SOME_REVOKED_CONTRIBUTIONS")
	}

	// 9. Revoked mainly for vandalism
	vandalismReverts := 0
	for _, revoked := range profile.RevokedContribs {
		if revoked.RevertType == "vandalism_revert" {
			vandalismReverts++
		}
	}

	if vandalismReverts > 10 {
		score += 25
		flags = append(flags, "VANDALISM_PATTERN")
	} else if vandalismReverts > 5 {
		score += 15
		flags = append(flags, "SOME_VANDALISM_REVERTS")
	}

	for username, count := range profile.RevertedByUsers {
		if username == "system_detected" || username == "detected" {
			continue
		}

		if count > 5 && profile.RevokedCount > 0 && float64(count)/float64(profile.RevokedCount) > 0.5 {
			score += 15
			conflictFlag := fmt.Sprintf("CONFLICT_WITH_SPECIFIC_USER_%s", username)
			flags = append(flags, conflictFlag)
			break
		}
	}

	// 11. Recently created with many revocations
	if profile.RegistrationDate != nil {
		daysSinceReg := int(time.Since(*profile.RegistrationDate).Hours() / 24)
		if daysSinceReg < 30 && profile.RevokedCount > 10 {
			score += 20
			flags = append(flags, "NEW_ACCOUNT_MANY_REVERTS")
		}
	}

	// Limit score to 100
	if score > 100 {
		score = 100
	}

	return score, flags
}

// detectRevert checks if a comment indicates a revert
func (ua *UserAnalyzer) detectRevert(comment string) bool {
	comment = strings.ToLower(comment)

	// Keywords by language
	var revertKeywords []string

	switch ua.client.Language() {
	case "fr":
		revertKeywords = []string{
			"révoqué", "révocation", "annulé", "annulation", "rv", "rvt",
			"restauré", "restoration", "rollback", "revert", "undo",
			"vandalisé", "vandalisme", "défait", "défaire",
		}
	case "de":
		revertKeywords = []string{
			"rückgängig", "revert", "undo", "rv", "zurückgesetzt",
			"vandalismus", "restore", "rollback",
		}
	case "es":
		revertKeywords = []string{
			"revertir", "deshacer", "rv", "vandalismo", "restaurar",
			"revert", "undo", "rollback",
		}
	default: // English and others
		revertKeywords = []string{
			"revert", "undo", "undid", "rv", "reverted",
			"restore", "restored", "rollback", "rolled back",
			"vandalism", "vandal",
		}
	}

	for _, keyword := range revertKeywords {
		if strings.Contains(comment, keyword) {
			return true
		}
	}
	return false
}

// GetDefaultRevokedAnalysisConfig returns default configuration for revoked analysis
func GetDefaultRevokedAnalysisConfig() RevokedAnalysisConfig {
	return RevokedAnalysisConfig{
		MaxPagesToAnalyze:   10,    // Analyze max 10 pages
		MaxRevisionsPerPage: 50,    // Max 50 revisions per page
		EnableDeepAnalysis:  false, // Light analysis by default
		RecentDaysOnly:      90,    // Only last 90 days
	}
}
