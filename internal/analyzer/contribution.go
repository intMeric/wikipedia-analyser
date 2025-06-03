// internal/analyzer/contribution.go
package analyzer

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/intMeric/wikipedia-analyser/internal/client"
	"github.com/intMeric/wikipedia-analyser/internal/models"
	"github.com/intMeric/wikipedia-analyser/internal/utils"
)

// ContributionAnalyzer analyzes Wikipedia contributions/revisions
type ContributionAnalyzer struct {
	client        *client.WikipediaClient
	analysisDepth string
}

type ContributionAnalysisOptions struct {
	AnalysisDepth  string // "basic", "standard", "deep"
	IncludeContent bool
	IncludeContext bool
}

// NewContributionAnalyzer creates a new contribution analyzer
func NewContributionAnalyzer(client *client.WikipediaClient, options ContributionAnalysisOptions) *ContributionAnalyzer {
	depth := options.AnalysisDepth
	if depth == "" {
		depth = "standard"
	}

	return &ContributionAnalyzer{
		client:        client,
		analysisDepth: depth,
	}
}

// GetContributionProfile retrieves and analyzes a complete contribution profile
func (ca *ContributionAnalyzer) GetContributionProfile(revisionID int, pageTitle string) (*models.ContributionProfile, error) {
	// 1. Get page revisions to find our specific revision
	revisions, err := ca.client.GetPageRevisions(pageTitle, 500)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve page revisions: %w", err)
	}

	// Find the specific revision
	var targetRevision *models.WikiRevision
	for _, rev := range revisions {
		if rev.RevID == revisionID {
			targetRevision = &rev
			break
		}
	}

	if targetRevision == nil {
		return nil, fmt.Errorf("revision %d not found in page %s", revisionID, pageTitle)
	}

	// 2. Get page information
	pageInfo, err := ca.client.GetPageInfo(pageTitle)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve page info: %w", err)
	}

	// 3. Create basic profile
	profile := &models.ContributionProfile{
		RevisionID:  targetRevision.RevID,
		PageTitle:   pageInfo.Title,
		PageID:      pageInfo.PageID,
		Language:    ca.client.Language(),
		Comment:     targetRevision.Comment,
		Size:        targetRevision.Size,
		IsMinor:     targetRevision.Minor == "true",
		IsRevert:    ca.detectRevert(targetRevision.Comment),
		RetrievedAt: time.Now(),
	}

	// Parse timestamp
	timestamp, err := time.Parse("2006-01-02T15:04:05Z", targetRevision.Timestamp)
	if err == nil {
		profile.Timestamp = timestamp
	}

	// 4. Analyze author
	profile.Author, err = ca.analyzeAuthor(*targetRevision)
	if err != nil {
		return nil, fmt.Errorf("unable to analyze author: %w", err)
	}

	// 5. Get content analysis if requested
	if ca.analysisDepth == "standard" || ca.analysisDepth == "deep" {
		profile.ContentAnalysis = ca.analyzeContentFromRevision(*targetRevision, revisions)
	}

	// 6. Analyze context if deep analysis requested
	if ca.analysisDepth == "deep" {
		profile.ContextAnalysis = ca.analyzeContext(*targetRevision, *pageInfo, revisions)
	}

	// 7. Calculate quality metrics
	profile.QualityMetrics = ca.analyzeQuality(profile)

	// 8. Calculate suspicion score
	profile.SuspicionScore, profile.SuspicionFlags = ca.calculateSuspicionScore(profile)

	return profile, nil
}

// analyzeAuthor analyzes the author of the contribution
func (ca *ContributionAnalyzer) analyzeAuthor(revision models.WikiRevision) (models.ContributionAuthor, error) {
	author := models.ContributionAuthor{
		Username:     revision.User,
		UserID:       revision.UserID,
		IsAnonymous:  revision.Anon == "true",
		IsRegistered: revision.UserID > 0,
	}

	// Skip detailed analysis for anonymous users
	if author.IsAnonymous {
		return author, nil
	}

	// Get user information
	userInfo, err := ca.client.GetUserInfo(revision.User)
	if err != nil {
		return author, fmt.Errorf("unable to get user info: %w", err)
	}

	author.EditCount = userInfo.EditCount
	author.Groups = userInfo.Groups

	// Parse registration date if available
	if userInfo.Registration != "" {
		regDate, err := time.Parse("2006-01-02T15:04:05Z", userInfo.Registration)
		if err == nil {
			author.RegistrationDate = &regDate
		}
	}

	// Check if user is blocked
	author.IsBlocked = userInfo.BlockedBy != ""

	// Analyze recent activity
	author.RecentActivity = ca.analyzeRecentUserActivity(revision.User)

	// Calculate basic author suspicion score
	userAnalyzer := NewUserAnalyzer(ca.client)
	userProfile, err := userAnalyzer.GetUserProfile(revision.User)
	if err == nil {
		author.SuspicionScore = userProfile.SuspicionScore
	}

	return author, nil
}

// analyzeRecentUserActivity analyzes recent activity of the user
func (ca *ContributionAnalyzer) analyzeRecentUserActivity(username string) models.RecentUserActivity {
	activity := models.RecentUserActivity{}

	// Get user contributions for the last 30 days
	contributions, err := ca.client.GetUserContributions(username, 500)
	if err != nil {
		return activity
	}

	now := time.Now()
	last24h := now.AddDate(0, 0, -1)
	last7d := now.AddDate(0, 0, -7)
	last30d := now.AddDate(0, 0, -30)

	pagesEdited := make(map[string]bool)
	namespaces := make(map[int]bool)

	for _, contrib := range contributions {
		timestamp, err := time.Parse("2006-01-02T15:04:05Z", contrib.Timestamp)
		if err != nil {
			continue
		}

		if timestamp.After(last24h) {
			activity.EditsLast24h++
		}
		if timestamp.After(last7d) {
			activity.EditsLast7d++
		}
		if timestamp.After(last30d) {
			activity.EditsLast30d++
			pagesEdited[contrib.Title] = true
			namespaces[contrib.NS] = true
		}

		// Track last edit time
		if activity.LastEditTime == nil || timestamp.After(*activity.LastEditTime) {
			activity.LastEditTime = &timestamp
		}
	}

	activity.PagesEdited = len(pagesEdited)
	for ns := range namespaces {
		activity.Namespaces = append(activity.Namespaces, ns)
	}

	return activity
}

// analyzeContentFromRevision analyzes content changes from revision data
func (ca *ContributionAnalyzer) analyzeContentFromRevision(revision models.WikiRevision, allRevisions []models.WikiRevision) models.ContributionContent {
	content := models.ContributionContent{}

	// Find parent revision for comparison
	var parentRevision *models.WikiRevision
	for _, rev := range allRevisions {
		if rev.RevID == revision.ParentID {
			parentRevision = &rev
			break
		}
	}

	// Basic text analysis from size difference
	content.TextChanges = models.TextChangeAnalysis{
		CharsAdded:   utils.Max(0, revision.Size-getParentSize(parentRevision)),
		CharsRemoved: utils.Max(0, getParentSize(parentRevision)-revision.Size),
	}

	// Estimate words from character changes (rough approximation)
	content.TextChanges.WordsAdded = content.TextChanges.CharsAdded / 5
	content.TextChanges.WordsRemoved = content.TextChanges.CharsRemoved / 5

	// Analyze comment for content indicators
	content.TextChanges.IsStructural = ca.isStructuralEdit(revision.Comment)
	content.TextChanges.IsTrivial = ca.isTrivialEdit(revision.Comment) ||
		(content.TextChanges.CharsAdded < 50 && content.TextChanges.CharsRemoved < 50)

	// Basic language analysis
	content.LanguageAnalysis = models.LanguageAnalysis{
		Language:     ca.client.Language(),
		ToneAnalysis: "neutral", // Default
	}

	// Analyze comment for bias indicators
	content.LanguageAnalysis.POVWords = ca.findPOVWords(revision.Comment)
	content.LanguageAnalysis.BiasScore = float64(len(content.LanguageAnalysis.POVWords)) / 10.0

	if len(content.LanguageAnalysis.POVWords) > 0 {
		content.LanguageAnalysis.ToneAnalysis = "potentially_biased"
	}

	// Determine content type
	content.ContentType = ca.determineContentType(revision.Comment, content.TextChanges)

	return content
}

// analyzeContext analyzes the context of the contribution
func (ca *ContributionAnalyzer) analyzeContext(revision models.WikiRevision, pageInfo models.WikiPageInfo, allRevisions []models.WikiRevision) models.ContributionContext {
	context := models.ContributionContext{}

	// Analyze page context
	context.PageContext = ca.analyzePageContext(pageInfo)

	// Analyze timing context
	timestamp, _ := time.Parse("2006-01-02T15:04:05Z", revision.Timestamp)
	context.TimingContext = ca.analyzeTimingContext(timestamp, allRevisions, revision.RevID)

	// Analyze author context
	context.AuthorContext = ca.analyzeAuthorContext(revision.User)

	// Find related edits
	context.RelatedEdits = ca.findRelatedEdits(revision, allRevisions)

	// Analyze conflict context
	context.ConflictContext = ca.analyzeConflictContext(revision, allRevisions)

	return context
}

// analyzePageContext analyzes the context of the page
func (ca *ContributionAnalyzer) analyzePageContext(pageInfo models.WikiPageInfo) models.PageContextInfo {
	context := models.PageContextInfo{}

	// Estimate page age (simplified)
	context.PageAge = 365 // Default 1 year

	// Basic controversiality estimation from title
	context.Controversiality = ca.estimateControversiality(pageInfo.Title, []string{})

	return context
}

// analyzeTimingContext analyzes timing patterns
func (ca *ContributionAnalyzer) analyzeTimingContext(timestamp time.Time, allRevisions []models.WikiRevision, currentRevID int) models.TimingContextInfo {
	context := models.TimingContextInfo{
		EditHour:      timestamp.Hour(),
		EditDayOfWeek: int(timestamp.Weekday()),
		IsWeekend:     timestamp.Weekday() == time.Saturday || timestamp.Weekday() == time.Sunday,
	}

	// Find time since last edit
	for i, rev := range allRevisions {
		if rev.RevID == currentRevID && i < len(allRevisions)-1 {
			prevRev := allRevisions[i+1]
			prevTime, _ := time.Parse("2006-01-02T15:04:05Z", prevRev.Timestamp)
			context.TimeSinceLastEdit = int(timestamp.Sub(prevTime).Minutes())
			break
		}
	}

	return context
}

// analyzeAuthorContext analyzes author context
func (ca *ContributionAnalyzer) analyzeAuthorContext(username string) models.AuthorContextInfo {
	context := models.AuthorContextInfo{}

	// Get user contributions to analyze patterns
	contributions, err := ca.client.GetUserContributions(username, 100)
	if err != nil {
		return context
	}

	// Analyze edit frequency
	context.EditFrequency = ca.analyzeEditFrequency(contributions)

	// Analyze edit patterns
	context.EditPattern = ca.analyzeEditPattern(contributions)

	// Analyze page focus
	context.PageFocus = ca.analyzePageFocus(contributions)

	return context
}

// analyzeEditFrequency analyzes editing frequency - FIXED
func (ca *ContributionAnalyzer) analyzeEditFrequency(contributions []models.WikiContribution) models.EditFrequencyInfo {
	info := models.EditFrequencyInfo{}

	if len(contributions) == 0 {
		return info
	}

	// Calculate frequency based on time span of contributions
	if len(contributions) > 1 {
		firstTime, _ := time.Parse("2006-01-02T15:04:05Z", contributions[len(contributions)-1].Timestamp)
		lastTime, _ := time.Parse("2006-01-02T15:04:05Z", contributions[0].Timestamp)

		daysDiff := lastTime.Sub(firstTime).Hours() / 24
		if daysDiff > 0 {
			info.EditsPerDay = float64(len(contributions)) / daysDiff
			info.EditsPerHour = info.EditsPerDay / 24.0
		}
	}

	return info
}

// analyzeEditPattern analyzes editing patterns
func (ca *ContributionAnalyzer) analyzeEditPattern(contributions []models.WikiContribution) models.EditPatternInfo {
	info := models.EditPatternInfo{}

	namespaces := make(map[int]int)
	for _, contrib := range contributions {
		namespaces[contrib.NS]++
	}

	// Find favorite namespaces
	maxCount := 0
	for ns, count := range namespaces {
		if count > maxCount {
			maxCount = count
			info.FavoriteNamespaces = []int{ns}
		} else if count == maxCount {
			info.FavoriteNamespaces = append(info.FavoriteNamespaces, ns)
		}
	}

	return info
}

// analyzePageFocus analyzes page focus patterns
func (ca *ContributionAnalyzer) analyzePageFocus(contributions []models.WikiContribution) models.PageFocusInfo {
	info := models.PageFocusInfo{}

	pages := make(map[string]int)
	for _, contrib := range contributions {
		pages[contrib.Title]++
	}

	info.PagesEditedCount = len(pages)

	// Calculate top page edit ratio
	if len(pages) > 0 && len(contributions) > 0 {
		maxEdits := 0
		for _, count := range pages {
			if count > maxEdits {
				maxEdits = count
			}
		}
		info.TopPageEditRatio = float64(maxEdits) / float64(len(contributions))
	}

	// Determine if specialized editor (>50% edits on single page)
	info.IsSpecializedEditor = info.TopPageEditRatio > 0.5

	return info
}

// findRelatedEdits finds edits related to this contribution
func (ca *ContributionAnalyzer) findRelatedEdits(revision models.WikiRevision, allRevisions []models.WikiRevision) []models.RelatedEdit {
	var relatedEdits []models.RelatedEdit

	revisionTime, _ := time.Parse("2006-01-02T15:04:05Z", revision.Timestamp)

	for _, rev := range allRevisions {
		if rev.RevID == revision.RevID {
			continue
		}

		revTime, _ := time.Parse("2006-01-02T15:04:05Z", rev.Timestamp)
		timeDiff := revTime.Sub(revisionTime)

		// Consider edits within 24 hours as potentially related
		if math.Abs(float64(timeDiff)) <= float64(24*time.Hour) {
			related := models.RelatedEdit{
				RevisionID: rev.RevID,
				Author:     rev.User,
				Timestamp:  revTime,
				Relation:   ca.determineRelation(revision, rev),
				Similarity: ca.calculateSimilarity(revision, rev),
			}
			relatedEdits = append(relatedEdits, related)
		}
	}

	return relatedEdits
}

// analyzeConflictContext analyzes conflict-related context
func (ca *ContributionAnalyzer) analyzeConflictContext(revision models.WikiRevision, allRevisions []models.WikiRevision) models.ConflictContextInfo {
	context := models.ConflictContextInfo{}

	// Check if this edit is a revert
	if ca.detectRevert(revision.Comment) {
		context.IsContested = true
		context.ConflictSeverity = 0.7
	}

	// Check for rapid back-and-forth edits
	revisionTime, _ := time.Parse("2006-01-02T15:04:05Z", revision.Timestamp)
	recentReverts := 0

	for _, rev := range allRevisions {
		if rev.RevID == revision.RevID {
			continue
		}

		revTime, _ := time.Parse("2006-01-02T15:04:05Z", rev.Timestamp)
		if revTime.After(revisionTime.Add(-24*time.Hour)) && revTime.Before(revisionTime.Add(24*time.Hour)) {
			if ca.detectRevert(rev.Comment) {
				recentReverts++
			}
		}
	}

	if recentReverts > 2 {
		context.IsContested = true
		context.ConflictSeverity = utils.MinFloat64(1.0, float64(recentReverts)/5.0)
	}

	return context
}

// analyzeQuality analyzes the quality of the contribution
func (ca *ContributionAnalyzer) analyzeQuality(profile *models.ContributionProfile) models.ContributionQuality {
	quality := models.ContributionQuality{}

	// Analyze content quality
	quality.ContentQuality = ca.analyzeContentQuality(profile)

	// Analyze source quality
	quality.SourceQuality = ca.analyzeSourceQuality(profile)

	// Analyze structure quality
	quality.StructureQuality = ca.analyzeStructureQuality(profile)

	// Analyze compliance
	quality.ComplianceScore = ca.analyzeCompliance(profile)

	// Calculate overall quality
	quality.OverallQuality = (quality.ContentQuality.Accuracy*0.3 +
		quality.SourceQuality.ReliabilityScore*0.25 +
		quality.StructureQuality.Formatting*0.2 +
		quality.ComplianceScore.PolicyCompliance*0.25)

	return quality
}

// analyzeContentQuality analyzes content quality metrics
func (ca *ContributionAnalyzer) analyzeContentQuality(profile *models.ContributionProfile) models.ContentQualityInfo {
	quality := models.ContentQualityInfo{
		Accuracy:     0.8, // Default values
		Completeness: 0.7,
		Neutrality:   0.9,
		Clarity:      0.8,
		Relevance:    0.9,
	}

	// Adjust based on language analysis
	if profile.ContentAnalysis.LanguageAnalysis.BiasScore > 0.3 {
		quality.Neutrality -= profile.ContentAnalysis.LanguageAnalysis.BiasScore
	}

	// Adjust based on edit size
	if profile.ContentAnalysis.TextChanges.IsTrivial {
		quality.Completeness -= 0.2
	}

	return quality
}

// analyzeSourceQuality analyzes source quality
func (ca *ContributionAnalyzer) analyzeSourceQuality(profile *models.ContributionProfile) models.SourceQualityInfo {
	quality := models.SourceQualityInfo{
		ReliabilityScore: 0.8, // Default values
		DiversityScore:   0.7,
		RecencyScore:     0.8,
		AuthorityScore:   0.7,
	}

	// Basic analysis from comment content
	comment := strings.ToLower(profile.Comment)
	if strings.Contains(comment, "source") || strings.Contains(comment, "ref") || strings.Contains(comment, "cite") {
		quality.ReliabilityScore += 0.1
	}

	return quality
}

// analyzeStructureQuality analyzes structural quality
func (ca *ContributionAnalyzer) analyzeStructureQuality(profile *models.ContributionProfile) models.StructureQualityInfo {
	quality := models.StructureQualityInfo{
		Formatting:      0.8, // Default values
		Organization:    0.8,
		WikimarkupScore: 0.9,
		LinkingQuality:  0.8,
		CategoryUsage:   0.7,
		TemplateUsage:   0.8,
	}

	// Adjust based on content analysis
	if profile.ContentAnalysis.TextChanges.IsStructural {
		quality.Organization += 0.1
		quality.WikimarkupScore += 0.05
	}

	return quality
}

// analyzeCompliance analyzes policy compliance
func (ca *ContributionAnalyzer) analyzeCompliance(profile *models.ContributionProfile) models.ComplianceInfo {
	compliance := models.ComplianceInfo{
		PolicyCompliance:    0.9, // Default values
		GuidelineCompliance: 0.8,
		COI_Risk:            0.1,
		AdvertisingRisk:     0.1,
		VandalismRisk:       0.1,
	}

	// Check for potential issues
	if profile.ContentAnalysis.LanguageAnalysis.BiasScore > 0.5 {
		compliance.PolicyCompliance -= 0.2
		compliance.ViolatedPolicies = append(compliance.ViolatedPolicies, "NPOV")
	}

	// Check for vandalism indicators
	if profile.IsRevert || strings.Contains(strings.ToLower(profile.Comment), "vandal") {
		compliance.VandalismRisk += 0.3
	}

	return compliance
}

// calculateSuspicionScore calculates suspicion score and flags
func (ca *ContributionAnalyzer) calculateSuspicionScore(profile *models.ContributionProfile) (int, []string) {
	score := 0
	flags := []string{}

	// Check author suspicion
	if profile.Author.SuspicionScore > 0 {
		score += profile.Author.SuspicionScore / 2 // Dilute author score
	}

	// Check for reverts
	if profile.IsRevert {
		score += 15
		flags = append(flags, "REVERT_EDIT")
	}

	// Check for rapid editing
	if profile.Author.RecentActivity.EditsLast24h > 50 {
		score += 20
		flags = append(flags, "RAPID_EDITING")
	}

	// Check for anonymous editing
	if profile.Author.IsAnonymous {
		score += 5
		flags = append(flags, "ANONYMOUS_EDIT")
	}

	// Check for new account
	if profile.Author.RegistrationDate != nil {
		daysSinceReg := int(time.Since(*profile.Author.RegistrationDate).Hours() / 24)
		if daysSinceReg < 7 {
			score += 15
			flags = append(flags, "NEW_ACCOUNT")
		}
	}

	// Check for bias indicators
	if profile.ContentAnalysis.LanguageAnalysis.BiasScore > 0.3 {
		score += 10
		flags = append(flags, "POTENTIAL_BIAS")
	}

	// Check for large content changes
	if profile.ContentAnalysis.TextChanges.CharsAdded > 5000 {
		score += 10
		flags = append(flags, "LARGE_ADDITION")
	}
	if profile.ContentAnalysis.TextChanges.CharsRemoved > 2000 {
		score += 15
		flags = append(flags, "LARGE_REMOVAL")
	}

	// Check for blocked user
	if profile.Author.IsBlocked {
		score += 25
		flags = append(flags, "BLOCKED_USER")
	}

	// Limit score to 100
	if score > 100 {
		score = 100
	}

	return score, flags
}

// Helper functions

// getParentSize returns the size of parent revision or 0 if not found
func getParentSize(parentRevision *models.WikiRevision) int {
	if parentRevision == nil {
		return 0
	}
	return parentRevision.Size
}

// detectRevert checks if a comment indicates a revert
func (ca *ContributionAnalyzer) detectRevert(comment string) bool {
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

// isStructuralEdit checks if edit comment indicates structural changes
func (ca *ContributionAnalyzer) isStructuralEdit(comment string) bool {
	comment = strings.ToLower(comment)
	structuralKeywords := []string{
		"section", "heading", "template", "infobox", "category",
		"reorganiz", "restructur", "format", "layout",
	}

	for _, keyword := range structuralKeywords {
		if strings.Contains(comment, keyword) {
			return true
		}
	}
	return false
}

// isTrivialEdit checks if edit comment indicates trivial changes
func (ca *ContributionAnalyzer) isTrivialEdit(comment string) bool {
	comment = strings.ToLower(comment)
	trivialKeywords := []string{
		"typo", "spelling", "grammar", "punctuation", "format",
		"minor", "fix", "correct",
	}

	for _, keyword := range trivialKeywords {
		if strings.Contains(comment, keyword) {
			return true
		}
	}
	return false
}

// findPOVWords finds potential POV words in text
func (ca *ContributionAnalyzer) findPOVWords(text string) []string {
	var povWords []string

	// Common POV indicators
	povIndicators := []string{
		"obviously", "clearly", "undoubtedly", "best", "worst",
		"always", "never", "perfect", "terrible", "amazing",
	}

	textLower := strings.ToLower(text)
	for _, word := range povIndicators {
		if strings.Contains(textLower, word) {
			povWords = append(povWords, word)
		}
	}

	return povWords
}

// determineContentType determines the type of content change
func (ca *ContributionAnalyzer) determineContentType(comment string, changes models.TextChangeAnalysis) string {
	comment = strings.ToLower(comment)

	if strings.Contains(comment, "typo") || strings.Contains(comment, "spelling") {
		return "typo_fix"
	}
	if strings.Contains(comment, "source") || strings.Contains(comment, "reference") {
		return "source_addition"
	}
	if changes.IsStructural {
		return "structural_change"
	}
	if changes.IsTrivial {
		return "minor_edit"
	}

	return "content_edit"
}

// determineRelation determines the relationship between two revisions
func (ca *ContributionAnalyzer) determineRelation(rev1, rev2 models.WikiRevision) string {
	if ca.detectRevert(rev2.Comment) && strings.Contains(rev2.Comment, fmt.Sprintf("%d", rev1.RevID)) {
		return "revert"
	}
	if rev1.User == rev2.User {
		return "follow-up"
	}
	return "related"
}

// calculateSimilarity calculates similarity between two revisions
func (ca *ContributionAnalyzer) calculateSimilarity(rev1, rev2 models.WikiRevision) float64 {
	// Simple similarity based on comment similarity
	if rev1.Comment == rev2.Comment {
		return 1.0
	}
	if strings.Contains(rev1.Comment, rev2.Comment) || strings.Contains(rev2.Comment, rev1.Comment) {
		return 0.7
	}
	if rev1.User == rev2.User {
		return 0.5
	}
	return 0.1
}

// estimateControversiality estimates page controversiality
func (ca *ContributionAnalyzer) estimateControversiality(title string, categories []string) float64 {
	// Simple heuristic based on title and categories
	controversialKeywords := []string{"politics", "religion", "war", "conflict", "controversy"}

	titleLower := strings.ToLower(title)
	for _, keyword := range controversialKeywords {
		if strings.Contains(titleLower, keyword) {
			return 0.8
		}
	}

	for _, category := range categories {
		categoryLower := strings.ToLower(category)
		for _, keyword := range controversialKeywords {
			if strings.Contains(categoryLower, keyword) {
				return 0.6
			}
		}
	}

	return 0.2 // Default low controversiality
}
