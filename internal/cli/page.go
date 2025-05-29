// internal/cli/page.go
package cli

import (
	"fmt"
	"os"

	"github.com/intMeric/wikipedia-analyser/internal/analyzer"
	"github.com/intMeric/wikipedia-analyser/internal/client"
	"github.com/intMeric/wikipedia-analyser/internal/formatter"
	"github.com/spf13/cobra"
)

var (
	pageOutputFormat    string
	pageLanguage        string
	pageSaveToFile      string
	pageAnalyzeDays     int
	pageMaxRevisions    int
	pageMaxContributors int
	pageMaxHistory      int
)

// pageCmd represents the page command
var pageCmd = &cobra.Command{
	Use:   "page",
	Short: "Wikipedia page analysis",
	Long: `Commands to analyze Wikipedia pages and detect
manipulations, edit wars, and suspicious activity.`,
}

// analyzeCmd represents the page analyze command
var analyzeCmd = &cobra.Command{
	Use:   "analyze [page_title]",
	Short: "Analyze a Wikipedia page",
	Long: `Comprehensive analysis of a Wikipedia page including:
- Edit history and revision patterns
- Top contributors and their behavior
- Conflict detection and edit wars
- Quality metrics and suspicion scoring
- Recent activity analysis

Configuration options:
  --max-revisions: Number of revisions to analyze (default: 100)
  --max-contributors: Number of contributors to analyze (default: 20)
  --max-history: Days of detailed history to analyze (default: 30)`,
	Args: cobra.ExactArgs(1),
	RunE: runPageAnalyze,
}

// historyCmd represents the page history command
var historyCmd = &cobra.Command{
	Use:   "history [page_title]",
	Short: "Analyze page edit history",
	Long: `Detailed analysis of page edit history including:
- Recent revisions and patterns
- Edit frequency over time
- Size changes and content evolution
- Contributor activity timeline

Configuration options:
  --max-revisions: Number of revisions to analyze (default: 100)
  --max-history: Days of detailed history to analyze (default: 30)`,
	Args: cobra.ExactArgs(1),
	RunE: runPageHistory,
}

// conflictsCmd represents the page conflicts command
var conflictsCmd = &cobra.Command{
	Use:   "conflicts [page_title]",
	Short: "Detect edit wars and conflicts",
	Long: `Analyze page for edit wars and conflicts including:
- Reversion detection
- Conflicting users identification
- Edit war periods
- Controversy scoring

Configuration options:
  --max-revisions: Number of revisions to analyze (default: 100)
  --max-history: Days of detailed history to analyze (default: 30)`,
	Args: cobra.ExactArgs(1),
	RunE: runPageConflicts,
}

func init() {
	// Add subcommands
	pageCmd.AddCommand(analyzeCmd)
	pageCmd.AddCommand(historyCmd)
	pageCmd.AddCommand(conflictsCmd)

	// Flags for analyze command
	analyzeCmd.Flags().StringVarP(&pageOutputFormat, "output", "o", "table", "output format (table, json, yaml)")
	analyzeCmd.Flags().StringVarP(&pageLanguage, "lang", "l", "en", "Wikipedia language (en, fr, de, etc.)")
	analyzeCmd.Flags().StringVar(&pageSaveToFile, "save", "", "save result to file")
	analyzeCmd.Flags().IntVar(&pageAnalyzeDays, "days", 30, "number of days to analyze")
	analyzeCmd.Flags().IntVar(&pageMaxRevisions, "max-revisions", 100, "maximum number of revisions to analyze")
	analyzeCmd.Flags().IntVar(&pageMaxContributors, "max-contributors", 20, "maximum number of contributors to analyze")
	analyzeCmd.Flags().IntVar(&pageMaxHistory, "max-history", 30, "maximum number of days for detailed history")

	// Flags for history command
	historyCmd.Flags().StringVarP(&pageOutputFormat, "output", "o", "table", "output format (table, json, yaml)")
	historyCmd.Flags().StringVarP(&pageLanguage, "lang", "l", "en", "Wikipedia language (en, fr, de, etc.)")
	historyCmd.Flags().StringVar(&pageSaveToFile, "save", "", "save result to file")
	historyCmd.Flags().IntVar(&pageAnalyzeDays, "days", 30, "number of days to analyze")
	historyCmd.Flags().IntVar(&pageMaxRevisions, "max-revisions", 100, "maximum number of revisions to analyze")
	historyCmd.Flags().IntVar(&pageMaxContributors, "max-contributors", 20, "maximum number of contributors to analyze")
	historyCmd.Flags().IntVar(&pageMaxHistory, "max-history", 30, "maximum number of days for detailed history")

	// Flags for conflicts command
	conflictsCmd.Flags().StringVarP(&pageOutputFormat, "output", "o", "table", "output format (table, json, yaml)")
	conflictsCmd.Flags().StringVarP(&pageLanguage, "lang", "l", "en", "Wikipedia language (en, fr, de, etc.)")
	conflictsCmd.Flags().StringVar(&pageSaveToFile, "save", "", "save result to file")
	conflictsCmd.Flags().IntVar(&pageAnalyzeDays, "days", 30, "number of days to analyze")
	conflictsCmd.Flags().IntVar(&pageMaxRevisions, "max-revisions", 100, "maximum number of revisions to analyze")
	conflictsCmd.Flags().IntVar(&pageMaxContributors, "max-contributors", 20, "maximum number of contributors to analyze")
	conflictsCmd.Flags().IntVar(&pageMaxHistory, "max-history", 30, "maximum number of days for detailed history")
}

func runPageAnalyze(cmd *cobra.Command, args []string) error {
	pageTitle := args[0]

	// Create Wikipedia client
	wikiClient := client.NewWikipediaClient(pageLanguage)

	// Create page analysis options
	analysisOptions := analyzer.PageAnalysisOptions{
		NumberOfPageRevisions: pageMaxRevisions,
		NumberOfDaysHistory:   pageMaxHistory,
		NumberOfContributors:  pageMaxContributors,
	}

	// Create page analyzer with options
	pageAnalyzer := analyzer.NewPageAnalyzer(wikiClient, analysisOptions)

	// Retrieve page data
	fmt.Printf("🔍 Analyzing Wikipedia page: %s\n", pageTitle)
	fmt.Printf("📡 Fetching data from %s.wikipedia.org...\n", pageLanguage)
	fmt.Printf("📊 Analysis parameters: %d revisions, %d contributors, %d days history\n",
		pageMaxRevisions, pageMaxContributors, pageMaxHistory)
	fmt.Printf("👥 Including detailed contributor analysis...\n")

	pageProfile, err := pageAnalyzer.GetPageProfile(pageTitle)
	if err != nil {
		return fmt.Errorf("error retrieving page profile: %w", err)
	}

	fmt.Printf("✅ Analysis completed! Found %d contributors, %d revisions\n",
		len(pageProfile.Contributors), len(pageProfile.RecentRevisions))

	// Format and display results
	output, err := formatter.FormatPageProfile(pageProfile, pageOutputFormat)
	if err != nil {
		return fmt.Errorf("error formatting output: %w", err)
	}

	// Display or save
	if pageSaveToFile != "" {
		err = os.WriteFile(pageSaveToFile, []byte(output), 0644)
		if err != nil {
			return fmt.Errorf("error saving file: %w", err)
		}
		fmt.Printf("✅ Results saved to: %s\n", pageSaveToFile)
	} else {
		fmt.Print(output)
	}

	return nil
}

func runPageHistory(cmd *cobra.Command, args []string) error {
	pageTitle := args[0]

	// Create Wikipedia client
	wikiClient := client.NewWikipediaClient(pageLanguage)

	// Create page analysis options
	analysisOptions := analyzer.PageAnalysisOptions{
		NumberOfPageRevisions: pageMaxRevisions,
		NumberOfDaysHistory:   pageMaxHistory,
		NumberOfContributors:  pageMaxContributors,
	}

	// Create page analyzer with options
	pageAnalyzer := analyzer.NewPageAnalyzer(wikiClient, analysisOptions)

	// Retrieve page data with focus on history
	fmt.Printf("🔍 Analyzing edit history for: %s\n", pageTitle)
	fmt.Printf("📡 Fetching revision data from %s.wikipedia.org...\n", pageLanguage)
	fmt.Printf("📊 Analysis parameters: %d revisions, %d days history\n",
		pageMaxRevisions, pageMaxHistory)

	pageProfile, err := pageAnalyzer.GetPageProfile(pageTitle)
	if err != nil {
		return fmt.Errorf("error retrieving page profile: %w", err)
	}

	// Format with focus on history (could be a separate formatter method)
	output, err := formatter.FormatPageHistory(pageProfile, pageOutputFormat)
	if err != nil {
		return fmt.Errorf("error formatting output: %w", err)
	}

	// Display or save
	if pageSaveToFile != "" {
		err = os.WriteFile(pageSaveToFile, []byte(output), 0644)
		if err != nil {
			return fmt.Errorf("error saving file: %w", err)
		}
		fmt.Printf("✅ Results saved to: %s\n", pageSaveToFile)
	} else {
		fmt.Print(output)
	}

	return nil
}

func runPageConflicts(cmd *cobra.Command, args []string) error {
	pageTitle := args[0]

	// Create Wikipedia client
	wikiClient := client.NewWikipediaClient(pageLanguage)

	// Create page analysis options
	analysisOptions := analyzer.PageAnalysisOptions{
		NumberOfPageRevisions: pageMaxRevisions,
		NumberOfDaysHistory:   pageMaxHistory,
		NumberOfContributors:  pageMaxContributors,
	}

	// Create page analyzer with options
	pageAnalyzer := analyzer.NewPageAnalyzer(wikiClient, analysisOptions)

	// Retrieve page data with focus on conflicts
	fmt.Printf("🔍 Analyzing conflicts for: %s\n", pageTitle)
	fmt.Printf("📡 Detecting edit wars on %s.wikipedia.org...\n", pageLanguage)
	fmt.Printf("📊 Analysis parameters: %d revisions, %d days for conflict detection\n",
		pageMaxRevisions, pageMaxHistory)

	pageProfile, err := pageAnalyzer.GetPageProfile(pageTitle)
	if err != nil {
		return fmt.Errorf("error retrieving page profile: %w", err)
	}

	// Format with focus on conflicts (could be a separate formatter method)
	output, err := formatter.FormatPageConflicts(pageProfile, pageOutputFormat)
	if err != nil {
		return fmt.Errorf("error formatting output: %w", err)
	}

	// Display or save
	if pageSaveToFile != "" {
		err = os.WriteFile(pageSaveToFile, []byte(output), 0644)
		if err != nil {
			return fmt.Errorf("error saving file: %w", err)
		}
		fmt.Printf("✅ Results saved to: %s\n", pageSaveToFile)
	} else {
		fmt.Print(output)
	}

	return nil
}
