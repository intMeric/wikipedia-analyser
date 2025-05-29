// internal/analyzer/page.go
package analyzer

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/intMeric/wikipedia-analyser/internal/client"

	"github.com/intMeric/wikipedia-analyser/internal/models"
	"github.com/intMeric/wikipedia-analyser/internal/utils"
)

// PageAnalyzer analyzes Wikipedia page data
type PageAnalyzer struct {
	client                *client.WikipediaClient
	numberOfPageRevisions int // Number of revisions to analyze
	numberOfDaysHistory   int // Number of days for detailed history
	numberOfContributors  int // Number of contributors to analyze
}

type PageAnalysisOptions struct {
	NumberOfPageRevisions int // Number of revisions to analyze
	NumberOfDaysHistory   int // Number of days for detailed history
	NumberOfContributors  int // Number of contributors to analyze
}

// NewPageAnalyzer creates a new page analyzer
func NewPageAnalyzer(client *client.WikipediaClient, pageAnalysisOptions PageAnalysisOptions) *PageAnalyzer {
	return &PageAnalyzer{
		client:                client,
		numberOfPageRevisions: utils.SetOrDefault(pageAnalysisOptions.NumberOfPageRevisions, 100),
		numberOfDaysHistory:   utils.SetOrDefault(pageAnalysisOptions.NumberOfDaysHistory, 30),
		numberOfContributors:  utils.SetOrDefault(pageAnalysisOptions.NumberOfContributors, 20),
	}
}

// GetPageProfile retrieves and analyzes a complete page profile
func (pa *PageAnalyzer) GetPageProfile(title string) (*models.PageProfile, error) {
	// 1. Get basic page information
	pageInfo, err := pa.client.GetPageInfo(title)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve page info: %w", err)
	}

	// 2. Get recent revisions (last 100)
	revisions, err := pa.client.GetPageRevisions(title, pa.numberOfPageRevisions)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve page revisions: %w", err)
	}

	// 3. Get detailed history for the last 30 days
	detailedHistory, err := pa.client.GetPageHistory(title, pa.numberOfDaysHistory)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve page history: %w", err)
	}

	// 4. Get contributors
	contributors, err := pa.client.GetPageContributors(title, pa.numberOfContributors)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve contributors: %w", err)
	}

	// 5. Create basic profile
	profile := &models.PageProfile{
		PageTitle:    pageInfo.Title,
		PageID:       pageInfo.PageID,
		Namespace:    pageInfo.NS,
		Language:     pa.client.Language(),
		LastModified: time.Now(), // Will be updated from revisions
		PageSize:     pageInfo.Length,
		RetrievedAt:  time.Now(),
	}

	// 6. Process revisions and calculate metrics
	profile.RecentRevisions = pa.convertRevisions(revisions)
	profile.TotalRevisions = len(revisions) // This would need a separate API call for exact count

	// 7. Analyze contributors
	profile.Contributors = pa.analyzeContributors(detailedHistory, contributors)

	// 8. Analyze conflicts and quality
	profile.ConflictStats = pa.analyzeConflicts(detailedHistory)
	profile.QualityMetrics = pa.analyzeQuality(detailedHistory, profile.Contributors)

	// 9. Calculate creation date from oldest revision
	if len(revisions) > 0 {
		oldestTimestamp, _ := time.Parse("2006-01-02T15:04:05Z", revisions[len(revisions)-1].Timestamp)
		profile.CreationDate = &oldestTimestamp

		// Update last modified from newest revision
		newestTimestamp, _ := time.Parse("2006-01-02T15:04:05Z", revisions[0].Timestamp)
		profile.LastModified = newestTimestamp
	}

	// 10. Calculate suspicion score
	profile.SuspicionScore, profile.SuspicionFlags = pa.calculateSuspicionScore(profile)

	return profile, nil
}

// convertRevisions converts API revisions to internal model
func (pa *PageAnalyzer) convertRevisions(wikiRevisions []models.WikiRevision) []models.Revision {
	revisions := make([]models.Revision, 0, len(wikiRevisions))

	var lastSize int
	for i, wr := range wikiRevisions {
		timestamp, _ := time.Parse("2006-01-02T15:04:05Z", wr.Timestamp)

		// Calculate size diff
		sizeDiff := 0
		if i < len(wikiRevisions)-1 {
			sizeDiff = wr.Size - lastSize
		}
		lastSize = wr.Size

		revision := models.Revision{
			RevID:       wr.RevID,
			ParentID:    wr.ParentID,
			Username:    wr.User,
			UserID:      wr.UserID,
			Timestamp:   timestamp,
			Comment:     wr.Comment,
			SizeDiff:    sizeDiff,
			NewSize:     wr.Size,
			IsMinor:     wr.Minor == "true",
			IsAnonymous: wr.Anon == "true",
			IsRevert:    pa.detectRevert(wr.Comment),
		}

		revisions = append(revisions, revision)
	}

	return revisions
}

// analyzeContributors analyzes page contributors and their patterns
func (pa *PageAnalyzer) analyzeContributors(revisions []models.WikiRevision, contributors []models.WikiContributor) []models.TopContributor {
	contributorStats := make(map[string]*models.TopContributor)

	// Process revisions to build contributor statistics
	for _, rev := range revisions {
		timestamp, _ := time.Parse("2006-01-02T15:04:05Z", rev.Timestamp)

		if existing, exists := contributorStats[rev.User]; exists {
			existing.EditCount++
			existing.TotalSizeDiff += rev.Size // This is approximate

			if timestamp.After(existing.LastEdit) {
				existing.LastEdit = timestamp
			}
			if timestamp.Before(existing.FirstEdit) {
				existing.FirstEdit = timestamp
			}
		} else {
			contributorStats[rev.User] = &models.TopContributor{
				Username:      rev.User,
				UserID:        rev.UserID,
				EditCount:     1,
				FirstEdit:     timestamp,
				LastEdit:      timestamp,
				TotalSizeDiff: rev.Size,
				IsAnonymous:   rev.Anon == "true",
				IsRegistered:  rev.UserID > 0,
			}
		}
	}

	// Convert to slice and sort by edit count
	var topContributors []models.TopContributor
	for _, contributor := range contributorStats {
		topContributors = append(topContributors, *contributor)
	}

	sort.Slice(topContributors, func(i, j int) bool {
		return topContributors[i].EditCount > topContributors[j].EditCount
	})

	// Limit to top 20
	if len(topContributors) > 20 {
		topContributors = topContributors[:20]
	}

	// Analyze each top contributor individually for suspicion scores
	pa.analyzeContributorSuspicion(topContributors)

	return topContributors
}

// analyzeContributorSuspicion analyzes each contributor individually for suspicion
func (pa *PageAnalyzer) analyzeContributorSuspicion(contributors []models.TopContributor) {
	// Create a user analyzer to analyze each contributor
	userAnalyzer := NewUserAnalyzer(pa.client)

	// Limit detailed analysis to top 10 contributors to avoid too many API calls
	limit := len(contributors)
	if limit > 10 {
		limit = 10
	}

	for i := 0; i < limit; i++ {
		contributor := &contributors[i]

		// Skip anonymous users as they can't be analyzed individually
		if contributor.IsAnonymous {
			contributor.SuspicionScore = 0
			contributor.SuspicionFlags = []string{"ANONYMOUS_USER"}
			continue
		}

		// Analyze the user profile
		userProfile, err := userAnalyzer.GetUserProfile(contributor.Username)
		if err != nil {
			contributor.SuspicionScore = -1
			contributor.AnalysisError = fmt.Sprintf("Analysis failed: %v", err)
			fmt.Printf("  ⚠️ Failed to analyze %s: %v\n", contributor.Username, err)
			continue
		}

		// Use the user's suspicion score and flags
		contributor.SuspicionScore = userProfile.SuspicionScore
		contributor.SuspicionFlags = userProfile.SuspicionFlags

		// Add page-specific flags based on contribution patterns
		pageSpecificFlags := pa.analyzeContributorPageBehavior(*contributor)
		contributor.SuspicionFlags = append(contributor.SuspicionFlags, pageSpecificFlags...)
	}

	// For contributors beyond the top 10, set basic suspicion indicators
	for i := limit; i < len(contributors); i++ {
		contributor := &contributors[i]

		if contributor.IsAnonymous {
			contributor.SuspicionScore = 5 // Low suspicion for anonymous
			contributor.SuspicionFlags = []string{"ANONYMOUS_USER"}
		} else {
			// Basic analysis without full API call
			contributor.SuspicionScore = pa.calculateBasicContributorSuspicion(*contributor)
			contributor.SuspicionFlags = pa.analyzeContributorPageBehavior(*contributor)
		}
	}
}

// analyzeContributorPageBehavior analyzes contributor behavior specific to this page
func (pa *PageAnalyzer) analyzeContributorPageBehavior(contributor models.TopContributor) []string {
	var flags []string

	// High edit concentration on this page
	if contributor.EditCount > 50 {
		flags = append(flags, "HIGH_PAGE_ACTIVITY")
	}

	// Recent account with high activity on this page
	daysSinceFirstEdit := int(time.Since(contributor.FirstEdit).Hours() / 24)
	if daysSinceFirstEdit < 7 && contributor.EditCount > 10 {
		flags = append(flags, "NEW_ACCOUNT_HIGH_PAGE_ACTIVITY")
	}

	// Very recent activity
	daysSinceLastEdit := int(time.Since(contributor.LastEdit).Hours() / 24)
	if daysSinceLastEdit < 1 {
		flags = append(flags, "VERY_RECENT_ACTIVITY")
	}

	// Large size changes (could indicate content manipulation)
	if contributor.TotalSizeDiff > 10000 || contributor.TotalSizeDiff < -5000 {
		flags = append(flags, "LARGE_CONTENT_CHANGES")
	}

	return flags
}

// calculateBasicContributorSuspicion calculates a basic suspicion score without full API analysis
func (pa *PageAnalyzer) calculateBasicContributorSuspicion(contributor models.TopContributor) int {
	score := 0

	// Recent account with high activity
	daysSinceFirstEdit := int(time.Since(contributor.FirstEdit).Hours() / 24)
	if daysSinceFirstEdit < 30 && contributor.EditCount > 20 {
		score += 15
	}

	// Very high activity on single page
	if contributor.EditCount > 100 {
		score += 10
	}

	// Large content changes
	if contributor.TotalSizeDiff > 15000 || contributor.TotalSizeDiff < -10000 {
		score += 10
	}

	// Very recent registration and activity
	if daysSinceFirstEdit < 7 && contributor.EditCount > 5 {
		score += 20
	}

	return score
}

// analyzeConflicts detects edit wars and conflicts
func (pa *PageAnalyzer) analyzeConflicts(revisions []models.WikiRevision) models.ConflictStats {
	stats := models.ConflictStats{
		ConflictingUsers: make([]string, 0),
		EditWarPeriods:   make([]models.EditWarPeriod, 0),
	}

	if len(revisions) == 0 {
		return stats
	}

	// Count reversions by looking for revert keywords in comments
	reversions := 0
	conflictUsers := make(map[string]bool)
	recentConflicts := 0
	sevenDaysAgo := time.Now().AddDate(0, 0, -7)

	for _, rev := range revisions {
		timestamp, _ := time.Parse("2006-01-02T15:04:05Z", rev.Timestamp)

		if pa.detectRevert(rev.Comment) {
			reversions++
			conflictUsers[rev.User] = true

			if timestamp.After(sevenDaysAgo) {
				recentConflicts++
			}
		}
	}

	stats.ReversionsCount = reversions
	stats.RecentConflicts = recentConflicts

	// Extract conflicting users
	for user := range conflictUsers {
		stats.ConflictingUsers = append(stats.ConflictingUsers, user)
	}

	// Calculate stability and controversy scores
	totalRevisions := len(revisions)
	if totalRevisions > 0 {
		stats.StabilityScore = 1.0 - (float64(reversions) / float64(totalRevisions))
		stats.ControversyScore = float64(reversions) / float64(totalRevisions)
	}

	// Detect edit war periods (simplified detection)
	stats.EditWarPeriods = pa.detectEditWarPeriods(revisions)

	return stats
}

// analyzeQuality calculates quality metrics for the page
func (pa *PageAnalyzer) analyzeQuality(revisions []models.WikiRevision, contributors []models.TopContributor) models.QualityMetrics {
	metrics := models.QualityMetrics{
		EditFrequency: models.EditFrequency{
			EditsByDay: make(map[string]int),
		},
	}

	if len(revisions) == 0 {
		return metrics
	}

	// Calculate various metrics
	totalSizeChanges := 0
	anonymousEdits := 0
	newEditorEdits := 0
	hourlyEdits := make(map[int]int)
	recentActivity := 0
	sevenDaysAgo := time.Now().AddDate(0, 0, -7)
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	ninetyDaysAgo := time.Now().AddDate(0, 0, -90)

	// Track new editors (registered within last 30 days of their first edit on this page)
	contributorFirstEdits := make(map[string]time.Time)

	for _, rev := range revisions {
		timestamp, _ := time.Parse("2006-01-02T15:04:05Z", rev.Timestamp)

		totalSizeChanges += rev.Size

		if rev.Anon == "true" {
			anonymousEdits++
		}

		// Track hourly activity
		hourlyEdits[timestamp.Hour()]++

		// Track daily activity
		dateKey := timestamp.Format("2006-01-02")
		metrics.EditFrequency.EditsByDay[dateKey]++

		// Count recent activity
		if timestamp.After(sevenDaysAgo) {
			recentActivity++
			metrics.EditFrequency.EditsLast7Days++
		}
		if timestamp.After(thirtyDaysAgo) {
			metrics.EditFrequency.EditsLast30Days++
		}
		if timestamp.After(ninetyDaysAgo) {
			metrics.EditFrequency.EditsLast90Days++
		}

		// Track first edits to detect new editors
		if firstEdit, exists := contributorFirstEdits[rev.User]; !exists || timestamp.Before(firstEdit) {
			contributorFirstEdits[rev.User] = timestamp
		}
	}

	// Detect new editors (users whose first edit on this page was within last 30 days)
	for user, firstEdit := range contributorFirstEdits {
		if time.Since(firstEdit) <= 30*24*time.Hour {
			// Count edits by this new editor
			for _, rev := range revisions {
				if rev.User == user {
					newEditorEdits++
				}
			}
		}
	}

	// Calculate metrics
	totalRevisions := len(revisions)
	if totalRevisions > 0 {
		metrics.AverageEditSize = float64(totalSizeChanges) / float64(totalRevisions)
		metrics.AnonymousEditRatio = float64(anonymousEdits) / float64(totalRevisions)
		metrics.NewEditorRatio = float64(newEditorEdits) / float64(totalRevisions)
	}

	// Detect recent activity burst (more than 10 edits in last 7 days)
	metrics.RecentActivityBurst = recentActivity > 10

	// Calculate contributor diversity (simplified Gini coefficient)
	if len(contributors) > 0 {
		metrics.ContributorDiversity = pa.calculateContributorDiversity(contributors)
	}

	// Find peak editing hours
	maxEdits := 0
	for hour, count := range hourlyEdits {
		if count > maxEdits {
			maxEdits = count
			metrics.EditFrequency.PeakEditingHours = []int{hour}
		} else if count == maxEdits {
			metrics.EditFrequency.PeakEditingHours = append(metrics.EditFrequency.PeakEditingHours, hour)
		}
	}

	return metrics
}

// calculateSuspicionScore calculates a suspicion score for the page
func (pa *PageAnalyzer) calculateSuspicionScore(profile *models.PageProfile) (int, []string) {
	score := 0
	flags := []string{}

	// 1. High conflict ratio
	if profile.ConflictStats.ControversyScore > 0.3 {
		score += 25
		flags = append(flags, "PAGE_HIGH_CONFLICT")
	}

	// 2. Few contributors for many edits
	if len(profile.Contributors) < 5 && profile.TotalRevisions > 100 {
		score += 20
		flags = append(flags, "PAGE_FEW_CONTRIBUTORS")
	}

	// 3. Recent intensive activity
	if profile.QualityMetrics.RecentActivityBurst {
		score += 15
		flags = append(flags, "PAGE_RECENT_INTENSIVE_ACTIVITY")
	}

	// 4. High anonymous editing ratio
	if profile.QualityMetrics.AnonymousEditRatio > 0.5 {
		score += 15
		flags = append(flags, "PAGE_ANONYMOUS_HEAVY_EDITING")
	}

	// 5. New editor dominance
	if len(profile.Contributors) > 0 {
		topContributor := profile.Contributors[0]
		daysSinceFirstEdit := int(time.Since(topContributor.FirstEdit).Hours() / 24)
		if daysSinceFirstEdit < 30 && float64(topContributor.EditCount)/float64(profile.TotalRevisions) > 0.5 {
			score += 20
			flags = append(flags, "PAGE_NEW_EDITOR_DOMINANCE")
		}
	}

	// 6. Low contributor diversity
	if profile.QualityMetrics.ContributorDiversity < 0.3 {
		score += 10
		flags = append(flags, "PAGE_LOW_DIVERSITY")
	}

	// 7. Recent conflicts
	if profile.ConflictStats.RecentConflicts > 5 {
		score += 15
		flags = append(flags, "PAGE_RECENT_CONFLICTS")
	}

	// Limit score to 100
	if score > 100 {
		score = 100
	}

	return score, flags
}

// Helper functions

// detectRevert checks if a comment indicates a revert
func (pa *PageAnalyzer) detectRevert(comment string) bool {
	comment = strings.ToLower(comment)
	revertKeywords := []string{
		"revert", "undo", "undid", "rv", "reverted",
		"restore", "restored", "rollback", "rolled back",
	}

	for _, keyword := range revertKeywords {
		if strings.Contains(comment, keyword) {
			return true
		}
	}
	return false
}

// detectEditWarPeriods identifies periods of intensive editing conflicts
func (pa *PageAnalyzer) detectEditWarPeriods(revisions []models.WikiRevision) []models.EditWarPeriod {
	var periods []models.EditWarPeriod

	// Simplified detection: look for periods with >5 revisions within 24 hours
	if len(revisions) < 5 {
		return periods
	}

	windowSize := 5
	for i := 0; i <= len(revisions)-windowSize; i++ {
		startTime, _ := time.Parse("2006-01-02T15:04:05Z", revisions[i].Timestamp)
		endTime, _ := time.Parse("2006-01-02T15:04:05Z", revisions[i+windowSize-1].Timestamp)

		// If 5+ revisions within 24 hours
		if endTime.Sub(startTime) <= 24*time.Hour {
			participants := make(map[string]bool)
			for j := i; j < i+windowSize; j++ {
				participants[revisions[j].User] = true
			}

			var participantList []string
			for user := range participants {
				participantList = append(participantList, user)
			}

			period := models.EditWarPeriod{
				StartTime:     startTime,
				EndTime:       endTime,
				Participants:  participantList,
				RevisionCount: windowSize,
			}

			periods = append(periods, period)
		}
	}

	return periods
}

// calculateContributorDiversity calculates a diversity score based on edit distribution
func (pa *PageAnalyzer) calculateContributorDiversity(contributors []models.TopContributor) float64 {
	if len(contributors) <= 1 {
		return 0.0
	}

	totalEdits := 0
	for _, contrib := range contributors {
		totalEdits += contrib.EditCount
	}

	if totalEdits == 0 {
		return 0.0
	}

	// Calculate Gini coefficient (simplified)
	var sumDiff float64
	for i, contrib1 := range contributors {
		for j, contrib2 := range contributors {
			if i != j {
				diff := float64(contrib1.EditCount - contrib2.EditCount)
				if diff < 0 {
					diff = -diff
				}
				sumDiff += diff
			}
		}
	}

	n := float64(len(contributors))
	meanEdits := float64(totalEdits) / n

	if meanEdits == 0 {
		return 0.0
	}

	gini := sumDiff / (2 * n * n * meanEdits)

	// Convert to diversity score (1 - gini, so higher = more diverse)
	return 1.0 - gini
}
