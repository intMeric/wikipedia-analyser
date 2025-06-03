// internal/formatter/common.go
package formatter

import (
	"github.com/fatih/color"
)

var (
	// Colors for terminal display - shared across all formatters
	headerColor    = color.New(color.FgCyan, color.Bold)
	successColor   = color.New(color.FgGreen)
	warningColor   = color.New(color.FgYellow)
	dangerColor    = color.New(color.FgRed, color.Bold)
	infoColor      = color.New(color.FgBlue)
	secondaryColor = color.New(color.FgHiBlack)
)

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

// truncateString truncates a string to specified length
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
