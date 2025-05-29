// internal/formatter/pages.go
package formatter

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/intMeric/wikipedia-analyser/internal/models"
)

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

// formatPageHistoryAsTable formats page history analysis with focus on temporal patterns
func formatPageHistoryAsTable(profile *models.PageProfile) string {
	var output strings.Builder

	// Header with page title
	output.WriteString(headerColor.Sprint("╭─────────────────────────────────────────────────────────────╮\n"))
	output.WriteString(headerColor.Sprintf("│  📚 EDIT HISTORY ANALYSIS: %-29s │\n", truncateString(profile.PageTitle, 29)))
	output.WriteString(headerColor.Sprint("╰─────────────────────────────────────────────────────────────╯\n\n"))

	// Basic page info
	output.WriteString(headerColor.Sprint("📋 PAGE OVERVIEW\n"))
	output.WriteString(strings.Repeat("─", 50) + "\n")
	output.WriteString("📄 Page Title:         " + profile.PageTitle + "\n")
	output.WriteString("📊 Total Revisions:    " + strconv.Itoa(profile.TotalRevisions) + "\n")
	output.WriteString("👥 Total Contributors: " + strconv.Itoa(len(profile.Contributors)) + "\n")
	output.WriteString("🔄 Last Modified:      " + profile.LastModified.Format("02/01/2006 15:04") + "\n")
	output.WriteString("\n")

	// Edit frequency analysis
	output.WriteString(headerColor.Sprint("📈 EDITING ACTIVITY TIMELINE\n"))
	output.WriteString(strings.Repeat("─", 50) + "\n")

	output.WriteString("📅 Last 7 days:       " + strconv.Itoa(profile.QualityMetrics.EditFrequency.EditsLast7Days) + " edits\n")
	output.WriteString("📅 Last 30 days:      " + strconv.Itoa(profile.QualityMetrics.EditFrequency.EditsLast30Days) + " edits\n")
	output.WriteString("📅 Last 90 days:      " + strconv.Itoa(profile.QualityMetrics.EditFrequency.EditsLast90Days) + " edits\n")

	if profile.QualityMetrics.RecentActivityBurst {
		output.WriteString("💥 Activity Pattern:   " + warningColor.Sprint("RECENT BURST DETECTED") + "\n")
	} else {
		output.WriteString("💥 Activity Pattern:   " + successColor.Sprint("Normal distribution") + "\n")
	}

	if len(profile.QualityMetrics.EditFrequency.PeakEditingHours) > 0 {
		hours := make([]string, len(profile.QualityMetrics.EditFrequency.PeakEditingHours))
		for i, hour := range profile.QualityMetrics.EditFrequency.PeakEditingHours {
			hours[i] = fmt.Sprintf("%02d:00", hour)
		}
		output.WriteString("🕐 Peak Hours:         " + strings.Join(hours, ", ") + "\n")
	}
	output.WriteString("\n")

	// Daily activity breakdown
	if len(profile.QualityMetrics.EditFrequency.EditsByDay) > 0 {
		output.WriteString(headerColor.Sprint("📅 DAILY ACTIVITY BREAKDOWN\n"))
		output.WriteString(strings.Repeat("─", 50) + "\n")

		// Show last 14 days of activity
		count := 0
		for date, edits := range profile.QualityMetrics.EditFrequency.EditsByDay {
			if count >= 14 {
				break
			}
			intensity := ""
			if edits > 10 {
				intensity = warningColor.Sprint(" (High)")
			} else if edits > 5 {
				intensity = infoColor.Sprint(" (Moderate)")
			}
			output.WriteString(fmt.Sprintf("📆 %s: %2d edits%s\n", date, edits, intensity))
			count++
		}
		output.WriteString("\n")
	}

	// Detailed revision history
	if len(profile.RecentRevisions) > 0 {
		output.WriteString(headerColor.Sprint("🕒 DETAILED REVISION HISTORY\n"))
		output.WriteString(strings.Repeat("─", 85) + "\n")

		for i, revision := range profile.RecentRevisions {
			if i >= 20 { // Show more revisions for history view
				break
			}

			username := revision.Username
			if len(username) > 18 {
				username = username[:18] + "..."
			}

			comment := revision.Comment
			if len(comment) > 35 {
				comment = comment[:35] + "..."
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

			minorFlag := ""
			if revision.IsMinor {
				minorFlag = secondaryColor.Sprint(" [m]")
			}

			output.WriteString(fmt.Sprintf("%-12s %-20s %s %s%s%s\n",
				revision.Timestamp.Format("02/01 15:04"),
				username,
				diffStr,
				comment,
				revertFlag,
				minorFlag,
			))
		}
		output.WriteString("\n")
	}

	// Contributor activity patterns
	if len(profile.Contributors) > 0 {
		output.WriteString(headerColor.Sprint("👥 CONTRIBUTOR ACTIVITY PATTERNS\n"))
		output.WriteString(strings.Repeat("─", 70) + "\n")

		for i, contributor := range profile.Contributors {
			if i >= 10 { // Top 10 for history view
				break
			}

			username := contributor.Username
			if len(username) > 20 {
				username = username[:20] + "..."
			}

			userType := "👤"
			if contributor.IsAnonymous {
				userType = "🌐"
				username = secondaryColor.Sprint(username)
			}

			// Calculate activity span
			activitySpan := int(contributor.LastEdit.Sub(contributor.FirstEdit).Hours() / 24)
			avgEditsPerDay := float64(contributor.EditCount) / float64(max(1, activitySpan))

			output.WriteString(fmt.Sprintf("%s %-25s %3d edits over %3d days (%.1f/day)\n",
				userType,
				username,
				contributor.EditCount,
				activitySpan,
				avgEditsPerDay,
			))

			// Show editing pattern
			if avgEditsPerDay > 5 {
				output.WriteString(fmt.Sprintf("   📊 %s\n", warningColor.Sprint("High-intensity editing pattern")))
			} else if avgEditsPerDay > 2 {
				output.WriteString(fmt.Sprintf("   📊 %s\n", infoColor.Sprint("Regular editing pattern")))
			} else {
				output.WriteString(fmt.Sprintf("   📊 %s\n", secondaryColor.Sprint("Occasional editing pattern")))
			}
		}
		output.WriteString("\n")
	}

	// Footer
	output.WriteString(secondaryColor.Sprint("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n"))
	output.WriteString(secondaryColor.Sprintf("📚 WikiOSINT History Analysis - %s.wikipedia.org\n", profile.Language))

	return output.String()
}

// formatPageConflictsAsTable formats page conflict analysis with focus on disputes
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

	// Conflict severity assessment
	output.WriteString(headerColor.Sprint("🚨 CONFLICT SEVERITY ASSESSMENT\n"))
	output.WriteString(strings.Repeat("─", 50) + "\n")

	conflictLevel := "🟢 LOW"
	if profile.ConflictStats.ControversyScore > 0.3 || profile.ConflictStats.RecentConflicts > 10 {
		conflictLevel = dangerColor.Sprint("🔴 HIGH")
	} else if profile.ConflictStats.ControversyScore > 0.1 || profile.ConflictStats.RecentConflicts > 5 {
		conflictLevel = warningColor.Sprint("🟡 MODERATE")
	} else {
		conflictLevel = successColor.Sprint("🟢 LOW")
	}

	output.WriteString("🎯 Conflict Level:     " + conflictLevel + "\n")
	output.WriteString(fmt.Sprintf("📈 Reversion Rate:     %.1f%% of total edits\n",
		float64(profile.ConflictStats.ReversionsCount)/float64(max(1, profile.TotalRevisions))*100))

	if profile.ConflictStats.RecentConflicts > 0 {
		output.WriteString("⚠️  Recent Activity:    " + warningColor.Sprint("Active conflicts detected") + "\n")
	} else {
		output.WriteString("✅ Recent Activity:    " + successColor.Sprint("No recent conflicts") + "\n")
	}
	output.WriteString("\n")

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
		output.WriteString(strings.Repeat("─", 70) + "\n")
		for i, period := range profile.ConflictStats.EditWarPeriods {
			if i >= 5 { // Limit to 5 most recent
				break
			}

			duration := period.EndTime.Sub(period.StartTime)
			output.WriteString(fmt.Sprintf("📅 %s - %s (%s duration)\n",
				period.StartTime.Format("02/01 15:04"),
				period.EndTime.Format("02/01 15:04"),
				duration.String()))
			output.WriteString(fmt.Sprintf("   👥 Participants: %s\n", strings.Join(period.Participants, ", ")))
			output.WriteString(fmt.Sprintf("   📊 Revisions: %d ", period.RevisionCount))

			// Intensity assessment
			if duration.Hours() > 0 {
				intensity := float64(period.RevisionCount) / duration.Hours()
				if intensity > 2 {
					output.WriteString(dangerColor.Sprint("(Very Intense)"))
				} else if intensity > 1 {
					output.WriteString(warningColor.Sprint("(Intense)"))
				} else {
					output.WriteString(infoColor.Sprint("(Moderate)"))
				}
			} else {
				output.WriteString(warningColor.Sprint("(Simultaneous)"))
			}
			output.WriteString("\n\n")
		}
	}

	// Recent reverts analysis
	revertCount := 0
	output.WriteString(headerColor.Sprint("🔄 RECENT REVERT ANALYSIS\n"))
	output.WriteString(strings.Repeat("─", 75) + "\n")

	for _, revision := range profile.RecentRevisions {
		if revision.IsRevert {
			revertCount++
			if revertCount > 15 { // Limit to 15
				break
			}

			username := revision.Username
			if len(username) > 18 {
				username = username[:18] + "..."
			}

			comment := revision.Comment
			if len(comment) > 30 {
				comment = comment[:30] + "..."
			}

			output.WriteString(fmt.Sprintf("%-12s %-20s %s\n",
				revision.Timestamp.Format("02/01 15:04"),
				username,
				comment,
			))
		}
	}

	if revertCount == 0 {
		output.WriteString(successColor.Sprint("✅ No recent reverts detected - page appears stable\n"))
	} else {
		output.WriteString(fmt.Sprintf("\n📊 Total recent reverts shown: %d\n", revertCount))
	}
	output.WriteString("\n")

	// Recommendations
	output.WriteString(headerColor.Sprint("💡 CONFLICT MANAGEMENT RECOMMENDATIONS\n"))
	output.WriteString(strings.Repeat("─", 50) + "\n")

	if profile.ConflictStats.ControversyScore > 0.3 {
		output.WriteString(dangerColor.Sprint("🚨 HIGH PRIORITY ACTIONS NEEDED:\n"))
		output.WriteString("   • Consider page protection or editing restrictions\n")
		output.WriteString("   • Review user conduct and consider blocks if needed\n")
		output.WriteString("   • Initiate dispute resolution procedures\n")
		output.WriteString("   • Monitor for sockpuppet activity\n")
	} else if profile.ConflictStats.ControversyScore > 0.1 {
		output.WriteString(warningColor.Sprint("⚠️ MONITORING RECOMMENDED:\n"))
		output.WriteString("   • Watch for escalation patterns\n")
		output.WriteString("   • Consider discussion page mediation\n")
		output.WriteString("   • Document conflict patterns\n")
	} else {
		output.WriteString(successColor.Sprint("✅ PAGE STATUS: STABLE\n"))
		output.WriteString("   • Continue regular monitoring\n")
		output.WriteString("   • No immediate action required\n")
	}
	output.WriteString("\n")

	// Footer
	output.WriteString(secondaryColor.Sprint("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n"))
	output.WriteString(secondaryColor.Sprintf("⚔️ WikiOSINT Conflict Analysis - %s.wikipedia.org\n", profile.Language))

	return output.String()
}
