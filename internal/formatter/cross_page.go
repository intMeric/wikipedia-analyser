// internal/formatter/cross_page.go
package formatter

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"github.com/intMeric/wikipedia-analyser/internal/models"
	"gopkg.in/yaml.v2"
)

// FormatCrossPageAnalysis formats the cross-page analysis according to the specified format
func FormatCrossPageAnalysis(analysis *models.CrossPageAnalysis, format string) (string, error) {
	switch strings.ToLower(format) {
	case "json":
		return formatCrossPageAsJSON(analysis)
	case "yaml", "yml":
		return formatCrossPageAsYAML(analysis)
	case "table", "":
		return formatCrossPageAsTable(analysis), nil
	default:
		return "", fmt.Errorf("unsupported format: %s (supported: table, json, yaml)", format)
	}
}

// formatCrossPageAsJSON formats cross-page analysis as JSON
func formatCrossPageAsJSON(analysis *models.CrossPageAnalysis) (string, error) {
	data, err := json.MarshalIndent(analysis, "", "  ")
	if err != nil {
		return "", fmt.Errorf("JSON formatting error: %w", err)
	}
	return string(data), nil
}

// formatCrossPageAsYAML formats cross-page analysis as YAML
func formatCrossPageAsYAML(analysis *models.CrossPageAnalysis) (string, error) {
	data, err := yaml.Marshal(analysis)
	if err != nil {
		return "", fmt.Errorf("YAML formatting error: %w", err)
	}
	return string(data), nil
}

// formatCrossPageAsTable formats cross-page analysis as readable table
func formatCrossPageAsTable(analysis *models.CrossPageAnalysis) string {
	var output strings.Builder

	// Header with pages and suspicion score
	output.WriteString(headerColor.Sprint("╭─────────────────────────────────────────────────────────────╮\n"))
	output.WriteString(headerColor.Sprintf("│  🔗 CROSS-PAGE COORDINATION ANALYSIS                      │\n"))
	output.WriteString(headerColor.Sprint("╰─────────────────────────────────────────────────────────────╯\n\n"))

	// Suspicion score with color
	suspicionText := getSuspicionText(analysis.SuspicionScore)
	suspicionColor := getSuspicionColor(analysis.SuspicionScore)
	output.WriteString(fmt.Sprintf("🚨 %s %s (%d/100)\n\n",
		suspicionColor.Sprint("Overall Coordination Score:"),
		suspicionColor.Sprint(suspicionText),
		analysis.SuspicionScore))

	// Analysis overview
	output.WriteString(headerColor.Sprint("📊 ANALYSIS OVERVIEW\n"))
	output.WriteString(strings.Repeat("─", 50) + "\n")

	output.WriteString("📄 Pages Analyzed:     " + strings.Join(analysis.Pages, ", ") + "\n")
	output.WriteString("🌍 Wikipedia Language: " + analysis.Language + "\n")
	output.WriteString("📊 Total Contributors: " + strconv.Itoa(analysis.TotalContributors) + "\n")
	output.WriteString("👥 Common Contributors: " + strconv.Itoa(len(analysis.CommonContributors)) + "\n")
	output.WriteString("🔍 Analysis Timestamp: " + analysis.AnalysisTimestamp.Format("02/01/2006 15:04:05") + "\n")
	output.WriteString("\n")

	// Suspicion flags
	if len(analysis.SuspicionFlags) > 0 {
		output.WriteString(warningColor.Sprint("⚠️  COORDINATION INDICATORS\n"))
		output.WriteString(strings.Repeat("─", 50) + "\n")
		for _, flag := range analysis.SuspicionFlags {
			flagText := formatCrossPageSuspicionFlag(flag)
			output.WriteString(fmt.Sprintf("🔸 %s\n", warningColor.Sprint(flagText)))
		}
		output.WriteString("\n")
	}

	// Mutual support patterns
	if len(analysis.CoordinatedPatterns.MutualSupportPairs) > 0 {
		output.WriteString(headerColor.Sprint("🛡️ MUTUAL SUPPORT PATTERNS\n"))
		output.WriteString(strings.Repeat("─", 80) + "\n")

		for i, pair := range analysis.CoordinatedPatterns.MutualSupportPairs {
			if i >= 10 { // Limit to top 10
				break
			}

			suspicionLevel := pair.SuspicionLevel
			levelColor := getSuspicionLevelColor(suspicionLevel)

			output.WriteString(fmt.Sprintf("👥 %s ↔ %s\n",
				pair.UserA, pair.UserB))
			output.WriteString(fmt.Sprintf("   📊 Support Ratio: %.1f%% | Avg Reaction: %d min | %s\n",
				pair.MutualSupportRatio*100,
				pair.AverageReactionTime,
				levelColor.Sprint(suspicionLevel)))
			output.WriteString(fmt.Sprintf("   🔄 Support Events: %d | Pages: %s\n",
				len(pair.SupportEvents),
				strings.Join(pair.PagesInvolved, ", ")))

			// Show most recent support events
			if len(pair.SupportEvents) > 0 {
				output.WriteString("   📝 Recent Events:\n")
				for j, event := range pair.SupportEvents {
					if j >= 3 { // Show only 3 most recent
						break
					}
					output.WriteString(fmt.Sprintf("      %s: %s defended %s (%s, %dm reaction)\n",
						event.Timestamp.Format("02/01 15:04"),
						event.DefenderUser,
						event.SupportedUser,
						event.SupportType,
						event.ReactionTime))
				}
			}
			output.WriteString("\n")
		}
	}

	// Common contributors analysis
	if len(analysis.CommonContributors) > 0 {
		output.WriteString(headerColor.Sprint("👥 CONTRIBUTORS ACROSS MULTIPLE PAGES\n"))
		output.WriteString(strings.Repeat("─", 80) + "\n")

		for i, contributor := range analysis.CommonContributors {
			if i >= 15 { // Limit to top 15
				break
			}

			username := contributor.Username
			if len(username) > 25 {
				username = username[:25] + "..."
			}

			userType := "👤"
			suspicionDisplay := ""

			if contributor.IsAnonymous {
				userType = "🌐"
				username = secondaryColor.Sprint(username)
				suspicionDisplay = secondaryColor.Sprint("(Anonymous)")
			} else {
				if contributor.SuspicionScore > 0 {
					suspicionText := getSuspicionText(contributor.SuspicionScore)
					suspicionColor := getSuspicionColor(contributor.SuspicionScore)
					suspicionDisplay = fmt.Sprintf("%s (%d/100)",
						suspicionColor.Sprint(suspicionText),
						contributor.SuspicionScore)
				} else {
					suspicionDisplay = secondaryColor.Sprint("(Not analyzed)")
				}
			}

			output.WriteString(fmt.Sprintf("%s %-28s %3d pages | %4d total edits | %s\n",
				userType,
				username,
				len(contributor.PagesEdited),
				contributor.TotalEdits,
				suspicionDisplay))

			// Show page breakdown
			if len(contributor.PagesEdited) > 1 {
				pageDetails := []string{}
				for _, page := range contributor.PagesEdited {
					if edits, exists := contributor.EditsByPage[page]; exists {
						pageDetails = append(pageDetails, fmt.Sprintf("%s (%d)", page, edits))
					}
				}
				if len(pageDetails) > 0 {
					pageDetailsStr := strings.Join(pageDetails, ", ")
					if len(pageDetailsStr) > 70 {
						pageDetailsStr = pageDetailsStr[:70] + "..."
					}
					output.WriteString(fmt.Sprintf("   📋 %s\n", secondaryColor.Sprint(pageDetailsStr)))
				}
			}
		}
		output.WriteString("\n")
	}

	// Coordination score breakdown
	output.WriteString(headerColor.Sprint("📈 COORDINATION METRICS\n"))
	output.WriteString(strings.Repeat("─", 50) + "\n")

	output.WriteString(fmt.Sprintf("🤝 Coordination Score:    %.1f/100\n", analysis.CoordinatedPatterns.CoordinationScore))
	output.WriteString(fmt.Sprintf("🛡️  Mutual Support Pairs:  %d\n", len(analysis.CoordinatedPatterns.MutualSupportPairs)))
	output.WriteString(fmt.Sprintf("🔄 Tag Team Patterns:     %d\n", len(analysis.CoordinatedPatterns.TagTeamEditing)))
	output.WriteString(fmt.Sprintf("⚔️  Coordinated Reverts:   %d\n", len(analysis.CoordinatedPatterns.CoordinatedReversions)))
	output.WriteString(fmt.Sprintf("🕸️  Support Networks:      %d\n", len(analysis.CoordinatedPatterns.SupportNetworks)))
	output.WriteString(fmt.Sprintf("🎭 Sockpuppet Networks:   %d\n", len(analysis.SockpuppetNetworks)))
	output.WriteString("\n")

	// Page-by-page summary
	if len(analysis.PageProfiles) > 0 {
		output.WriteString(headerColor.Sprint("📄 PAGE-BY-PAGE SUMMARY\n"))
		output.WriteString(strings.Repeat("─", 80) + "\n")

		for _, pageName := range analysis.Pages {
			if profile, exists := analysis.PageProfiles[pageName]; exists {
				pageTitle := pageName
				if len(pageTitle) > 40 {
					pageTitle = pageTitle[:40] + "..."
				}

				suspicionText := getSuspicionText(profile.SuspicionScore)
				suspicionColor := getSuspicionColor(profile.SuspicionScore)

				output.WriteString(fmt.Sprintf("📄 %-43s %s (%d/100)\n",
					pageTitle,
					suspicionColor.Sprint(suspicionText),
					profile.SuspicionScore))
				output.WriteString(fmt.Sprintf("   📊 %d revisions | %d contributors | %.1f%% conflict rate\n",
					profile.TotalRevisions,
					len(profile.Contributors),
					profile.ConflictStats.ControversyScore*100))
			}
		}
		output.WriteString("\n")
	}

	// Recommendations
	output.WriteString(headerColor.Sprint("💡 ANALYSIS RECOMMENDATIONS\n"))
	output.WriteString(strings.Repeat("─", 50) + "\n")

	if analysis.SuspicionScore >= 70 {
		output.WriteString(dangerColor.Sprint("🚨 HIGH COORDINATION DETECTED\n"))
		output.WriteString("   • Investigate identified user pairs for sockpuppetry\n")
		output.WriteString("   • Review coordination timeline for organized campaigns\n")
		output.WriteString("   • Consider checking additional related pages\n")
	} else if analysis.SuspicionScore >= 40 {
		output.WriteString(warningColor.Sprint("⚠️ MODERATE COORDINATION DETECTED\n"))
		output.WriteString("   • Monitor identified patterns for escalation\n")
		output.WriteString("   • Review user behavior for policy violations\n")
	} else {
		output.WriteString(successColor.Sprint("✅ LOW COORDINATION RISK\n"))
		output.WriteString("   • Normal collaborative editing patterns observed\n")
		output.WriteString("   • Continue standard monitoring procedures\n")
	}
	output.WriteString("\n")

	// Footer
	output.WriteString(secondaryColor.Sprint("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n"))
	output.WriteString(secondaryColor.Sprintf("🔗 WikiOSINT Cross-Page Analysis - %d pages analyzed on %s.wikipedia.org\n",
		analysis.TotalPages, analysis.Language))

	return output.String()
}

// Helper functions for cross-page formatting

// formatCrossPageSuspicionFlag formats cross-page suspicion flags into readable text
func formatCrossPageSuspicionFlag(flag string) string {
	switch flag {
	case "MUTUAL_SUPPORT_DETECTED":
		return "Mutual support patterns detected between users"
	case "HIGH_COORDINATION_SCORE":
		return "High overall coordination score"
	case "SOCKPUPPET_NETWORK_DETECTED":
		return "Potential sockpuppet network identified"
	case "HIGH_CONTRIBUTOR_OVERLAP":
		return "High overlap of contributors across pages"
	case "TEMPORAL_SYNCHRONIZATION":
		return "Synchronized editing patterns detected"
	case "TAG_TEAM_EDITING":
		return "Tag-team editing strategies observed"
	case "COORDINATED_REVERSIONS":
		return "Coordinated reversion campaigns detected"
	default:
		return flag
	}
}

// getSuspicionLevelColor returns appropriate color for suspicion level
func getSuspicionLevelColor(level string) *color.Color {
	switch level {
	case "VERY_HIGH":
		return dangerColor
	case "HIGH":
		return color.New(color.FgRed)
	case "MODERATE":
		return warningColor
	case "LOW":
		return color.New(color.FgYellow)
	case "NONE":
		return successColor
	default:
		return secondaryColor
	}
}
