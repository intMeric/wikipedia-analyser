// internal/cli/user.go
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
	outputFormat string
	language     string
	saveToFile   string

	// Revoked contributions analysis options
	maxPagesToAnalyze   int
	maxRevisionsPerPage int
	enableDeepAnalysis  bool
	recentDaysOnly      int
	skipRevokedAnalysis bool
)

// userCmd represents the user command
var userCmd = &cobra.Command{
	Use:   "user",
	Short: "Wikipedia user analysis",
	Long: `Commands to analyze Wikipedia users and detect
suspicious behavior or potential manipulations.`,
}

// profileCmd represents the user profile command
var profileCmd = &cobra.Command{
	Use:   "profile [username]",
	Short: "Display detailed user profile",
	Long: `Retrieves and analyzes a Wikipedia user profile including:
- Basic account information
- Edit statistics
- User groups
- Recent contributions
- Revoked contributions analysis
- Basic suspicion analysis

Revoked Contributions Analysis:
The tool analyzes contributions that have been reverted or undone by other users.
This helps identify potential vandals, sockpuppets, or problematic editors.

Configuration options:
  --max-pages-analyze: Maximum number of pages to analyze for reverts (default: 10)
  --max-revisions-page: Maximum revisions per page to check (default: 50)
  --enable-deep-analysis: Enable thorough analysis (slower but more accurate)
  --recent-days-only: Only analyze contributions from last N days (default: 90)
  --skip-revoked-analysis: Skip revoked contributions analysis entirely

Examples:
  wikiosint user profile "Username"
  wikiosint user profile "Username" --enable-deep-analysis --max-pages-analyze 20
  wikiosint user profile "Username" --recent-days-only 30 --output json`,
	Args: cobra.ExactArgs(1),
	RunE: runUserProfile,
}

func init() {
	// Add subcommands
	userCmd.AddCommand(profileCmd)

	// Flags for profile command
	profileCmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "output format (table, json, yaml)")
	profileCmd.Flags().StringVarP(&language, "lang", "l", "en", "Wikipedia language (en, fr, de, etc.)")
	profileCmd.Flags().StringVar(&saveToFile, "save", "", "save result to file")

	// Revoked contributions analysis flags
	profileCmd.Flags().IntVar(&maxPagesToAnalyze, "max-pages-analyze", 10, "Maximum number of pages to analyze for revoked contributions.")
	profileCmd.Flags().IntVar(&maxRevisionsPerPage, "max-revisions-page", 50, "Maximum number of revisions to check per page for revoked contributions.")
	profileCmd.Flags().BoolVar(&enableDeepAnalysis, "enable-deep-analysis", false, "Enable thorough analysis for revoked contributions (slower but more accurate).")
	profileCmd.Flags().IntVar(&recentDaysOnly, "recent-days-only", 90, "Only analyze revoked contributions from the last N days.")
	profileCmd.Flags().BoolVar(&skipRevokedAnalysis, "skip-revoked-analysis", false, "Skip the entire revoked contributions analysis.")
}

func runUserProfile(cmd *cobra.Command, args []string) error {
	username := args[0]

	// Create Wikipedia client
	wikiClient := client.NewWikipediaClient(language)

	// Create user analyzer
	userAnalyzer := analyzer.NewUserAnalyzer(wikiClient)

	// Configure revoked analysis if not skipped
	if !skipRevokedAnalysis {
		fmt.Printf("🔍 Analyzing user profile: %s\n", username)
		fmt.Printf("📡 Fetching data from %s.wikipedia.org...\n", language)
		fmt.Printf("🚫 Revoked contributions analysis: enabled\n")
		fmt.Printf("   📊 Max pages to analyze: %d\n", maxPagesToAnalyze)
		fmt.Printf("   📄 Max revisions per page: %d\n", maxRevisionsPerPage)
		fmt.Printf("   📅 Recent days only: %d\n", recentDaysOnly)
		if enableDeepAnalysis {
			fmt.Printf("   🔬 Deep analysis: enabled (slower but more accurate)\n")
		} else {
			fmt.Printf("   ⚡ Quick analysis: enabled (faster but less detailed)\n")
		}
	} else {
		fmt.Printf("🔍 Analyzing user profile: %s\n", username)
		fmt.Printf("📡 Fetching data from %s.wikipedia.org...\n", language)
		fmt.Printf("⚠️  Revoked contributions analysis: skipped\n")
	}

	// Create revoked analysis configuration from CLI flags
	var revokedConfig *analyzer.RevokedAnalysisConfig
	if !skipRevokedAnalysis {
		revokedConfig = &analyzer.RevokedAnalysisConfig{
			MaxPagesToAnalyze:   maxPagesToAnalyze,
			MaxRevisionsPerPage: maxRevisionsPerPage,
			EnableDeepAnalysis:  enableDeepAnalysis,
			RecentDaysOnly:      recentDaysOnly,
		}
	}

	// Get user profile with custom configuration
	userProfile, err := userAnalyzer.GetUserProfileWithConfig(username, revokedConfig)
	if err != nil {
		return fmt.Errorf("error retrieving profile: %w", err)
	}

	// Display analysis results summary
	if !skipRevokedAnalysis && userProfile.RevokedCount > 0 {
		fmt.Printf("🚫 Found %d revoked contributions (%.1f%% of total)\n",
			userProfile.RevokedCount, userProfile.RevokedRatio*100)

		if userProfile.RevokedRatio > 0.3 {
			fmt.Printf("⚠️  High revocation rate detected - potential issues\n")
		}
	}

	// Format and display results
	output, err := formatter.FormatUserProfile(userProfile, outputFormat)
	if err != nil {
		return fmt.Errorf("error formatting output: %w", err)
	}

	// Display or save
	if saveToFile != "" {
		err = os.WriteFile(saveToFile, []byte(output), 0644)
		if err != nil {
			return fmt.Errorf("error saving file: %w", err)
		}
		fmt.Printf("✅ Results saved to: %s\n", saveToFile)
	} else {
		fmt.Print(output)
	}

	return nil
}
