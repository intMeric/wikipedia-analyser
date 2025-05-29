// internal/cli/pages.go
package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/intMeric/wikipedia-analyser/internal/analyzer"
	"github.com/intMeric/wikipedia-analyser/internal/client"
	"github.com/intMeric/wikipedia-analyser/internal/formatter"
	"github.com/intMeric/wikipedia-analyser/internal/models"
	"github.com/spf13/cobra"
)

var (
	// Cross-page analysis specific flags
	pagesOutputFormat           string
	pagesLanguage               string
	pagesSaveToFile             string
	pagesMaxRevisions           int
	pagesMaxContributors        int
	pagesMaxHistory             int
	crossPageMinCommonEdits     int
	crossPageMaxReactionTime    int
	crossPageMinSupportRatio    float64
	crossPageEnableDeepAnalysis bool
)

// pagesCmd represents the cross-page analysis command
var pagesCmd = &cobra.Command{
	Use:   "pages [page1] [page2] [page3...]",
	Short: "Cross-page analysis to detect coordination patterns",
	Long: `Analyze multiple Wikipedia pages together to detect:
- Coordinated editing campaigns
- Mutual support patterns between users
- Sockpuppet networks
- Temporal synchronization
- Tag-team editing strategies

This command helps identify sophisticated manipulation campaigns
that operate across multiple pages simultaneously.

Configuration options:
  --max-revisions: Number of revisions per page (default: 200)
  --max-contributors: Number of contributors per page (default: 50)
  --max-history: Days of detailed history (default: 90)
  --min-common-edits: Minimum edits to be considered common contributor (default: 3)
  --max-reaction-time: Maximum minutes for suspicious reaction time (default: 60)
  --min-support-ratio: Minimum ratio for mutual support detection (default: 0.3)
  --enable-deep-analysis: Enable resource-intensive analysis (default: false)

Examples:
  wikiosint pages "Bitcoin" "Ethereum" "Cryptocurrency"
  wikiosint pages "Climate change" "Global warming" --lang en --max-history 180
  wikiosint pages "Company A" "Company B" --enable-deep-analysis --output json`,
	Args: cobra.MinimumNArgs(2),
	RunE: runCrossPageAnalysis,
}

func init() {
	// Flags for cross-page analysis
	pagesCmd.Flags().StringVarP(&pagesOutputFormat, "output", "o", "table", "output format (table, json, yaml)")
	pagesCmd.Flags().StringVarP(&pagesLanguage, "lang", "l", "en", "Wikipedia language (en, fr, de, etc.)")
	pagesCmd.Flags().StringVar(&pagesSaveToFile, "save", "", "save result to file")
	pagesCmd.Flags().IntVar(&pagesMaxRevisions, "max-revisions", 200, "maximum number of revisions per page")
	pagesCmd.Flags().IntVar(&pagesMaxContributors, "max-contributors", 50, "maximum number of contributors per page")
	pagesCmd.Flags().IntVar(&pagesMaxHistory, "max-history", 90, "maximum number of days for detailed history")
	pagesCmd.Flags().IntVar(&crossPageMinCommonEdits, "min-common-edits", 3, "minimum edits to be considered common contributor")
	pagesCmd.Flags().IntVar(&crossPageMaxReactionTime, "max-reaction-time", 60, "maximum minutes for suspicious reaction time")
	pagesCmd.Flags().Float64Var(&crossPageMinSupportRatio, "min-support-ratio", 0.3, "minimum ratio for mutual support detection")
	pagesCmd.Flags().BoolVar(&crossPageEnableDeepAnalysis, "enable-deep-analysis", false, "enable resource-intensive analysis")
}

func runCrossPageAnalysis(cmd *cobra.Command, args []string) error {
	pageNames := args

	// Create Wikipedia client
	wikiClient := client.NewWikipediaClient(pagesLanguage)

	// Create cross-page analysis options
	analysisOptions := models.CrossPageAnalysisOptions{
		MaxRevisionsPerPage:    pagesMaxRevisions,
		MaxContributorsPerPage: pagesMaxContributors,
		HistoryDays:            pagesMaxHistory,
		MinCommonEdits:         crossPageMinCommonEdits,
		MaxReactionTime:        crossPageMaxReactionTime,
		MinMutualSupportRatio:  crossPageMinSupportRatio,
		EnableDeepAnalysis:     crossPageEnableDeepAnalysis,
	}

	// Create cross-page analyzer
	crossPageAnalyzer := analyzer.NewCrossPageAnalyzer(wikiClient, analysisOptions)

	// Start analysis
	fmt.Printf("🔍 Starting cross-page coordination analysis\n")
	fmt.Printf("📄 Pages to analyze: %s\n", strings.Join(pageNames, ", "))
	fmt.Printf("🌍 Wikipedia language: %s\n", pagesLanguage)
	fmt.Printf("📊 Analysis parameters:\n")
	fmt.Printf("   - Max revisions per page: %d\n", pagesMaxRevisions)
	fmt.Printf("   - Max contributors per page: %d\n", pagesMaxContributors)
	fmt.Printf("   - History depth: %d days\n", pagesMaxHistory)
	fmt.Printf("   - Min common edits: %d\n", crossPageMinCommonEdits)
	fmt.Printf("   - Max reaction time: %d minutes\n", crossPageMaxReactionTime)
	fmt.Printf("   - Min support ratio: %.2f\n", crossPageMinSupportRatio)
	if crossPageEnableDeepAnalysis {
		fmt.Printf("   - Deep analysis: enabled\n")
	}
	fmt.Println()

	// Perform analysis
	analysis, err := crossPageAnalyzer.AnalyzePages(pageNames)
	if err != nil {
		return fmt.Errorf("error performing cross-page analysis: %w", err)
	}

	// Format and display results
	output, err := formatter.FormatCrossPageAnalysis(analysis, pagesOutputFormat)
	if err != nil {
		return fmt.Errorf("error formatting output: %w", err)
	}

	// Display or save
	if pagesSaveToFile != "" {
		err = os.WriteFile(pagesSaveToFile, []byte(output), 0644)
		if err != nil {
			return fmt.Errorf("error saving file: %w", err)
		}
		fmt.Printf("✅ Cross-page analysis results saved to: %s\n", pagesSaveToFile)
	} else {
		fmt.Print(output)
	}

	return nil
}
