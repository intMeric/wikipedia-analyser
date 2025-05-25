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
- Basic suspicion analysis`,
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
}

func runUserProfile(cmd *cobra.Command, args []string) error {
	username := args[0]

	// Create Wikipedia client
	wikiClient := client.NewWikipediaClient(language)

	// Create user analyzer
	userAnalyzer := analyzer.NewUserAnalyzer(wikiClient)

	// Retrieve user data
	fmt.Printf("🔍 Analyzing user profile: %s\n", username)
	fmt.Printf("📡 Fetching data from %s.wikipedia.org...\n", language)

	userProfile, err := userAnalyzer.GetUserProfile(username)
	if err != nil {
		return fmt.Errorf("error retrieving profile: %w", err)
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
