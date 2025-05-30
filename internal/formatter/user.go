// internal/formatter/user.go
package formatter

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/intMeric/wikipedia-analyser/internal/models"
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
		return formatUserAsJSON(profile)
	case "yaml", "yml":
		return formatUserAsYAML(profile)
	case "table", "":
		return formatUserAsTable(profile), nil
	default:
		return "", fmt.Errorf("unsupported format: %s (supported: table, json, yaml)", format)
	}
}

// formatUserAsJSON formats user profile as JSON
func formatUserAsJSON(profile *models.UserProfile) (string, error) {
	data, err := json.MarshalIndent(profile, "", "  ")
	if err != nil {
		return "", fmt.Errorf("JSON formatting error: %w", err)
	}
	return string(data), nil
}

// formatUserAsYAML formats user profile as YAML
func formatUserAsYAML(profile *models.UserProfile) (string, error) {
	data, err := yaml.Marshal(profile)
	if err != nil {
		return "", fmt.Errorf("YAML formatting error: %w", err)
	}
	return string(data), nil
}

// formatUserAsTable formats user profile as readable table
func formatUserAsTable(profile *models.UserProfile) string {
	var output strings.Builder

	// Header with username and suspicion score
	output.WriteString(headerColor.Sprint("╭─────────────────────────────────────────────────────────────╮\n"))
	output.WriteString(headerColor.Sprintf("│  📊 WIKIPEDIA USER PROFILE: %-27s │\n", profile.Username))
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

	// Add revoked contributions percentage in basic info
	if profile.RevokedCount > 0 {
		revokedPercentage := profile.RevokedRatio * 100
		var revokedDisplay string
		if revokedPercentage > 50 {
			revokedDisplay = dangerColor.Sprintf("%.1f%% (VERY HIGH)", revokedPercentage)
		} else if revokedPercentage > 30 {
			revokedDisplay = warningColor.Sprintf("%.1f%% (HIGH)", revokedPercentage)
		} else if revokedPercentage > 20 {
			revokedDisplay = warningColor.Sprintf("%.1f%% (MODERATE)", revokedPercentage)
		} else if revokedPercentage > 10 {
			revokedDisplay = infoColor.Sprintf("%.1f%% (LOW)", revokedPercentage)
		} else {
			revokedDisplay = successColor.Sprintf("%.1f%% (MINIMAL)", revokedPercentage)
		}
		output.WriteString("🚫 Revoked Ratio:      " + revokedDisplay + "\n")
	} else {
		output.WriteString("🚫 Revoked Ratio:      " + successColor.Sprint("0.0% (NONE)") + "\n")
	}

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
			flagText := formatUserSuspicionFlag(flag)
			output.WriteString(fmt.Sprintf("🔸 %s\n", warningColor.Sprint(flagText)))
		}
		output.WriteString("\n")
	}

	// Revoked contributions analysis
	if profile.RevokedCount > 0 {
		output.WriteString(warningColor.Sprint("🚫 REVOKED CONTRIBUTIONS ANALYSIS\n"))
		output.WriteString(strings.Repeat("─", 50) + "\n")

		output.WriteString("🔄 Total Revoked:      " + strconv.Itoa(profile.RevokedCount) + "\n")
		output.WriteString(fmt.Sprintf("📊 Revoked Ratio:      %.1f%% of all contributions\n", profile.RevokedRatio*100))

		// Display suspicion level based on ratio
		var revokedStatus string
		if profile.RevokedRatio > 0.5 {
			revokedStatus = dangerColor.Sprint("VERY HIGH - Potential vandal")
		} else if profile.RevokedRatio > 0.3 {
			revokedStatus = warningColor.Sprint("HIGH - Suspicious activity")
		} else if profile.RevokedRatio > 0.2 {
			revokedStatus = warningColor.Sprint("MODERATE - Needs monitoring")
		} else if profile.RevokedRatio > 0.1 {
			revokedStatus = infoColor.Sprint("LOW - Some issues")
		} else {
			revokedStatus = successColor.Sprint("MINIMAL - Normal conflicts")
		}

		output.WriteString("⚠️  Risk Level:        " + revokedStatus + "\n")

		// Analyze revert types
		revertTypes := make(map[string]int)
		for _, revoked := range profile.RevokedContribs {
			revertTypes[revoked.RevertType]++
		}

		if len(revertTypes) > 0 {
			output.WriteString("📋 Revert Types:\n")
			for revertType, count := range revertTypes {
				typeDescription := formatRevertType(revertType)
				output.WriteString(fmt.Sprintf("   • %s: %d times\n", typeDescription, count))
			}
		}

		// Top users who revert this user
		if len(profile.RevertedByUsers) > 0 {
			output.WriteString("👥 Most Frequent Reverters:\n")

			// Sort by number of reverts
			type userRevertCount struct {
				user  string
				count int
			}

			var reverterList []userRevertCount
			for user, count := range profile.RevertedByUsers {
				reverterList = append(reverterList, userRevertCount{user, count})
			}

			sort.Slice(reverterList, func(i, j int) bool {
				return reverterList[i].count > reverterList[j].count
			})

			// Display top 5
			for i, reverter := range reverterList {
				if i >= 5 {
					break
				}

				percentage := float64(reverter.count) / float64(profile.RevokedCount) * 100
				output.WriteString(fmt.Sprintf("   • %s: %d reverts (%.1f%%)\n",
					reverter.user, reverter.count, percentage))
			}
		}
		output.WriteString("\n")
	}

	// Detailed revoked contributions list
	if len(profile.RevokedContribs) > 0 {
		output.WriteString(dangerColor.Sprint("📋 DETAILED REVOKED CONTRIBUTIONS\n"))
		output.WriteString(strings.Repeat("─", 100) + "\n")

		// Sort revoked contributions by date (most recent first)
		sortedRevoked := make([]models.RevokedContribution, len(profile.RevokedContribs))
		copy(sortedRevoked, profile.RevokedContribs)
		sort.Slice(sortedRevoked, func(i, j int) bool {
			return sortedRevoked[i].OriginalContrib.Timestamp.After(sortedRevoked[j].OriginalContrib.Timestamp)
		})

		// Limit to most recent 20 for readability, but show all if <= 20
		displayCount := len(sortedRevoked)
		if displayCount > 20 {
			displayCount = 20
			output.WriteString(fmt.Sprintf("📊 Showing 20 most recent revoked contributions (total: %d)\n\n", len(sortedRevoked)))
		} else {
			output.WriteString(fmt.Sprintf("📊 All %d revoked contributions:\n\n", len(sortedRevoked)))
		}

		for i := range displayCount {
			revoked := sortedRevoked[i]
			contrib := revoked.OriginalContrib

			// Format page title
			title := contrib.PageTitle
			if len(title) > 35 {
				title = title[:35] + "..."
			}

			// Format comment
			comment := contrib.Comment
			if len(comment) > 30 {
				comment = comment[:30] + "..."
			}
			if comment == "" {
				comment = secondaryColor.Sprint("(no comment)")
			}

			// Format size diff
			diffStr := fmt.Sprintf("%+d", contrib.SizeDiff)
			if contrib.SizeDiff > 0 {
				diffStr = successColor.Sprint(diffStr)
			} else if contrib.SizeDiff < 0 {
				diffStr = warningColor.Sprint(diffStr)
			}

			// Calculate time between edit and revocation
			revertDelay := revoked.RevokedAt.Sub(contrib.Timestamp)
			var delayStr string
			if revertDelay < time.Hour {
				delayStr = fmt.Sprintf("%dm", int(revertDelay.Minutes()))
			} else if revertDelay < 24*time.Hour {
				delayStr = fmt.Sprintf("%dh", int(revertDelay.Hours()))
			} else {
				delayStr = fmt.Sprintf("%dd", int(revertDelay.Hours()/24))
			}

			// Format revert type with color
			revertTypeDisplay := formatRevertTypeShort(revoked.RevertType)
			var revertColor *color.Color
			switch revoked.RevertType {
			case "vandalism_revert":
				revertColor = dangerColor
			case "rollback":
				revertColor = warningColor
			default:
				revertColor = infoColor
			}

			// Format revoked by (truncate system names)
			revokedBy := revoked.RevokedBy
			if revokedBy == "system_detected" {
				revokedBy = secondaryColor.Sprint("system")
			} else if revokedBy == "detected" {
				revokedBy = secondaryColor.Sprint("detect")
			} else if len(revokedBy) > 15 {
				revokedBy = revokedBy[:15] + "..."
			}

			// Main line: Date | Page | Size | Comment | Reverted by | Delay | Type
			output.WriteString(fmt.Sprintf("%-12s %-37s %s %-32s rev:%s (%s) %s\n",
				contrib.Timestamp.Format("02/01 15:04"),
				title,
				diffStr,
				comment,
				revokedBy,
				delayStr,
				revertColor.Sprint(revertTypeDisplay),
			))

			// Second line: Revert comment (if meaningful and not too long)
			if revoked.RevertComment != "" &&
				revoked.RevertComment != "Detected from revision tags" &&
				len(strings.TrimSpace(revoked.RevertComment)) > 5 {
				revertComment := revoked.RevertComment
				if len(revertComment) > 80 {
					revertComment = revertComment[:80] + "..."
				}
				output.WriteString(fmt.Sprintf("             %s\n",
					secondaryColor.Sprintf("↳ \"%s\"", revertComment)))
			}
		}

		if len(sortedRevoked) > 20 {
			output.WriteString(fmt.Sprintf("\n... and %d more revoked contributions \n",
				len(sortedRevoked)-20))
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

	// Recent contributions (preview) - modified to show revocations
	if len(profile.RecentContribs) > 0 {
		output.WriteString(headerColor.Sprint("🕒 RECENT CONTRIBUTIONS (last 5)\n"))
		output.WriteString(strings.Repeat("─", 90) + "\n")

		for i, contrib := range profile.RecentContribs {
			if i >= 5 {
				break
			}

			title := contrib.PageTitle
			if len(title) > 30 {
				title = title[:30] + "..."
			}

			comment := contrib.Comment
			if len(comment) > 25 {
				comment = comment[:25] + "..."
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

			// Revocation indicator
			revokedIndicator := ""
			if contrib.IsRevoked {
				revokedIndicator = dangerColor.Sprint(" [REVOKED]")

				// Add who revoked and when
				revokedAge := int(time.Since(contrib.RevokedAt).Hours() / 24)
				if revokedAge == 0 {
					revokedIndicator += secondaryColor.Sprint(" by " + contrib.RevokedBy + " (today)")
				} else {
					revokedIndicator += secondaryColor.Sprintf(" by %s (%dd ago)",
						contrib.RevokedBy, revokedAge)
				}
			}

			output.WriteString(fmt.Sprintf("%-12s %-32s %s %s%s\n",
				contrib.Timestamp.Format("02/01 15:04"),
				title,
				diffStr,
				comment,
				revokedIndicator,
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

// formatUserSuspicionFlag formats user suspicion flags into readable text - FIXED
func formatUserSuspicionFlag(flag string) string {
	// Gérer les flags de conflit avec utilisateur spécifique
	if strings.HasPrefix(flag, "CONFLICT_WITH_SPECIFIC_USER_") {
		username := strings.TrimPrefix(flag, "CONFLICT_WITH_SPECIFIC_USER_")
		return fmt.Sprintf("Repeated conflicts with user: %s", username)
	}

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
	case "VERY_HIGH_REVOKED_RATIO":
		return "Very high ratio of revoked contributions (>50%)"
	case "HIGH_REVOKED_RATIO":
		return "High ratio of revoked contributions (>30%)"
	case "MODERATE_REVOKED_RATIO":
		return "Moderate ratio of revoked contributions (>20%)"
	case "MANY_REVOKED_CONTRIBUTIONS":
		return "Many contributions have been revoked (>50)"
	case "SOME_REVOKED_CONTRIBUTIONS":
		return "Several contributions have been revoked (>20)"
	case "VANDALISM_PATTERN":
		return "Pattern of vandalism-related reverts detected"
	case "SOME_VANDALISM_REVERTS":
		return "Some contributions reverted as vandalism"
	case "CONFLICT_WITH_SPECIFIC_USER":
		return "Repeated conflicts with specific user"
	case "NEW_ACCOUNT_MANY_REVERTS":
		return "New account with many revoked contributions"
	default:
		return flag
	}
}

// formatRevertType formats revert types into readable text
func formatRevertType(revertType string) string {
	switch revertType {
	case "vandalism_revert":
		return "Vandalism (serious)"
	case "rollback":
		return "Rollback (admin tool)"
	case "undo":
		return "Manual undo"
	case "restore":
		return "Content restoration"
	case "manual_revert":
		return "Manual revert"
	case "generic_revert":
		return "Generic revert"
	case "detected_light":
		return "Detected (light analysis)"
	default:
		return revertType
	}
}

// formatRevertTypeShort formats revert types into short readable text for compact display
func formatRevertTypeShort(revertType string) string {
	switch revertType {
	case "vandalism_revert":
		return "VANDAL"
	case "rollback":
		return "ROLLBACK"
	case "undo":
		return "UNDO"
	case "restore":
		return "RESTORE"
	case "manual_revert":
		return "REVERT"
	case "generic_revert":
		return "GENERIC"
	case "detected_light":
		return "DETECTED"
	default:
		return strings.ToUpper(revertType)
	}
}
