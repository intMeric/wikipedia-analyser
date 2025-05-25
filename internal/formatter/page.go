// internal/formatter/page.go
package formatter

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/intMeric/wikipedia-analyser/internal/models"
	"gopkg.in/yaml.v2"
)

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
	output.WriteString(headerColor.Sprintf("│  📄 WIKIPEDIA PAGE ANALYSIS: %-27s │\n", truncateString(profile.PageTitle, 27)))
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
		output.WriteString(headerColor.Sprint("👥 TOP CONTRIBUTORS ANALYSIS\n"))
		output.WriteString(strings.Repeat("─", 80) + "\n")

		for i, contributor := range profile.Contributors {
			if i >= 15 { // Limit to top 15
				break
			}

			username := contributor.Username
			if len(username) > 22 {
				username = username[:22] + "..."
			}

			userType := "👤"
			suspicionDisplay := ""

			if contributor.IsAnonymous {
				userType = "🌐"
				username = secondaryColor.Sprint(username)
				suspicionDisplay = secondaryColor.Sprint("(Anonymous)")
			} else {
				// Display suspicion score with color
				if contributor.SuspicionScore == -1 {
					suspicionDisplay = warningColor.Sprint("(Analysis failed)")
				} else {
					suspicionText := getSuspicionText(contributor.SuspicionScore)
					suspicionColor := getSuspicionColor(contributor.SuspicionScore)
					suspicionDisplay = fmt.Sprintf("%s (%d/100)",
						suspicionColor.Sprint(suspicionText),
						contributor.SuspicionScore)
				}
			}

			output.WriteString(fmt.Sprintf("%s %-25s %4d edits %+6d bytes %s %s\n",
				userType,
				username,
				contributor.EditCount,
				contributor.TotalSizeDiff,
				contributor.LastEdit.Format("02/01/06"),
				suspicionDisplay,
			))

			// Show contributor-specific flags if any (and not anonymous)
			if !contributor.IsAnonymous && len(contributor.SuspicionFlags) > 0 && i < 10 {
				contributorFlags := filterContributorFlags(contributor.SuspicionFlags)
				if len(contributorFlags) > 0 {
					flagsText := strings.Join(contributorFlags[:min(3, len(contributorFlags))], ", ")
					output.WriteString(fmt.Sprintf("   📋 %s\n", secondaryColor.Sprint(flagsText)))
				}
			}
		}
		output.WriteString("\n")
	}

	// Suspicious contributors section
	suspiciousContributors := []models.TopContributor{}
	for _, contributor := range profile.Contributors {
		if !contributor.IsAnonymous && contributor.SuspicionScore >= 40 {
			suspiciousContributors = append(suspiciousContributors, contributor)
		}
	}

	if len(suspiciousContributors) > 0 {
		output.WriteString(warningColor.Sprint("🚨 SUSPICIOUS CONTRIBUTORS DETECTED\n"))
		output.WriteString(strings.Repeat("─", 50) + "\n")

		for i, contributor := range suspiciousContributors {
			if i >= 5 { // Limit to 5 most suspicious
				break
			}

			suspicionText := getSuspicionText(contributor.SuspicionScore)
			suspicionColor := getSuspicionColor(contributor.SuspicionScore)

			output.WriteString(fmt.Sprintf("⚠️  %s - %s (%d/100)\n",
				contributor.Username,
				suspicionColor.Sprint(suspicionText),
				contributor.SuspicionScore,
			))

			// Show top flags for this contributor
			if len(contributor.SuspicionFlags) > 0 {
				topFlags := contributor.SuspicionFlags[:min(3, len(contributor.SuspicionFlags))]
				for _, flag := range topFlags {
					flagText := formatContributorSuspicionFlag(flag)
					output.WriteString(fmt.Sprintf("   🔸 %s\n", secondaryColor.Sprint(flagText)))
				}
			}
			output.WriteString("\n")
		}
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
	output.WriteString(headerColor.Sprintf("│  ⚔️ CONFLICT ANALYSIS: %-32s │\n", truncateString(profile.PageTitle, 32)))
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

// filterContributorFlags filters and formats contributor-specific flags
func filterContributorFlags(flags []string) []string {
	var filtered []string
	flagDescriptions := map[string]string{
		"HIGH_PAGE_ACTIVITY":             "High page activity",
		"NEW_ACCOUNT_HIGH_PAGE_ACTIVITY": "New account, high activity",
		"VERY_RECENT_ACTIVITY":           "Very recent edits",
		"LARGE_CONTENT_CHANGES":          "Large content changes",
		"RECENT_ACCOUNT_HIGH_ACTIVITY":   "Recent account, active",
		"USER_BLOCKED":                   "Currently blocked",
		"SINGLE_PAGE_FOCUS":              "Single page focus",
		"NO_SPECIAL_GROUPS":              "No special groups",
		"SENSITIVE_NAMESPACE_FOCUS":      "Sensitive namespace focus",
		"FREQUENT_EMPTY_COMMENTS":        "Empty comments",
	}

	for _, flag := range flags {
		if description, exists := flagDescriptions[flag]; exists {
			filtered = append(filtered, description)
		}
	}

	return filtered
}

// formatContributorSuspicionFlag formats contributor suspicion flags into readable text
func formatContributorSuspicionFlag(flag string) string {
	switch flag {
	case "HIGH_PAGE_ACTIVITY":
		return "Unusually high activity on this page"
	case "NEW_ACCOUNT_HIGH_PAGE_ACTIVITY":
		return "New account with intense page activity"
	case "VERY_RECENT_ACTIVITY":
		return "Very recent editing activity"
	case "LARGE_CONTENT_CHANGES":
		return "Made large content modifications"
	case "RECENT_ACCOUNT_HIGH_ACTIVITY":
		return "Recent account with high overall activity"
	case "USER_BLOCKED":
		return "Currently blocked user"
	case "SINGLE_PAGE_FOCUS":
		return "Focuses primarily on single pages"
	case "NO_SPECIAL_GROUPS":
		return "No special user groups despite activity"
	case "SENSITIVE_NAMESPACE_FOCUS":
		return "Edits mainly in sensitive namespaces"
	case "FREQUENT_EMPTY_COMMENTS":
		return "Often leaves empty edit comments"
	case "ANONYMOUS_USER":
		return "Anonymous IP address"
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
