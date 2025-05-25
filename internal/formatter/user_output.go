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
