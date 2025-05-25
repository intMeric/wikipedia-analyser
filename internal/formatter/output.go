// internal/formatter/output.go
package formatter

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
	"wikianalyser/internal/models"

	"github.com/fatih/color"
	"gopkg.in/yaml.v2"
)

var (
	// Colors for terminal display
	headerColor    = color.New(color.FgCyan, color.Bold)
	successColor   = color.New(color.FgGreen)
	warningColor   = color.New(color.FgYellow)
	dangerColor    = color.New(color.FgRed, color.Bold)
	infoColor      = color.New(color.FgBlue)
	secondaryColor = color.New(color.FgHiBlack)
)

// FormatUserProfile formats the user profile according to the specified format
func FormatUserProfile(profile *models.UserProfile, format string) (string, error) {
	switch strings.ToLower(format) {
	case "json":
		return formatAsJSON(profile)
	case "yaml", "yml":
		return formatAsYAML(profile)
	case "table", "":
		return formatAsTable(profile), nil
	default:
		return "", fmt.Errorf("unsupported format: %s (supported: table, json, yaml)", format)
	}
}

// formatAsJSON formats as JSON
func formatAsJSON(profile *models.UserProfile) (string, error) {
	data, err := json.MarshalIndent(profile, "", "  ")
	if err != nil {
		return "", fmt.Errorf("JSON formatting error: %w", err)
	}
	return string(data), nil
}

// formatAsYAML formats as YAML
func formatAsYAML(profile *models.UserProfile) (string, error) {
	data, err := yaml.Marshal(profile)
	if err != nil {
		return "", fmt.Errorf("YAML formatting error: %w", err)
	}
	return string(data), nil
}

// formatAsTable formats as readable table
func formatAsTable(profile *models.UserProfile) string {
	var output strings.Builder

	// Header with username and suspicion score
	output.WriteString(headerColor.Sprint("╭─────────────────────────────────────────────────────────────╮\n"))
	output.WriteString(headerColor.Sprintf("│  📊 WIKIPEDIA USER PROFILE: %-27s         │\n", profile.Username))
	output.WriteString(headerColor.Sprint("╰─────────────────────────────────────────────────────────────╯\n\n"))

	// Suspicion score with color
	suspicionText := getSuspicionText(profile.SuspicionScore)
	suspicionColor := getSuspicionColor(profile.SuspicionScore)
	output.WriteString(fmt.Sprintf("🚨 %s %s (%d/100)\n\n",
		suspicionColor.Sprint("Suspicion Score:"),
		suspicionColor.Sprint(suspicionText),
		profile.SuspicionScore))

	// Basic information
	output.WriteString(headerColor.Sprint("📋 BASIC INFORMATION\n"))
	output.WriteString(strings.Repeat("─", 50) + "\n")

	// Basic information - using simple formatting instead of complex table
	output.WriteString("👤 Username:           " + profile.Username + "\n")
	output.WriteString("🆔 User ID:            " + strconv.Itoa(profile.UserID) + "\n")
	output.WriteString("✏️ Edit Count:         " + strconv.Itoa(profile.EditCount) + "\n")

	if profile.RegistrationDate != nil {
		regDate := profile.RegistrationDate.Format("02/01/2006")
		daysSince := int(time.Since(*profile.RegistrationDate).Hours() / 24)
		output.WriteString(fmt.Sprintf("📅 Registration Date:  %s (%d days ago)\n", regDate, daysSince))
	}

	output.WriteString("🌍 Wikipedia Language: " + profile.Language + "\n")
	output.WriteString("🔍 Analysis Performed: " + profile.RetrievedAt.Format("02/01/2006 15:04:05") + "\n")
	output.WriteString("\n")

	// Groups and rights
	if len(profile.Groups) > 0 || len(profile.ImplicitGroups) > 0 {
		output.WriteString(headerColor.Sprint("👥 GROUPS AND RIGHTS\n"))
		output.WriteString(strings.Repeat("─", 50) + "\n")

		if len(profile.Groups) > 0 {
			output.WriteString(fmt.Sprintf("🏷️  Explicit Groups: %s\n",
				infoColor.Sprint(strings.Join(profile.Groups, ", "))))
		}
		if len(profile.ImplicitGroups) > 0 {
			output.WriteString(fmt.Sprintf("🔒 Implicit Groups: %s\n",
				secondaryColor.Sprint(strings.Join(profile.ImplicitGroups, ", "))))
		}
		output.WriteString("\n")
	}

	// Block information
	if profile.BlockInfo != nil && profile.BlockInfo.Blocked {
		output.WriteString(dangerColor.Sprint("🚫 USER BLOCKED\n"))
		output.WriteString(strings.Repeat("─", 50) + "\n")
		output.WriteString(fmt.Sprintf("👮 Blocked by: %s\n", profile.BlockInfo.BlockedBy))
		output.WriteString(fmt.Sprintf("📝 Reason: %s\n", profile.BlockInfo.Reason))
		if !profile.BlockInfo.BlockEnd.IsZero() {
			output.WriteString(fmt.Sprintf("⏰ Block expires: %s\n",
				profile.BlockInfo.BlockEnd.Format("02/01/2006 15:04:05")))
		}
		output.WriteString("\n")
	}

	// Suspicion flags
	if len(profile.SuspicionFlags) > 0 {
		output.WriteString(warningColor.Sprint("⚠️  SUSPICION INDICATORS\n"))
		output.WriteString(strings.Repeat("─", 50) + "\n")
		for _, flag := range profile.SuspicionFlags {
			flagText := formatSuspicionFlag(flag)
			output.WriteString(fmt.Sprintf("🔸 %s\n", warningColor.Sprint(flagText)))
		}
		output.WriteString("\n")
	}

	// Activity statistics - using simple formatting
	output.WriteString(headerColor.Sprint("📈 ACTIVITY STATISTICS\n"))
	output.WriteString(strings.Repeat("─", 50) + "\n")

	if profile.ActivityStats.DaysActive > 0 {
		output.WriteString("📅 Days Active:        " + strconv.Itoa(profile.ActivityStats.DaysActive) + "\n")
		output.WriteString(fmt.Sprintf("📊 Edits/day (average): %.2f\n", profile.ActivityStats.AverageEditsPerDay))
	}
	output.WriteString(fmt.Sprintf("🕐 Most Active Hour:   %02d:00\n", profile.ActivityStats.MostActiveHour))
	output.WriteString("📆 Most Active Day:    " + profile.ActivityStats.MostActiveDay + "\n")
	output.WriteString("\n")

	// Namespace distribution - using simple formatting
	if len(profile.ActivityStats.NamespaceDistrib) > 0 {
		output.WriteString(headerColor.Sprint("📂 NAMESPACE DISTRIBUTION\n"))
		output.WriteString(strings.Repeat("─", 50) + "\n")

		totalEdits := 0
		for _, count := range profile.ActivityStats.NamespaceDistrib {
			totalEdits += count
		}

		for ns, count := range profile.ActivityStats.NamespaceDistrib {
			percentage := float64(count) / float64(totalEdits) * 100
			output.WriteString(fmt.Sprintf("%-15s %5d edits (%.1f%%)\n", ns, count, percentage))
		}
		output.WriteString("\n")
	}

	// Most edited pages - using simple formatting
	if len(profile.TopPages) > 0 {
		output.WriteString(headerColor.Sprint("📄 MOST EDITED PAGES\n"))
		output.WriteString(strings.Repeat("─", 80) + "\n")

		for i, page := range profile.TopPages {
			if i >= 5 { // Limit to 5 pages
				break
			}

			title := page.PageTitle
			if len(title) > 50 {
				title = title[:50] + "..."
			}

			output.WriteString(fmt.Sprintf("%-55s %3d edits %+5d diff %s\n",
				title,
				page.EditCount,
				page.TotalSizeDiff,
				page.LastEdit.Format("02/01/06"),
			))
		}
		output.WriteString("\n")
	}

	// Recent contributions (preview) - using simple formatting
	if len(profile.RecentContribs) > 0 {
		output.WriteString(headerColor.Sprint("🕒 RECENT CONTRIBUTIONS (last 5)\n"))
		output.WriteString(strings.Repeat("─", 80) + "\n")

		for i, contrib := range profile.RecentContribs {
			if i >= 5 {
				break
			}

			title := contrib.PageTitle
			if len(title) > 35 {
				title = title[:35] + "..."
			}

			comment := contrib.Comment
			if len(comment) > 30 {
				comment = comment[:30] + "..."
			}
			if comment == "" {
				comment = secondaryColor.Sprint("(no comment)")
			}

			diffStr := fmt.Sprintf("%+d", contrib.SizeDiff)
			if contrib.SizeDiff > 0 {
				diffStr = successColor.Sprint(diffStr)
			} else if contrib.SizeDiff < 0 {
				diffStr = warningColor.Sprint(diffStr)
			}

			output.WriteString(fmt.Sprintf("%-12s %-38s %s %s\n",
				contrib.Timestamp.Format("02/01 15:04"),
				title,
				diffStr,
				comment,
			))
		}
		output.WriteString("\n")
	}

	// Footer
	output.WriteString(secondaryColor.Sprint("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n"))
	output.WriteString(secondaryColor.Sprintf("📊 WikiOSINT Analysis - %d contributions analyzed on %s.wikipedia.org\n",
		len(profile.RecentContribs), profile.Language))

	return output.String()
}

// FormatPageProfile formats the page profile according to the specified format
func FormatPageProfile(profile *models.PageProfile, format string) (string, error) {
	switch strings.ToLower(format) {
	case "json":
		return formatPageAsJSON(profile)
	case "yaml", "yml":
		return formatPageAsYAML(profile)
	case "table", "":
		return formatPageAsTable(profile), nil
	default:
		return "", fmt.Errorf("unsupported format: %s (supported: table, json, yaml)", format)
	}
}

// FormatPageHistory formats page history analysis
func FormatPageHistory(profile *models.PageProfile, format string) (string, error) {
	switch strings.ToLower(format) {
	case "json":
		return formatPageAsJSON(profile)
	case "yaml", "yml":
		return formatPageAsYAML(profile)
	case "table", "":
		return formatPageHistoryAsTable(profile), nil
	default:
		return "", fmt.Errorf("unsupported format: %s (supported: table, json, yaml)", format)
	}
}

// FormatPageConflicts formats page conflict analysis
func FormatPageConflicts(profile *models.PageProfile, format string) (string, error) {
	switch strings.ToLower(format) {
	case "json":
		return formatPageAsJSON(profile)
	case "yaml", "yml":
		return formatPageAsYAML(profile)
	case "table", "":
		return formatPageConflictsAsTable(profile), nil
	default:
		return "", fmt.Errorf("unsupported format: %s (supported: table, json, yaml)", format)
	}
}

// formatPageAsJSON formats page profile as JSON
func formatPageAsJSON(profile *models.PageProfile) (string, error) {
	data, err := json.MarshalIndent(profile, "", "  ")
	if err != nil {
		return "", fmt.Errorf("JSON formatting error: %w", err)
	}
	return string(data), nil
}

// formatPageAsYAML formats page profile as YAML
func formatPageAsYAML(profile *models.PageProfile) (string, error) {
	data, err := yaml.Marshal(profile)
	if err != nil {
		return "", fmt.Errorf("YAML formatting error: %w", err)
	}
	return string(data), nil
}

// formatPageAsTable formats page profile as readable table
func formatPageAsTable(profile *models.PageProfile) string {
	var output strings.Builder

	// Header with page title and suspicion score
	output.WriteString(headerColor.Sprint("╭─────────────────────────────────────────────────────────────╮\n"))
	output.WriteString(headerColor.Sprintf("│  📄 WIKIPEDIA PAGE ANALYSIS: %-27s       │\n", truncateString(profile.PageTitle, 27)))
	output.WriteString(headerColor.Sprint("╰─────────────────────────────────────────────────────────────╯\n\n"))

	// Suspicion score with color
	suspicionText := getSuspicionText(profile.SuspicionScore)
	suspicionColor := getSuspicionColor(profile.SuspicionScore)
	output.WriteString(fmt.Sprintf("🚨 %s %s (%d/100)\n\n",
		suspicionColor.Sprint("Suspicion Score:"),
		suspicionColor.Sprint(suspicionText),
		profile.SuspicionScore))

	// Basic information
	output.WriteString(headerColor.Sprint("📋 PAGE INFORMATION\n"))
	output.WriteString(strings.Repeat("─", 50) + "\n")

	output.WriteString("📄 Page Title:         " + profile.PageTitle + "\n")
	output.WriteString("🆔 Page ID:            " + strconv.Itoa(profile.PageID) + "\n")
	output.WriteString("📊 Total Revisions:    " + strconv.Itoa(profile.TotalRevisions) + "\n")
	output.WriteString("📏 Current Size:       " + strconv.Itoa(profile.PageSize) + " bytes\n")

	if profile.CreationDate != nil {
		creationDate := profile.CreationDate.Format("02/01/2006")
		daysSince := int(time.Since(*profile.CreationDate).Hours() / 24)
		output.WriteString(fmt.Sprintf("📅 Created:            %s (%d days ago)\n", creationDate, daysSince))
	}

	output.WriteString("🔄 Last Modified:      " + profile.LastModified.Format("02/01/2006 15:04") + "\n")
	output.WriteString("🌍 Wikipedia Language: " + profile.Language + "\n")
	output.WriteString("🔍 Analysis Performed: " + profile.RetrievedAt.Format("02/01/2006 15:04:05") + "\n")
	output.WriteString("\n")

	// Suspicion flags
	if len(profile.SuspicionFlags) > 0 {
		output.WriteString(warningColor.Sprint("⚠️  SUSPICION INDICATORS\n"))
		output.WriteString(strings.Repeat("─", 50) + "\n")
		for _, flag := range profile.SuspicionFlags {
			flagText := formatPageSuspicionFlag(flag)
			output.WriteString(fmt.Sprintf("🔸 %s\n", warningColor.Sprint(flagText)))
		}
		output.WriteString("\n")
	}

	// Conflict statistics
	output.WriteString(headerColor.Sprint("⚔️ CONFLICT ANALYSIS\n"))
	output.WriteString(strings.Repeat("─", 50) + "\n")

	output.WriteString("🔄 Total Reversions:   " + strconv.Itoa(profile.ConflictStats.ReversionsCount) + "\n")
	output.WriteString("📅 Recent Conflicts:   " + strconv.Itoa(profile.ConflictStats.RecentConflicts) + " (last 7 days)\n")
	output.WriteString(fmt.Sprintf("📈 Stability Score:    %.2f/1.00\n", profile.ConflictStats.StabilityScore))
	output.WriteString(fmt.Sprintf("⚡ Controversy Score:  %.2f\n", profile.ConflictStats.ControversyScore))

	if len(profile.ConflictStats.ConflictingUsers) > 0 {
		output.WriteString("👥 Conflicting Users:  " + strings.Join(profile.ConflictStats.ConflictingUsers[:min(5, len(profile.ConflictStats.ConflictingUsers))], ", "))
		if len(profile.ConflictStats.ConflictingUsers) > 5 {
			output.WriteString(fmt.Sprintf(" (+%d more)", len(profile.ConflictStats.ConflictingUsers)-5))
		}
		output.WriteString("\n")
	}
	output.WriteString("\n")

	// Quality metrics
	output.WriteString(headerColor.Sprint("📊 QUALITY METRICS\n"))
	output.WriteString(strings.Repeat("─", 50) + "\n")

	output.WriteString(fmt.Sprintf("📝 Average Edit Size:  %.1f bytes\n", profile.QualityMetrics.AverageEditSize))
	output.WriteString(fmt.Sprintf("👤 Anonymous Ratio:    %.1f%%\n", profile.QualityMetrics.AnonymousEditRatio*100))
	output.WriteString(fmt.Sprintf("🆕 New Editor Ratio:   %.1f%%\n", profile.QualityMetrics.NewEditorRatio*100))
	output.WriteString(fmt.Sprintf("🏆 Contributor Diversity: %.2f/1.00\n", profile.QualityMetrics.ContributorDiversity))

	if profile.QualityMetrics.RecentActivityBurst {
		output.WriteString("💥 Recent Activity:    " + warningColor.Sprint("HIGH BURST DETECTED") + "\n")
	} else {
		output.WriteString("💥 Recent Activity:    " + successColor.Sprint("Normal") + "\n")
	}
	output.WriteString("\n")

	// Edit frequency
	output.WriteString(headerColor.Sprint("📈 EDIT FREQUENCY\n"))
	output.WriteString(strings.Repeat("─", 50) + "\n")

	output.WriteString("📅 Last 7 days:       " + strconv.Itoa(profile.QualityMetrics.EditFrequency.EditsLast7Days) + " edits\n")
	output.WriteString("📅 Last 30 days:      " + strconv.Itoa(profile.QualityMetrics.EditFrequency.EditsLast30Days) + " edits\n")
	output.WriteString("📅 Last 90 days:      " + strconv.Itoa(profile.QualityMetrics.EditFrequency.EditsLast90Days) + " edits\n")

	if len(profile.QualityMetrics.EditFrequency.PeakEditingHours) > 0 {
		hours := make([]string, len(profile.QualityMetrics.EditFrequency.PeakEditingHours))
		for i, hour := range profile.QualityMetrics.EditFrequency.PeakEditingHours {
			hours[i] = fmt.Sprintf("%02d:00", hour)
		}
		output.WriteString("🕐 Peak Hours:         " + strings.Join(hours, ", ") + "\n")
	}
	output.WriteString("\n")

	// Top contributors
	if len(profile.Contributors) > 0 {
		output.WriteString(headerColor.Sprint("👥 TOP CONTRIBUTORS\n"))
		output.WriteString(strings.Repeat("─", 80) + "\n")

		for i, contributor := range profile.Contributors {
			if i >= 10 { // Limit to top 10
				break
			}

			username := contributor.Username
			if len(username) > 25 {
				username = username[:25] + "..."
			}

			userType := "👤"
			if contributor.IsAnonymous {
				userType = "🌐"
				username = secondaryColor.Sprint(username)
			}

			output.WriteString(fmt.Sprintf("%s %-28s %4d edits %+6d bytes %s\n",
				userType,
				username,
				contributor.EditCount,
				contributor.TotalSizeDiff,
				contributor.LastEdit.Format("02/01/06"),
			))
		}
		output.WriteString("\n")
	}

	// Recent revisions (preview)
	if len(profile.RecentRevisions) > 0 {
		output.WriteString(headerColor.Sprint("🕒 RECENT REVISIONS (last 10)\n"))
		output.WriteString(strings.Repeat("─", 80) + "\n")

		for i, revision := range profile.RecentRevisions {
			if i >= 10 {
				break
			}

			username := revision.Username
			if len(username) > 20 {
				username = username[:20] + "..."
			}

			comment := revision.Comment
			if len(comment) > 30 {
				comment = comment[:30] + "..."
			}
			if comment == "" {
				comment = secondaryColor.Sprint("(no comment)")
			}

			diffStr := fmt.Sprintf("%+d", revision.SizeDiff)
			if revision.SizeDiff > 0 {
				diffStr = successColor.Sprint(diffStr)
			} else if revision.SizeDiff < 0 {
				diffStr = warningColor.Sprint(diffStr)
			}

			revertFlag := ""
			if revision.IsRevert {
				revertFlag = dangerColor.Sprint(" [REVERT]")
			}

			output.WriteString(fmt.Sprintf("%-12s %-22s %s %s%s\n",
				revision.Timestamp.Format("02/01 15:04"),
				username,
				diffStr,
				comment,
				revertFlag,
			))
		}
		output.WriteString("\n")
	}

	// Footer
	output.WriteString(secondaryColor.Sprint("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n"))
	output.WriteString(secondaryColor.Sprintf("📊 WikiOSINT Page Analysis - %d revisions analyzed on %s.wikipedia.org\n",
		len(profile.RecentRevisions), profile.Language))

	return output.String()
}

// formatPageHistoryAsTable formats page history analysis
func formatPageHistoryAsTable(profile *models.PageProfile) string {
	// This could be a more detailed history-focused view
	// For now, reuse the main formatter but could be enhanced
	return formatPageAsTable(profile)
}

// formatPageConflictsAsTable formats page conflict analysis
func formatPageConflictsAsTable(profile *models.PageProfile) string {
	var output strings.Builder

	// Header
	output.WriteString(headerColor.Sprint("╭─────────────────────────────────────────────────────────────╮\n"))
	output.WriteString(headerColor.Sprintf("│  ⚔️ CONFLICT ANALYSIS: %-32s      │\n", truncateString(profile.PageTitle, 32)))
	output.WriteString(headerColor.Sprint("╰─────────────────────────────────────────────────────────────╯\n\n"))

	// Conflict overview
	output.WriteString(headerColor.Sprint("📊 CONFLICT OVERVIEW\n"))
	output.WriteString(strings.Repeat("─", 50) + "\n")

	output.WriteString("🔄 Total Reversions:   " + strconv.Itoa(profile.ConflictStats.ReversionsCount) + "\n")
	output.WriteString("📅 Recent Conflicts:   " + strconv.Itoa(profile.ConflictStats.RecentConflicts) + " (last 7 days)\n")
	output.WriteString(fmt.Sprintf("📈 Stability Score:    %.2f/1.00 ", profile.ConflictStats.StabilityScore))

	if profile.ConflictStats.StabilityScore < 0.7 {
		output.WriteString(dangerColor.Sprint("(UNSTABLE)"))
	} else if profile.ConflictStats.StabilityScore < 0.9 {
		output.WriteString(warningColor.Sprint("(MODERATE)"))
	} else {
		output.WriteString(successColor.Sprint("(STABLE)"))
	}
	output.WriteString("\n")

	output.WriteString(fmt.Sprintf("⚡ Controversy Score:  %.2f ", profile.ConflictStats.ControversyScore))
	if profile.ConflictStats.ControversyScore > 0.3 {
		output.WriteString(dangerColor.Sprint("(HIGH CONTROVERSY)"))
	} else if profile.ConflictStats.ControversyScore > 0.1 {
		output.WriteString(warningColor.Sprint("(SOME CONTROVERSY)"))
	} else {
		output.WriteString(successColor.Sprint("(LOW CONTROVERSY)"))
	}
	output.WriteString("\n\n")

	// Conflicting users
	if len(profile.ConflictStats.ConflictingUsers) > 0 {
		output.WriteString(headerColor.Sprint("👥 USERS INVOLVED IN CONFLICTS\n"))
		output.WriteString(strings.Repeat("─", 50) + "\n")
		for i, user := range profile.ConflictStats.ConflictingUsers {
			if i >= 10 { // Limit to 10
				output.WriteString(fmt.Sprintf("... and %d more users\n", len(profile.ConflictStats.ConflictingUsers)-10))
				break
			}
			output.WriteString("🔸 " + user + "\n")
		}
		output.WriteString("\n")
	}

	// Edit war periods
	if len(profile.ConflictStats.EditWarPeriods) > 0 {
		output.WriteString(headerColor.Sprint("💥 DETECTED EDIT WAR PERIODS\n"))
		output.WriteString(strings.Repeat("─", 50) + "\n")
		for i, period := range profile.ConflictStats.EditWarPeriods {
			if i >= 5 { // Limit to 5 most recent
				break
			}
			output.WriteString(fmt.Sprintf("📅 %s - %s\n",
				period.StartTime.Format("02/01 15:04"),
				period.EndTime.Format("02/01 15:04")))
			output.WriteString(fmt.Sprintf("   👥 Participants: %s\n", strings.Join(period.Participants, ", ")))
			output.WriteString(fmt.Sprintf("   📊 Revisions: %d\n\n", period.RevisionCount))
		}
	}

	// Recent reverts
	revertCount := 0
	output.WriteString(headerColor.Sprint("🔄 RECENT REVERTS\n"))
	output.WriteString(strings.Repeat("─", 50) + "\n")

	for _, revision := range profile.RecentRevisions {
		if revision.IsRevert {
			revertCount++
			if revertCount > 10 { // Limit to 10
				break
			}

			output.WriteString(fmt.Sprintf("%-12s %-20s %s\n",
				revision.Timestamp.Format("02/01 15:04"),
				revision.Username,
				revision.Comment,
			))
		}
	}

	if revertCount == 0 {
		output.WriteString(successColor.Sprint("No recent reverts detected\n"))
	}
	output.WriteString("\n")

	// Footer
	output.WriteString(secondaryColor.Sprint("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n"))
	output.WriteString(secondaryColor.Sprintf("⚔️ WikiOSINT Conflict Analysis - %s.wikipedia.org\n", profile.Language))

	return output.String()
}

// Helper functions for page formatting

// formatPageSuspicionFlag formats page suspicion flags into readable text
func formatPageSuspicionFlag(flag string) string {
	switch flag {
	case "PAGE_HIGH_CONFLICT":
		return "High conflict ratio detected"
	case "PAGE_FEW_CONTRIBUTORS":
		return "Too few contributors for edit volume"
	case "PAGE_RECENT_INTENSIVE_ACTIVITY":
		return "Recent intensive editing activity"
	case "PAGE_ANONYMOUS_HEAVY_EDITING":
		return "Heavy anonymous editing"
	case "PAGE_NEW_EDITOR_DOMINANCE":
		return "Dominated by new editor accounts"
	case "PAGE_LOW_DIVERSITY":
		return "Low contributor diversity"
	case "PAGE_RECENT_CONFLICTS":
		return "Recent editing conflicts detected"
	default:
		return flag
	}
}

// truncateString truncates a string to specified length
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// getSuspicionText returns descriptive text for suspicion score
func getSuspicionText(score int) string {
	switch {
	case score >= 80:
		return "VERY HIGH"
	case score >= 60:
		return "HIGH"
	case score >= 40:
		return "MODERATE"
	case score >= 20:
		return "LOW"
	default:
		return "MINIMAL"
	}
}

// getSuspicionColor returns appropriate color for the score
func getSuspicionColor(score int) *color.Color {
	switch {
	case score >= 80:
		return dangerColor
	case score >= 60:
		return color.New(color.FgRed)
	case score >= 40:
		return warningColor
	case score >= 20:
		return color.New(color.FgYellow)
	default:
		return successColor
	}
}

// formatSuspicionFlag formats suspicion flags into readable text
func formatSuspicionFlag(flag string) string {
	switch flag {
	case "RECENT_ACCOUNT_HIGH_ACTIVITY":
		return "Recent account with intense activity"
	case "USER_BLOCKED":
		return "User currently blocked"
	case "SINGLE_PAGE_FOCUS":
		return "Excessive focus on single page"
	case "NO_SPECIAL_GROUPS":
		return "No special groups despite activity"
	case "SENSITIVE_NAMESPACE_FOCUS":
		return "Edits only in sensitive namespaces"
	case "FREQUENT_EMPTY_COMMENTS":
		return "Edit comments frequently empty"
	default:
		return flag
	}
}
