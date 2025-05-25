// internal/analyzer/user.go
package analyzer

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/intMeric/wikipedia-analyser/internal/client"

	"github.com/intMeric/wikipedia-analyser/internal/models"
)

// UserAnalyzer analyzes Wikipedia user data
type UserAnalyzer struct {
	client *client.WikipediaClient
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

	// 2. Get recent contributions
	contributions, err := ua.client.GetUserContributions(username, 100)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve contributions: %w", err)
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

	// 7. Calculate suspicion score
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

// calculateSuspicionScore calculates a basic suspicion score
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

	// Limit score to 100
	if score > 100 {
		score = 100
	}

	return score, flags
}
