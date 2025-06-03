// internal/formatter/contribution.go
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

// FormatContributionProfile formats the contribution profile according to the specified format
func FormatContributionProfile(profile *models.ContributionProfile, format string) (string, error) {
	switch strings.ToLower(format) {
	case "json":
		return formatContributionAsJSON(profile)
	case "yaml", "yml":
		return formatContributionAsYAML(profile)
	case "table", "":
		return formatContributionAsTable(profile), nil
	default:
		return "", fmt.Errorf("unsupported format: %s (supported: table, json, yaml)", format)
	}
}

// formatContributionAsJSON formats contribution profile as JSON
func formatContributionAsJSON(profile *models.ContributionProfile) (string, error) {
	data, err := json.MarshalIndent(profile, "", "  ")
	if err != nil {
		return "", fmt.Errorf("JSON formatting error: %w", err)
	}
	return string(data), nil
}

// formatContributionAsYAML formats contribution profile as YAML
func formatContributionAsYAML(profile *models.ContributionProfile) (string, error) {
	data, err := yaml.Marshal(profile)
	if err != nil {
		return "", fmt.Errorf("YAML formatting error: %w", err)
	}
	return string(data), nil
}

// formatContributionAsTable formats contribution profile as readable table
func formatContributionAsTable(profile *models.ContributionProfile) string {
	var output strings.Builder

	// Header with revision ID and suspicion score
	output.WriteString(headerColor.Sprint("╭─────────────────────────────────────────────────────────────╮\n"))
	output.WriteString(headerColor.Sprintf("│  📝 CONTRIBUTION ANALYSIS: Revision %-18d │\n", profile.RevisionID))
	output.WriteString(headerColor.Sprint("╰─────────────────────────────────────────────────────────────╯\n\n"))

	// Suspicion score with color
	suspicionText := getSuspicionText(profile.SuspicionScore)
	suspicionColor := getSuspicionColor(profile.SuspicionScore)
	output.WriteString(fmt.Sprintf("🚨 %s %s (%d/100)\n\n",
		suspicionColor.Sprint("Suspicion Score:"),
		suspicionColor.Sprint(suspicionText),
		profile.SuspicionScore))

	// Basic information
	output.WriteString(headerColor.Sprint("📋 CONTRIBUTION INFORMATION\n"))
	output.WriteString(strings.Repeat("─", 50) + "\n")

	output.WriteString("📝 Revision ID:        " + strconv.Itoa(profile.RevisionID) + "\n")
	output.WriteString("📄 Page:               " + profile.PageTitle + "\n")
	output.WriteString("🌍 Language:           " + profile.Language + "\n")
	output.WriteString("⏰ Timestamp:          " + profile.Timestamp.Format("02/01/2006 15:04:05") + "\n")
	output.WriteString("📏 Size:               " + strconv.Itoa(profile.Size) + " bytes\n")

	if profile.IsMinor {
		output.WriteString("🔍 Edit Type:          " + infoColor.Sprint("Minor edit") + "\n")
	} else {
		output.WriteString("🔍 Edit Type:          " + "Major edit" + "\n")
	}

	if profile.IsRevert {
		output.WriteString("🔄 Revert Status:      " + warningColor.Sprint("This is a revert") + "\n")
	} else {
		output.WriteString("🔄 Revert Status:      " + successColor.Sprint("Regular edit") + "\n")
	}

	// Format comment
	comment := profile.Comment
	if comment == "" {
		comment = secondaryColor.Sprint("(no comment)")
	} else if len(comment) > 80 {
		comment = comment[:80] + "..."
	}
	output.WriteString("💬 Comment:            " + comment + "\n")
	output.WriteString("\n")

	// Suspicion flags
	if len(profile.SuspicionFlags) > 0 {
		output.WriteString(warningColor.Sprint("⚠️  SUSPICION INDICATORS\n"))
		output.WriteString(strings.Repeat("─", 50) + "\n")
		for _, flag := range profile.SuspicionFlags {
			flagText := formatContributionSuspicionFlag(flag)
			output.WriteString(fmt.Sprintf("🔸 %s\n", warningColor.Sprint(flagText)))
		}
		output.WriteString("\n")
	}

	// Author analysis
	output.WriteString(headerColor.Sprint("👤 AUTHOR ANALYSIS\n"))
	output.WriteString(strings.Repeat("─", 50) + "\n")

	author := profile.Author
	output.WriteString("👤 Username:           " + author.Username + "\n")

	if author.IsAnonymous {
		output.WriteString("🌐 User Type:          " + secondaryColor.Sprint("Anonymous IP") + "\n")
	} else {
		output.WriteString("🌐 User Type:          " + "Registered user\n")
		output.WriteString("🆔 User ID:            " + strconv.Itoa(author.UserID) + "\n")
		output.WriteString("✏️ Total Edits:        " + strconv.Itoa(author.EditCount) + "\n")

		if author.RegistrationDate != nil {
			regDate := author.RegistrationDate.Format("02/01/2006")
			daysSince := int(time.Since(*author.RegistrationDate).Hours() / 24)
			output.WriteString(fmt.Sprintf("📅 Registration:       %s (%d days ago)\n", regDate, daysSince))

			// New account warning
			if daysSince < 30 {
				output.WriteString("⚠️  Account Age:        " + warningColor.Sprint("Very new account") + "\n")
			}
		}

		if len(author.Groups) > 0 {
			output.WriteString("👥 Groups:             " + strings.Join(author.Groups, ", ") + "\n")
		}

		if author.IsBlocked {
			output.WriteString("🚫 Status:             " + dangerColor.Sprint("BLOCKED") + "\n")
		}

		// Author suspicion score
		if author.SuspicionScore > 0 {
			authorSuspicionText := getSuspicionText(author.SuspicionScore)
			authorSuspicionColor := getSuspicionColor(author.SuspicionScore)
			output.WriteString(fmt.Sprintf("🚨 Author Suspicion:   %s (%d/100)\n",
				authorSuspicionColor.Sprint(authorSuspicionText),
				author.SuspicionScore))
		}
	}
	output.WriteString("\n")

	// Recent activity
	if !author.IsAnonymous {
		output.WriteString(headerColor.Sprint("📊 RECENT ACTIVITY\n"))
		output.WriteString(strings.Repeat("─", 50) + "\n")

		activity := author.RecentActivity
		output.WriteString("📅 Last 24h:           " + strconv.Itoa(activity.EditsLast24h) + " edits\n")
		output.WriteString("📅 Last 7 days:        " + strconv.Itoa(activity.EditsLast7d) + " edits\n")
		output.WriteString("📅 Last 30 days:       " + strconv.Itoa(activity.EditsLast30d) + " edits\n")
		output.WriteString("📄 Pages edited:       " + strconv.Itoa(activity.PagesEdited) + " pages\n")

		// Activity intensity warnings
		if activity.EditsLast24h > 50 {
			output.WriteString("⚠️  Activity Level:     " + warningColor.Sprint("Very high (>50/day)") + "\n")
		} else if activity.EditsLast24h > 20 {
			output.WriteString("⚠️  Activity Level:     " + infoColor.Sprint("High (>20/day)") + "\n")
		}

		if activity.LastEditTime != nil {
			timeSince := time.Since(*activity.LastEditTime)
			if timeSince < time.Hour {
				output.WriteString("🕒 Last Edit:          " + warningColor.Sprintf("%d minutes ago", int(timeSince.Minutes())) + "\n")
			} else if timeSince < 24*time.Hour {
				output.WriteString("🕒 Last Edit:          " + infoColor.Sprintf("%d hours ago", int(timeSince.Hours())) + "\n")
			} else {
				output.WriteString("🕒 Last Edit:          " + fmt.Sprintf("%d days ago", int(timeSince.Hours()/24)) + "\n")
			}
		}
		output.WriteString("\n")
	}

	// Content analysis
	output.WriteString(headerColor.Sprint("📝 CONTENT ANALYSIS\n"))
	output.WriteString(strings.Repeat("─", 50) + "\n")

	content := profile.ContentAnalysis
	output.WriteString("📂 Content Type:       " + formatContentType(content.ContentType) + "\n")

	changes := content.TextChanges
	if changes.CharsAdded > 0 {
		output.WriteString("➕ Characters Added:   " + successColor.Sprint(strconv.Itoa(changes.CharsAdded)) + "\n")
	}
	if changes.CharsRemoved > 0 {
		output.WriteString("➖ Characters Removed: " + warningColor.Sprint(strconv.Itoa(changes.CharsRemoved)) + "\n")
	}

	if changes.WordsAdded > 0 {
		output.WriteString("📝 Words Added:        " + strconv.Itoa(changes.WordsAdded) + "\n")
	}
	if changes.WordsRemoved > 0 {
		output.WriteString("📝 Words Removed:      " + strconv.Itoa(changes.WordsRemoved) + "\n")
	}

	if changes.IsStructural {
		output.WriteString("🏗️  Change Type:        " + infoColor.Sprint("Structural changes") + "\n")
	} else if changes.IsTrivial {
		output.WriteString("🏗️  Change Type:        " + secondaryColor.Sprint("Trivial changes") + "\n")
	} else {
		output.WriteString("🏗️  Change Type:        " + "Content changes" + "\n")
	}

	if len(changes.SectionsAffected) > 0 {
		output.WriteString("📋 Sections Affected:  " + strings.Join(changes.SectionsAffected, ", ") + "\n")
	}

	// Language analysis
	language := content.LanguageAnalysis
	if len(language.POVWords) > 0 {
		output.WriteString("⚠️  POV Words Found:    " + warningColor.Sprint(strings.Join(language.POVWords, ", ")) + "\n")
		output.WriteString(fmt.Sprintf("📊 Bias Score:         %.2f/1.00\n", language.BiasScore))
	}

	output.WriteString("🎭 Tone Analysis:      " + formatToneAnalysis(language.ToneAnalysis) + "\n")
	output.WriteString("\n")

	// Quality metrics
	output.WriteString(headerColor.Sprint("🏆 QUALITY METRICS\n"))
	output.WriteString(strings.Repeat("─", 50) + "\n")

	quality := profile.QualityMetrics
	output.WriteString(fmt.Sprintf("📊 Overall Quality:    %.2f/1.00\n", quality.OverallQuality))
	output.WriteString(fmt.Sprintf("📝 Content Quality:    %.2f/1.00\n", quality.ContentQuality.Accuracy))
	output.WriteString(fmt.Sprintf("📚 Source Quality:     %.2f/1.00\n", quality.SourceQuality.ReliabilityScore))
	output.WriteString(fmt.Sprintf("🏗️  Structure Quality:  %.2f/1.00\n", quality.StructureQuality.Formatting))
	output.WriteString(fmt.Sprintf("📋 Policy Compliance:  %.2f/1.00\n", quality.ComplianceScore.PolicyCompliance))

	// Risk assessment
	compliance := quality.ComplianceScore
	if compliance.VandalismRisk > 0.3 {
		output.WriteString("⚠️  Vandalism Risk:     " + dangerColor.Sprintf("%.1f%% (HIGH)", compliance.VandalismRisk*100) + "\n")
	} else if compliance.VandalismRisk > 0.1 {
		output.WriteString("⚠️  Vandalism Risk:     " + warningColor.Sprintf("%.1f%% (MODERATE)", compliance.VandalismRisk*100) + "\n")
	} else {
		output.WriteString("⚠️  Vandalism Risk:     " + successColor.Sprintf("%.1f%% (LOW)", compliance.VandalismRisk*100) + "\n")
	}

	if len(compliance.ViolatedPolicies) > 0 {
		output.WriteString("🚫 Policy Violations:  " + dangerColor.Sprint(strings.Join(compliance.ViolatedPolicies, ", ")) + "\n")
	}
	output.WriteString("\n")

	// Context analysis (if available)
	if profile.ContextAnalysis.PageContext.Controversiality > 0 {
		output.WriteString(headerColor.Sprint("🌍 CONTEXT ANALYSIS\n"))
		output.WriteString(strings.Repeat("─", 50) + "\n")

		context := profile.ContextAnalysis
		pageContext := context.PageContext

		if pageContext.Controversiality > 0.5 {
			output.WriteString("📄 Page Controversy:   " + warningColor.Sprintf("%.1f%% (HIGH)", pageContext.Controversiality*100) + "\n")
		} else {
			output.WriteString("📄 Page Controversy:   " + successColor.Sprintf("%.1f%% (LOW)", pageContext.Controversiality*100) + "\n")
		}

		timing := context.TimingContext
		output.WriteString(fmt.Sprintf("🕐 Edit Hour:          %02d:00\n", timing.EditHour))

		if timing.IsWeekend {
			output.WriteString("📅 Edit Day:           " + infoColor.Sprint("Weekend") + "\n")
		} else {
			output.WriteString("📅 Edit Day:           " + "Weekday" + "\n")
		}

		if timing.TimeSinceLastEdit > 0 {
			if timing.TimeSinceLastEdit < 60 {
				output.WriteString("⏱️  Time Since Last:    " + warningColor.Sprintf("%d minutes (RAPID)", timing.TimeSinceLastEdit) + "\n")
			} else if timing.TimeSinceLastEdit < 1440 { // 24 hours
				output.WriteString("⏱️  Time Since Last:    " + fmt.Sprintf("%d minutes", timing.TimeSinceLastEdit) + "\n")
			} else {
				days := timing.TimeSinceLastEdit / 1440
				output.WriteString("⏱️  Time Since Last:    " + fmt.Sprintf("%d days", days) + "\n")
			}
		}

		if context.ConflictContext.IsContested {
			output.WriteString("⚔️  Conflict Status:    " + warningColor.Sprint("CONTESTED EDIT") + "\n")
			output.WriteString(fmt.Sprintf("📊 Conflict Severity:  %.1f/1.0\n", context.ConflictContext.ConflictSeverity))
		}

		// Related edits
		if len(context.RelatedEdits) > 0 {
			output.WriteString(fmt.Sprintf("🔗 Related Edits:      %d found\n", len(context.RelatedEdits)))

			// Show top 3 related edits
			for i, related := range context.RelatedEdits {
				if i >= 3 {
					break
				}
				output.WriteString(fmt.Sprintf("   • Rev %d by %s (%s, %.2f similarity)\n",
					related.RevisionID, related.Author, related.Relation, related.Similarity))
			}
		}
		output.WriteString("\n")
	}

	// Recommendations
	output.WriteString(headerColor.Sprint("💡 RECOMMENDATIONS\n"))
	output.WriteString(strings.Repeat("─", 50) + "\n")

	if profile.SuspicionScore >= 70 {
		output.WriteString(dangerColor.Sprint("🚨 HIGH RISK CONTRIBUTION\n"))
		output.WriteString("   • Investigate this edit immediately\n")
		output.WriteString("   • Check author's other recent contributions\n")
		output.WriteString("   • Consider reverting if problematic\n")
	} else if profile.SuspicionScore >= 40 {
		output.WriteString(warningColor.Sprint("⚠️ MODERATE RISK CONTRIBUTION\n"))
		output.WriteString("   • Monitor this edit for issues\n")
		output.WriteString("   • Review content for policy compliance\n")
	} else {
		output.WriteString(successColor.Sprint("✅ LOW RISK CONTRIBUTION\n"))
		output.WriteString("   • Edit appears to be constructive\n")
		output.WriteString("   • Continue normal monitoring\n")
	}
	output.WriteString("\n")

	// Footer
	output.WriteString(secondaryColor.Sprint("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n"))
	output.WriteString(secondaryColor.Sprintf("📝 WikiOSINT Contribution Analysis - Revision %d on %s.wikipedia.org\n",
		profile.RevisionID, profile.Language))

	return output.String()
}

// Helper functions for contribution formatting

// formatContributionSuspicionFlag formats contribution suspicion flags into readable text
func formatContributionSuspicionFlag(flag string) string {
	switch flag {
	case "REVERT_EDIT":
		return "This edit is a revert of previous content"
	case "RAPID_EDITING":
		return "Author shows rapid editing patterns"
	case "ANONYMOUS_EDIT":
		return "Edit made by anonymous user"
	case "NEW_ACCOUNT":
		return "Edit made by very new account"
	case "POTENTIAL_BIAS":
		return "Content may contain biased language"
	case "LARGE_ADDITION":
		return "Very large content addition"
	case "LARGE_REMOVAL":
		return "Significant content removal"
	case "BLOCKED_USER":
		return "Edit made by currently blocked user"
	default:
		return flag
	}
}

// formatContentType formats content type into readable text
func formatContentType(contentType string) string {
	switch contentType {
	case "typo_fix":
		return "Typo/spelling correction"
	case "source_addition":
		return "Source/reference addition"
	case "structural_change":
		return "Structural modification"
	case "minor_edit":
		return "Minor edit"
	case "content_edit":
		return "Content modification"
	default:
		return contentType
	}
}

// formatToneAnalysis formats tone analysis into readable text
func formatToneAnalysis(tone string) string {
	switch tone {
	case "neutral":
		return successColor.Sprint("Neutral")
	case "potentially_biased":
		return warningColor.Sprint("Potentially biased")
	case "biased":
		return dangerColor.Sprint("Biased")
	default:
		return tone
	}
}
