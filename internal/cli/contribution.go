// internal/cli/contribution.go
package cli

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/intMeric/wikipedia-analyser/internal/analyzer"
	"github.com/intMeric/wikipedia-analyser/internal/client"
	"github.com/intMeric/wikipedia-analyser/internal/formatter"
	"github.com/intMeric/wikipedia-analyser/internal/models"
	"github.com/spf13/cobra"
)

var (
	contributionOutputFormat   string
	contributionLanguage       string
	contributionSaveToFile     string
	contributionAnalysisDepth  string
	contributionIncludeContent bool
	contributionIncludeContext bool
)

// contributionCmd represents the contribution command
var contributionCmd = &cobra.Command{
	Use:   "contribution",
	Short: "Wikipedia contribution analysis",
	Long: `Commands to analyze individual Wikipedia contributions/revisions and detect
suspicious editing patterns, content quality issues, and potential violations.`,
}

// analyzeContributionCmd represents the contribution analyze command
var analyzeContributionCmd = &cobra.Command{
	Use:   "analyze [revision_id] [page_title]",
	Short: "Analyze a specific Wikipedia contribution",
	Long: `Comprehensive analysis of a specific Wikipedia contribution including:
- Author analysis and behavior patterns
- Content quality and compliance assessment
- Edit context and timing analysis
- Conflict detection and revert analysis
- Risk assessment for various violations
- Detailed suspicion scoring

You can specify either:
  - Just revision ID: analyze [revision_id]
  - Revision ID and page: analyze [revision_id] [page_title]
  - Page and find latest: analyze latest [page_title]

Configuration options:
  --depth: Analysis depth (basic, standard, deep) - default: standard
  --include-content: Include detailed content analysis - default: true
  --include-context: Include contextual analysis - default: false (only for deep)`,
	Args: cobra.RangeArgs(1, 2),
	RunE: runContributionAnalyze,
}

// recentContributionsCmd represents the contribution recent command
var recentContributionsCmd = &cobra.Command{
	Use:   "recent [page_title]",
	Short: "Analyze recent contributions to a page",
	Long: `Analyze the most recent contributions to a Wikipedia page including:
- Last 10-20 contributions analysis
- Pattern detection across recent edits
- Contributor behavior comparison
- Quality trends over time

Configuration options:
  --depth: Analysis depth (basic, standard) - default: basic
  --limit: Number of recent contributions to analyze (5-50) - default: 10`,
	Args: cobra.ExactArgs(1),
	RunE: runRecentContributions,
}

// suspiciousContributionsCmd represents the contribution suspicious command
var suspiciousContributionsCmd = &cobra.Command{
	Use:   "suspicious [page_title]",
	Short: "Find suspicious contributions to a page",
	Long: `Scan recent contributions to identify potentially suspicious edits including:
- High suspicion score contributions
- Anomalous editing patterns
- Potential policy violations
- Edit war participation
- Sockpuppet activity indicators

Configuration options:
  --threshold: Minimum suspicion score threshold (0-100) - default: 40
  --days: Number of days to scan back - default: 30
  --limit: Maximum suspicious contributions to show - default: 20`,
	Args: cobra.ExactArgs(1),
	RunE: runSuspiciousContributions,
}

func init() {
	// Add subcommands
	contributionCmd.AddCommand(analyzeContributionCmd)
	contributionCmd.AddCommand(recentContributionsCmd)
	contributionCmd.AddCommand(suspiciousContributionsCmd)

	// Flags for analyze command
	analyzeContributionCmd.Flags().StringVarP(&contributionOutputFormat, "output", "o", "table", "output format (table, json, yaml)")
	analyzeContributionCmd.Flags().StringVarP(&contributionLanguage, "lang", "l", "en", "Wikipedia language (en, fr, de, etc.)")
	analyzeContributionCmd.Flags().StringVar(&contributionSaveToFile, "save", "", "save result to file")
	analyzeContributionCmd.Flags().StringVar(&contributionAnalysisDepth, "depth", "standard", "analysis depth (basic, standard, deep)")
	analyzeContributionCmd.Flags().BoolVar(&contributionIncludeContent, "include-content", true, "include detailed content analysis")
	analyzeContributionCmd.Flags().BoolVar(&contributionIncludeContext, "include-context", false, "include contextual analysis (auto-enabled for deep)")

	// Flags for recent command
	recentContributionsCmd.Flags().StringVarP(&contributionOutputFormat, "output", "o", "table", "output format (table, json, yaml)")
	recentContributionsCmd.Flags().StringVarP(&contributionLanguage, "lang", "l", "en", "Wikipedia language (en, fr, de, etc.)")
	recentContributionsCmd.Flags().StringVar(&contributionSaveToFile, "save", "", "save result to file")
	recentContributionsCmd.Flags().StringVar(&contributionAnalysisDepth, "depth", "basic", "analysis depth (basic, standard)")
	recentContributionsCmd.Flags().IntVar(&recentLimit, "limit", 10, "number of recent contributions to analyze (5-50)")

	// Flags for suspicious command
	suspiciousContributionsCmd.Flags().StringVarP(&contributionOutputFormat, "output", "o", "table", "output format (table, json, yaml)")
	suspiciousContributionsCmd.Flags().StringVarP(&contributionLanguage, "lang", "l", "en", "Wikipedia language (en, fr, de, etc.)")
	suspiciousContributionsCmd.Flags().StringVar(&contributionSaveToFile, "save", "", "save result to file")
	suspiciousContributionsCmd.Flags().IntVar(&suspicionThreshold, "threshold", 40, "minimum suspicion score threshold (0-100)")
	suspiciousContributionsCmd.Flags().IntVar(&scanDays, "days", 30, "number of days to scan back")
	suspiciousContributionsCmd.Flags().IntVar(&suspiciousLimit, "limit", 20, "maximum suspicious contributions to show")
}

var (
	recentLimit        int = 10
	suspicionThreshold int = 40
	scanDays           int = 30
	suspiciousLimit    int = 20
)

func runContributionAnalyze(cmd *cobra.Command, args []string) error {
	// Parse arguments
	var revisionID int
	var pageTitle string
	var err error

	if strings.ToLower(args[0]) == "latest" {
		// Special case: analyze latest revision of a page
		if len(args) != 2 {
			return fmt.Errorf("when using 'latest', you must specify a page title")
		}
		pageTitle = args[1]
		revisionID = 0 // Will be resolved to latest revision
	} else {
		// Parse revision ID
		revisionID, err = strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid revision ID: %s", args[0])
		}

		// Page title is optional
		if len(args) > 1 {
			pageTitle = args[1]
		}
	}

	// Validate analysis depth
	if contributionAnalysisDepth != "basic" && contributionAnalysisDepth != "standard" && contributionAnalysisDepth != "deep" {
		return fmt.Errorf("invalid analysis depth: %s (must be: basic, standard, deep)", contributionAnalysisDepth)
	}

	// Auto-enable context analysis for deep analysis
	if contributionAnalysisDepth == "deep" {
		contributionIncludeContext = true
	}

	// Create Wikipedia client
	wikiClient := client.NewWikipediaClient(contributionLanguage)

	// Create contribution analysis options
	analysisOptions := analyzer.ContributionAnalysisOptions{
		AnalysisDepth:  contributionAnalysisDepth,
		IncludeContent: contributionIncludeContent,
		IncludeContext: contributionIncludeContext,
	}

	// Create contribution analyzer with options
	contributionAnalyzer := analyzer.NewContributionAnalyzer(wikiClient, analysisOptions)

	// Display analysis start info
	if revisionID == 0 {
		fmt.Printf("🔍 Analyzing latest contribution to: %s\n", pageTitle)
	} else if pageTitle != "" {
		fmt.Printf("🔍 Analyzing contribution: Revision %d on %s\n", revisionID, pageTitle)
	} else {
		fmt.Printf("🔍 Analyzing contribution: Revision %d\n", revisionID)
	}
	fmt.Printf("📡 Fetching data from %s.wikipedia.org...\n", contributionLanguage)
	fmt.Printf("📊 Analysis depth: %s\n", contributionAnalysisDepth)
	if contributionIncludeContent {
		fmt.Printf("📝 Including detailed content analysis...\n")
	}
	if contributionIncludeContext {
		fmt.Printf("🔍 Including contextual analysis...\n")
	}

	// Retrieve and analyze contribution
	contributionProfile, err := contributionAnalyzer.GetContributionProfile(revisionID, pageTitle)
	if err != nil {
		return fmt.Errorf("error retrieving contribution profile: %w", err)
	}

	fmt.Printf("✅ Analysis completed! Revision %d analyzed\n", contributionProfile.RevisionID)
	if contributionProfile.SuspicionScore > 50 {
		fmt.Printf("⚠️  High suspicion score detected: %d/100\n", contributionProfile.SuspicionScore)
	}

	// Format and display results
	output, err := formatter.FormatContributionProfile(contributionProfile, contributionOutputFormat)
	if err != nil {
		return fmt.Errorf("error formatting output: %w", err)
	}

	// Display or save
	if contributionSaveToFile != "" {
		err = os.WriteFile(contributionSaveToFile, []byte(output), 0644)
		if err != nil {
			return fmt.Errorf("error saving file: %w", err)
		}
		fmt.Printf("✅ Results saved to: %s\n", contributionSaveToFile)
	} else {
		fmt.Print(output)
	}

	return nil
}

func runRecentContributions(cmd *cobra.Command, args []string) error {
	pageTitle := args[0]

	// Validate limit
	if recentLimit < 5 || recentLimit > 50 {
		return fmt.Errorf("limit must be between 5 and 50")
	}

	// Validate analysis depth for recent analysis
	if contributionAnalysisDepth != "basic" && contributionAnalysisDepth != "standard" {
		return fmt.Errorf("for recent contributions analysis, depth must be 'basic' or 'standard'")
	}

	// Create Wikipedia client
	wikiClient := client.NewWikipediaClient(contributionLanguage)

	// Create analysis options
	analysisOptions := analyzer.ContributionAnalysisOptions{
		AnalysisDepth:  contributionAnalysisDepth,
		IncludeContent: contributionAnalysisDepth == "standard",
		IncludeContext: false, // Too expensive for bulk analysis
	}

	contributionAnalyzer := analyzer.NewContributionAnalyzer(wikiClient, analysisOptions)

	fmt.Printf("🔍 Analyzing %d recent contributions to: %s\n", recentLimit, pageTitle)
	fmt.Printf("📡 Fetching data from %s.wikipedia.org...\n", contributionLanguage)
	fmt.Printf("📊 Analysis depth: %s\n", contributionAnalysisDepth)

	// Get recent revisions
	revisions, err := wikiClient.GetPageRevisions(pageTitle, recentLimit)
	if err != nil {
		return fmt.Errorf("error retrieving page revisions: %w", err)
	}

	if len(revisions) == 0 {
		fmt.Printf("❌ No revisions found for page: %s\n", pageTitle)
		return nil
	}

	fmt.Printf("📊 Found %d recent revisions, analyzing...\n", len(revisions))

	// Analyze each revision
	var results []string
	suspiciousCount := 0

	for i, revision := range revisions {
		fmt.Printf("📝 Analyzing revision %d/%d (ID: %d)...\n", i+1, len(revisions), revision.RevID)

		profile, err := contributionAnalyzer.GetContributionProfile(revision.RevID, pageTitle)
		if err != nil {
			fmt.Printf("⚠️  Failed to analyze revision %d: %v\n", revision.RevID, err)
			continue
		}

		if profile.SuspicionScore >= 30 {
			suspiciousCount++
		}

		// Format individual result
		output, err := formatter.FormatContributionProfile(profile, contributionOutputFormat)
		if err != nil {
			fmt.Printf("⚠️  Failed to format revision %d: %v\n", revision.RevID, err)
			continue
		}

		results = append(results, output)
		results = append(results, "\n"+strings.Repeat("═", 80)+"\n\n")
	}

	fmt.Printf("✅ Analysis completed! %d revisions analyzed\n", len(results)/2)
	if suspiciousCount > 0 {
		fmt.Printf("⚠️  Found %d contributions with elevated suspicion scores\n", suspiciousCount)
	}

	// Combine all results
	finalOutput := strings.Join(results, "")

	// Display or save
	if contributionSaveToFile != "" {
		err = os.WriteFile(contributionSaveToFile, []byte(finalOutput), 0644)
		if err != nil {
			return fmt.Errorf("error saving file: %w", err)
		}
		fmt.Printf("✅ Results saved to: %s\n", contributionSaveToFile)
	} else {
		fmt.Print(finalOutput)
	}

	return nil
}

func runSuspiciousContributions(cmd *cobra.Command, args []string) error {
	pageTitle := args[0]

	// Validate parameters
	if suspicionThreshold < 0 || suspicionThreshold > 100 {
		return fmt.Errorf("suspicion threshold must be between 0 and 100")
	}
	if scanDays < 1 || scanDays > 365 {
		return fmt.Errorf("scan days must be between 1 and 365")
	}
	if suspiciousLimit < 1 || suspiciousLimit > 100 {
		return fmt.Errorf("suspicious limit must be between 1 and 100")
	}

	// Create Wikipedia client
	wikiClient := client.NewWikipediaClient(contributionLanguage)

	// Create analysis options (use basic for bulk scanning)
	analysisOptions := analyzer.ContributionAnalysisOptions{
		AnalysisDepth:  "basic",
		IncludeContent: false,
		IncludeContext: false,
	}

	contributionAnalyzer := analyzer.NewContributionAnalyzer(wikiClient, analysisOptions)

	fmt.Printf("🔍 Scanning for suspicious contributions to: %s\n", pageTitle)
	fmt.Printf("📊 Threshold: %d/100, Scanning: %d days back\n", suspicionThreshold, scanDays)
	fmt.Printf("📡 Fetching data from %s.wikipedia.org...\n", contributionLanguage)

	// Get page history for the specified time period
	history, err := wikiClient.GetPageHistory(pageTitle, scanDays)
	if err != nil {
		return fmt.Errorf("error retrieving page history: %w", err)
	}

	if len(history) == 0 {
		fmt.Printf("❌ No revisions found in the last %d days for page: %s\n", scanDays, pageTitle)
		return nil
	}

	fmt.Printf("📊 Found %d revisions in the last %d days, scanning for suspicious activity...\n", len(history), scanDays)

	// Scan and analyze suspicious revisions
	var suspiciousProfiles []*models.ContributionProfile
	scannedCount := 0

	for _, revision := range history {
		scannedCount++
		if scannedCount%10 == 0 {
			fmt.Printf("📝 Scanned %d/%d revisions...\n", scannedCount, len(history))
		}

		// Quick analysis to get suspicion score
		profile, err := contributionAnalyzer.GetContributionProfile(revision.RevID, pageTitle)
		if err != nil {
			continue // Skip failed analyses
		}

		// Check if meets suspicion threshold
		if profile.SuspicionScore >= suspicionThreshold {
			suspiciousProfiles = append(suspiciousProfiles, profile)

			// Stop if we've found enough suspicious contributions
			if len(suspiciousProfiles) >= suspiciousLimit {
				break
			}
		}
	}

	fmt.Printf("✅ Scan completed! Found %d suspicious contributions\n", len(suspiciousProfiles))

	if len(suspiciousProfiles) == 0 {
		fmt.Printf("🎉 No suspicious contributions found with threshold %d/100\n", suspicionThreshold)
		return nil
	}

	// Sort by suspicion score (highest first)
	// Note: In a real implementation, you'd want to sort the slice
	// For now, they're already in chronological order

	// Format results
	var results []string
	results = append(results, fmt.Sprintf("🚨 SUSPICIOUS CONTRIBUTIONS REPORT\n"))
	results = append(results, fmt.Sprintf("Page: %s | Threshold: %d/100 | Found: %d contributions\n\n", pageTitle, suspicionThreshold, len(suspiciousProfiles)))

	for i, profile := range suspiciousProfiles {
		results = append(results, fmt.Sprintf("=== SUSPICIOUS CONTRIBUTION #%d ===\n", i+1))

		output, err := formatter.FormatContributionProfile(profile, contributionOutputFormat)
		if err != nil {
			results = append(results, fmt.Sprintf("Error formatting contribution %d: %v\n", profile.RevisionID, err))
			continue
		}

		results = append(results, output)
		results = append(results, "\n"+strings.Repeat("═", 80)+"\n\n")
	}

	finalOutput := strings.Join(results, "")

	// Display or save
	if contributionSaveToFile != "" {
		err = os.WriteFile(contributionSaveToFile, []byte(finalOutput), 0644)
		if err != nil {
			return fmt.Errorf("error saving file: %w", err)
		}
		fmt.Printf("✅ Suspicious contributions report saved to: %s\n", contributionSaveToFile)
	} else {
		fmt.Print(finalOutput)
	}

	return nil
}
