// internal/cli/page.go
package cli

import (
	"fmt"
	"os"
	"wikianalyser/internal/analyzer"
	"wikianalyser/internal/client"
	"wikianalyser/internal/formatter"

	"github.com/spf13/cobra"
)

var (
	pageOutputFormat string
	pageLanguage     string
	pageSaveToFile   string
	pageAnalyzeDays  int
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
- Recent activity analysis`,
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
- Contributor activity timeline`,
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
- Controversy scoring`,
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

	// Flags for history command
	historyCmd.Flags().StringVarP(&pageOutputFormat, "output", "o", "table", "output format (table, json, yaml)")
	historyCmd.Flags().StringVarP(&pageLanguage, "lang", "l", "en", "Wikipedia language (en, fr, de, etc.)")
	historyCmd.Flags().StringVar(&pageSaveToFile, "save", "", "save result to file")
	historyCmd.Flags().IntVar(&pageAnalyzeDays, "days", 30, "number of days to analyze")

	// Flags for conflicts command
	conflictsCmd.Flags().StringVarP(&pageOutputFormat, "output", "o", "table", "output format (table, json, yaml)")
	conflictsCmd.Flags().StringVarP(&pageLanguage, "lang", "l", "en", "Wikipedia language (en, fr, de, etc.)")
	conflictsCmd.Flags().StringVar(&pageSaveToFile, "save", "", "save result to file")
	conflictsCmd.Flags().IntVar(&pageAnalyzeDays, "days", 30, "number of days to analyze")
}

func runPageAnalyze(cmd *cobra.Command, args []string) error {
	pageTitle := args[0]

	// Create Wikipedia client
	wikiClient := client.NewWikipediaClient(pageLanguage)

	// Create page analyzer
	pageAnalyzer := analyzer.NewPageAnalyzer(wikiClient)

	// Retrieve page data
	fmt.Printf("🔍 Analyzing Wikipedia page: %s\n", pageTitle)
	fmt.Printf("📡 Fetching data from %s.wikipedia.org...\n", pageLanguage)
	fmt.Printf("📅 Analyzing last %d days of activity...\n", pageAnalyzeDays)

	pageProfile, err := pageAnalyzer.GetPageProfile(pageTitle)
	if err != nil {
		return fmt.Errorf("error retrieving page profile: %w", err)
	}

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

	// Create page analyzer
	pageAnalyzer := analyzer.NewPageAnalyzer(wikiClient)

	// Retrieve page data with focus on history
	fmt.Printf("🔍 Analyzing edit history for: %s\n", pageTitle)
	fmt.Printf("📡 Fetching revision data from %s.wikipedia.org...\n", pageLanguage)
	fmt.Printf("📅 Analyzing last %d days...\n", pageAnalyzeDays)

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

	// Create page analyzer
	pageAnalyzer := analyzer.NewPageAnalyzer(wikiClient)

	// Retrieve page data with focus on conflicts
	fmt.Printf("🔍 Analyzing conflicts for: %s\n", pageTitle)
	fmt.Printf("📡 Detecting edit wars on %s.wikipedia.org...\n", pageLanguage)
	fmt.Printf("📅 Analyzing last %d days for conflicts...\n", pageAnalyzeDays)

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
