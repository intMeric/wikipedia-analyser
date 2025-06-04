// internal/analyzer/cross_page.go
package analyzer

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/intMeric/wikipedia-analyser/internal/client"
	"github.com/intMeric/wikipedia-analyser/internal/models"
	"github.com/intMeric/wikipedia-analyser/internal/utils"
)

// CrossPageAnalyzer analyzes patterns across multiple Wikipedia pages
type CrossPageAnalyzer struct {
	client       *client.WikipediaClient
	pageAnalyzer *PageAnalyzer
	options      models.CrossPageAnalysisOptions
}

// NewCrossPageAnalyzer creates a new cross-page analyzer
func NewCrossPageAnalyzer(client *client.WikipediaClient, options models.CrossPageAnalysisOptions) *CrossPageAnalyzer {
	// Set default options
	if options.MaxRevisionsPerPage == 0 {
		options.MaxRevisionsPerPage = 200
	}
	if options.MaxContributorsPerPage == 0 {
		options.MaxContributorsPerPage = 50
	}
	if options.HistoryDays == 0 {
		options.HistoryDays = 90
	}
	if options.MinCommonEdits == 0 {
		options.MinCommonEdits = 3
	}
	if options.MaxReactionTime == 0 {
		options.MaxReactionTime = 60 // 1 hour
	}
	if options.MinMutualSupportRatio == 0 {
		options.MinMutualSupportRatio = 0.3
	}

	pageAnalysisOptions := PageAnalysisOptions{
		NumberOfPageRevisions: options.MaxRevisionsPerPage,
		NumberOfDaysHistory:   options.HistoryDays,
		NumberOfContributors:  options.MaxContributorsPerPage,
	}

	return &CrossPageAnalyzer{
		client:       client,
		pageAnalyzer: NewPageAnalyzer(client, pageAnalysisOptions),
		options:      options,
	}
}

// AnalyzePages performs cross-page analysis on multiple pages
func (cpa *CrossPageAnalyzer) AnalyzePages(pageNames []string) (*models.CrossPageAnalysis, error) {
	fmt.Printf("[PAGES ANALYZER]🔍 Starting cross-page analysis of %d pages...\n", len(pageNames))

	// 1. Analyze each page individually
	pageProfiles := make(map[string]*models.PageProfile)
	allContributors := make(map[string]*models.CommonContributor)
	allRevisions := []models.EditEvent{}

	for i, pageName := range pageNames {
		fmt.Printf("[PAGES ANALYZER]📄 Analyzing page %d/%d: %s\n", i+1, len(pageNames), pageName)

		profile, err := cpa.pageAnalyzer.GetPageProfile(pageName)
		if err != nil {
			fmt.Printf("[PAGES ANALYZER]⚠️ Failed to analyze page %s: %v\n", pageName, err)
			continue
		}

		pageProfiles[pageName] = profile

		// Extract contributors and revisions for cross-page analysis
		cpa.extractContributors(profile, pageName, allContributors)
		cpa.extractRevisions(profile, pageName, &allRevisions)
	}

	fmt.Printf("[PAGES ANALYZER]📊 Found %d unique contributors across all pages\n", len(allContributors))

	// 2. Identify common contributors
	commonContributors := cpa.identifyCommonContributors(allContributors)

	// 3. Analyze coordination patterns
	coordinatedPatterns := cpa.analyzeCoordinationPatterns(commonContributors, allRevisions)

	// 4. Analyze temporal patterns
	temporalPatterns := cpa.analyzeTemporalPatterns(allRevisions, commonContributors)

	// 5. Detect sockpuppet networks
	sockpuppetNetworks := cpa.detectSockpuppetNetworks(commonContributors, allRevisions)

	// 6. Calculate overall suspicion score
	suspicionScore, suspicionFlags := cpa.calculateCrossPageSuspicion(
		coordinatedPatterns, temporalPatterns, sockpuppetNetworks, commonContributors)

	analysis := &models.CrossPageAnalysis{
		Pages:               pageNames,
		Language:            cpa.client.Language(),
		TotalPages:          len(pageProfiles),
		TotalContributors:   len(allContributors),
		CommonContributors:  commonContributors,
		CoordinatedPatterns: coordinatedPatterns,
		TemporalPatterns:    temporalPatterns,
		SockpuppetNetworks:  sockpuppetNetworks,
		SuspicionScore:      suspicionScore,
		SuspicionFlags:      suspicionFlags,
		AnalysisTimestamp:   time.Now(),
		PageProfiles:        pageProfiles,
	}

	fmt.Printf("[PAGES ANALYZER]✅ Cross-page analysis completed. Suspicion score: %d/100\n", suspicionScore)
	return analysis, nil
}

// extractContributors extracts contributors from a page profile
func (cpa *CrossPageAnalyzer) extractContributors(profile *models.PageProfile, pageName string, allContributors map[string]*models.CommonContributor) {
	for _, contributor := range profile.Contributors {
		if existing, exists := allContributors[contributor.Username]; exists {
			// Update existing contributor
			existing.PagesEdited = append(existing.PagesEdited, pageName)
			existing.TotalEdits += contributor.EditCount
			existing.EditsByPage[pageName] = contributor.EditCount

			if contributor.FirstEdit.Before(existing.FirstEdit) {
				existing.FirstEdit = contributor.FirstEdit
			}
			if contributor.LastEdit.After(existing.LastEdit) {
				existing.LastEdit = contributor.LastEdit
			}
		} else {
			// Create new common contributor
			allContributors[contributor.Username] = &models.CommonContributor{
				Username:            contributor.Username,
				UserID:              contributor.UserID,
				PagesEdited:         []string{pageName},
				TotalEdits:          contributor.EditCount,
				EditsByPage:         map[string]int{pageName: contributor.EditCount},
				FirstEdit:           contributor.FirstEdit,
				LastEdit:            contributor.LastEdit,
				SuspicionScore:      contributor.SuspicionScore,
				SuspicionFlags:      contributor.SuspicionFlags,
				MutualSupportEvents: []models.MutualSupportEvent{},
				IsAnonymous:         contributor.IsAnonymous,
			}
		}
	}
}

// extractRevisions extracts revisions as edit events
func (cpa *CrossPageAnalyzer) extractRevisions(profile *models.PageProfile, pageName string, allRevisions *[]models.EditEvent) {
	for _, revision := range profile.RecentRevisions {
		editEvent := models.EditEvent{
			Timestamp:  revision.Timestamp,
			Username:   revision.Username,
			PageTitle:  pageName,
			RevisionID: revision.RevID,
			SizeDiff:   revision.SizeDiff,
			Comment:    revision.Comment,
			IsRevert:   revision.IsRevert,
		}
		*allRevisions = append(*allRevisions, editEvent)
	}
}

// identifyCommonContributors filters contributors who edited multiple pages
func (cpa *CrossPageAnalyzer) identifyCommonContributors(allContributors map[string]*models.CommonContributor) []models.CommonContributor {
	var commonContributors []models.CommonContributor

	for _, contributor := range allContributors {
		// Include contributors who edited multiple pages OR have high edit count on single page
		if len(contributor.PagesEdited) > 1 || contributor.TotalEdits >= cpa.options.MinCommonEdits {
			commonContributors = append(commonContributors, *contributor)
		}
	}

	// Sort by total edits descending
	sort.Slice(commonContributors, func(i, j int) bool {
		return commonContributors[i].TotalEdits > commonContributors[j].TotalEdits
	})

	return commonContributors
}

// analyzeCoordinationPatterns detects coordinated editing patterns
func (cpa *CrossPageAnalyzer) analyzeCoordinationPatterns(contributors []models.CommonContributor, revisions []models.EditEvent) models.CoordinatedPatterns {
	patterns := models.CoordinatedPatterns{
		MutualSupportPairs:    []models.MutualSupportPair{},
		TagTeamEditing:        []models.TagTeamPattern{},
		CoordinatedReversions: []models.CoordinatedRevert{},
		SupportNetworks:       []models.SupportNetwork{},
	}

	// 1. Detect mutual support pairs
	mutualSupportPairs := cpa.detectMutualSupport(contributors, revisions)
	patterns.MutualSupportPairs = mutualSupportPairs

	// 2. Detect tag-team editing
	tagTeamPatterns := cpa.detectTagTeamEditing(contributors, revisions)
	patterns.TagTeamEditing = tagTeamPatterns

	// 3. Detect coordinated reversions
	coordinatedReverts := cpa.detectCoordinatedReversions(revisions)
	patterns.CoordinatedReversions = coordinatedReverts

	// 4. Build support networks
	supportNetworks := cpa.buildSupportNetworks(mutualSupportPairs, contributors)
	patterns.SupportNetworks = supportNetworks

	// Calculate overall coordination score
	patterns.CoordinationScore = cpa.calculateCoordinationScore(patterns)

	return patterns
}

// detectMutualSupport identifies pairs of users who defend each other
func (cpa *CrossPageAnalyzer) detectMutualSupport(contributors []models.CommonContributor, revisions []models.EditEvent) []models.MutualSupportPair {
	var mutualSupportPairs []models.MutualSupportPair

	// Sort revisions by timestamp
	sort.Slice(revisions, func(i, j int) bool {
		return revisions[i].Timestamp.Before(revisions[j].Timestamp)
	})

	// Create user pairs to analyze
	userPairs := cpa.createUserPairs(contributors)

	for _, pair := range userPairs {
		supportEvents := cpa.findSupportEvents(pair[0], pair[1], revisions)

		if len(supportEvents) == 0 {
			continue
		}

		// Calculate support metrics
		mutualSupportRatio := cpa.calculateMutualSupportRatio(supportEvents, pair[0], pair[1])
		averageReactionTime := cpa.calculateAverageReactionTime(supportEvents)
		reciprocityScore := cpa.calculateReciprocityScore(supportEvents, pair[0], pair[1])
		exclusivityRatio := cpa.calculateExclusivityRatio(supportEvents, pair[0], pair[1], contributors)

		// Determine suspicion level
		suspicionLevel := cpa.determineSupportSuspicionLevel(mutualSupportRatio, float64(averageReactionTime), reciprocityScore, exclusivityRatio)

		if suspicionLevel != "NONE" {
			pagesInvolved := cpa.extractPagesFromSupportEvents(supportEvents)

			mutualSupportPair := models.MutualSupportPair{
				UserA:               pair[0],
				UserB:               pair[1],
				SupportEvents:       supportEvents,
				MutualSupportRatio:  mutualSupportRatio,
				AverageReactionTime: averageReactionTime,
				ReciprocityScore:    reciprocityScore,
				ExclusivityRatio:    exclusivityRatio,
				PagesInvolved:       pagesInvolved,
				SuspicionLevel:      suspicionLevel,
			}

			mutualSupportPairs = append(mutualSupportPairs, mutualSupportPair)
		}
	}

	// Sort by suspicion level and support ratio
	sort.Slice(mutualSupportPairs, func(i, j int) bool {
		if mutualSupportPairs[i].SuspicionLevel != mutualSupportPairs[j].SuspicionLevel {
			return cpa.getSuspicionLevelScore(mutualSupportPairs[i].SuspicionLevel) >
				cpa.getSuspicionLevelScore(mutualSupportPairs[j].SuspicionLevel)
		}
		return mutualSupportPairs[i].MutualSupportRatio > mutualSupportPairs[j].MutualSupportRatio
	})

	return mutualSupportPairs
}

// createUserPairs creates all possible pairs of users for analysis
func (cpa *CrossPageAnalyzer) createUserPairs(contributors []models.CommonContributor) [][2]string {
	var pairs [][2]string

	for i := 0; i < len(contributors); i++ {
		for j := i + 1; j < len(contributors); j++ {
			// Skip anonymous users as they can't have sockpuppet relationships
			if !contributors[i].IsAnonymous && !contributors[j].IsAnonymous {
				pairs = append(pairs, [2]string{contributors[i].Username, contributors[j].Username})
			}
		}
	}

	return pairs
}

// findSupportEvents finds events where one user supports another
func (cpa *CrossPageAnalyzer) findSupportEvents(userA, userB string, revisions []models.EditEvent) []models.MutualSupportEvent {
	var supportEvents []models.MutualSupportEvent

	for i := 0; i < len(revisions)-1; i++ {
		currentEdit := revisions[i]

		// Look for potential support scenarios within the next few edits
		for j := i + 1; j < len(revisions) && j < i+10; j++ {
			nextEdit := revisions[j]

			// Check if this could be a support event
			if cpa.isSupportEvent(currentEdit, nextEdit, userA, userB) {
				reactionTime := int(nextEdit.Timestamp.Sub(currentEdit.Timestamp).Minutes())

				// Only consider as support if reaction is within reasonable time
				if reactionTime <= cpa.options.MaxReactionTime {
					supportType := cpa.determineSupportType(currentEdit, nextEdit)

					supportEvent := models.MutualSupportEvent{
						Timestamp:     nextEdit.Timestamp,
						PageTitle:     nextEdit.PageTitle,
						SupportType:   supportType,
						ReactionTime:  reactionTime,
						AttackerUser:  currentEdit.Username,
						DefenderUser:  nextEdit.Username,
						SupportedUser: cpa.determineSupportedUser(currentEdit, nextEdit, userA, userB),
						RevisionID:    nextEdit.RevisionID,
						Comment:       nextEdit.Comment,
					}

					supportEvents = append(supportEvents, supportEvent)
				}
			}
		}
	}

	return supportEvents
}

// isSupportEvent checks if the second edit supports the first user against the first edit
func (cpa *CrossPageAnalyzer) isSupportEvent(edit1, edit2 models.EditEvent, userA, userB string) bool {
	// Same page
	if edit1.PageTitle != edit2.PageTitle {
		return false
	}

	// Different users involved
	if edit1.Username == edit2.Username {
		return false
	}

	// One of our target users is involved in the support
	if edit2.Username != userA && edit2.Username != userB {
		return false
	}

	// Potential support scenarios:
	// 1. Edit1 is a revert, Edit2 restores content (revert defense)
	// 2. Edit1 removes content from userA/userB, Edit2 restores it
	// 3. Edit2 is defending against edit1's changes

	return edit1.IsRevert ||
		edit2.IsRevert ||
		cpa.isContentRestoration(edit1, edit2) ||
		cpa.isDefensiveEdit(edit1, edit2)
}

// Helper functions for support event detection
func (cpa *CrossPageAnalyzer) isContentRestoration(edit1, edit2 models.EditEvent) bool {
	// If edit1 removed content (negative size) and edit2 adds similar amount back
	return edit1.SizeDiff < -100 && edit2.SizeDiff > 50
}

func (cpa *CrossPageAnalyzer) isDefensiveEdit(edit1, edit2 models.EditEvent) bool {
	// Check edit comments for defensive language
	comment := strings.ToLower(edit2.Comment)
	defensiveWords := []string{"restore", "fix", "correct", "undo", "revert vandalism", "rv"}

	for _, word := range defensiveWords {
		if strings.Contains(comment, word) {
			return true
		}
	}
	return false
}

func (cpa *CrossPageAnalyzer) determineSupportType(edit1, edit2 models.EditEvent) string {
	if edit1.IsRevert && !edit2.IsRevert {
		return "revert_defense"
	}
	if edit2.IsRevert {
		return "counter_revert"
	}
	if cpa.isContentRestoration(edit1, edit2) {
		return "content_restoration"
	}
	return "defensive_edit"
}

func (cpa *CrossPageAnalyzer) determineSupportedUser(edit1, edit2 models.EditEvent, userA, userB string) string {
	// The supported user is the one being defended
	if edit2.Username == userA {
		return userB // userA is defending userB
	}
	if edit2.Username == userB {
		return userA // userB is defending userA
	}
	return ""
}

// Calculation functions for mutual support metrics
func (cpa *CrossPageAnalyzer) calculateMutualSupportRatio(events []models.MutualSupportEvent, userA, userB string) float64 {
	if len(events) == 0 {
		return 0.0
	}

	aSupportsB := 0
	bSupportsA := 0

	for _, event := range events {
		if event.DefenderUser == userA && event.SupportedUser == userB {
			aSupportsB++
		} else if event.DefenderUser == userB && event.SupportedUser == userA {
			bSupportsA++
		}
	}

	totalSupport := aSupportsB + bSupportsA
	if totalSupport == 0 {
		return 0.0
	}

	return float64(totalSupport) / float64(len(events))
}

func (cpa *CrossPageAnalyzer) calculateAverageReactionTime(events []models.MutualSupportEvent) int {
	if len(events) == 0 {
		return 0
	}

	totalTime := 0
	for _, event := range events {
		totalTime += event.ReactionTime
	}

	return totalTime / len(events)
}

func (cpa *CrossPageAnalyzer) calculateReciprocityScore(events []models.MutualSupportEvent, userA, userB string) float64 {
	aSupportsB := 0
	bSupportsA := 0

	for _, event := range events {
		if event.DefenderUser == userA && event.SupportedUser == userB {
			aSupportsB++
		} else if event.DefenderUser == userB && event.SupportedUser == userA {
			bSupportsA++
		}
	}

	if aSupportsB == 0 && bSupportsA == 0 {
		return 0.0
	}

	min := utils.Min(aSupportsB, bSupportsA)
	max := utils.Max(aSupportsB, bSupportsA)

	if max == 0 {
		return 0.0
	}

	return float64(min) / float64(max)
}

func (cpa *CrossPageAnalyzer) calculateExclusivityRatio(events []models.MutualSupportEvent, userA, userB string, contributors []models.CommonContributor) float64 {
	// This is a simplified implementation
	// In a real implementation, you'd analyze if these users only support each other
	mutualEvents := 0

	for _, event := range events {
		if (event.DefenderUser == userA && event.SupportedUser == userB) ||
			(event.DefenderUser == userB && event.SupportedUser == userA) {
			mutualEvents++
		}
	}

	if len(events) == 0 {
		return 0.0
	}

	return float64(mutualEvents) / float64(len(events))
}

func (cpa *CrossPageAnalyzer) determineSupportSuspicionLevel(mutualRatio, avgReactionTime, reciprocity, exclusivity float64) string {
	score := 0

	if mutualRatio > 0.7 {
		score += 3
	} else if mutualRatio > 0.5 {
		score += 2
	} else if mutualRatio > 0.3 {
		score += 1
	}

	if avgReactionTime < 10 {
		score += 3
	} else if avgReactionTime < 30 {
		score += 2
	} else if avgReactionTime < 60 {
		score += 1
	}

	if reciprocity > 0.8 {
		score += 2
	} else if reciprocity > 0.6 {
		score += 1
	}

	if exclusivity > 0.9 {
		score += 2
	} else if exclusivity > 0.7 {
		score += 1
	}

	switch {
	case score >= 8:
		return "VERY_HIGH"
	case score >= 6:
		return "HIGH"
	case score >= 4:
		return "MODERATE"
	case score >= 2:
		return "LOW"
	default:
		return "NONE"
	}
}

func (cpa *CrossPageAnalyzer) getSuspicionLevelScore(level string) int {
	switch level {
	case "VERY_HIGH":
		return 5
	case "HIGH":
		return 4
	case "MODERATE":
		return 3
	case "LOW":
		return 2
	case "NONE":
		return 1
	default:
		return 0
	}
}

func (cpa *CrossPageAnalyzer) extractPagesFromSupportEvents(events []models.MutualSupportEvent) []string {
	pageSet := make(map[string]bool)
	for _, event := range events {
		pageSet[event.PageTitle] = true
	}

	var pages []string
	for page := range pageSet {
		pages = append(pages, page)
	}

	return pages
}

// Stub implementations for other analysis methods (to be implemented later)
func (cpa *CrossPageAnalyzer) detectTagTeamEditing(contributors []models.CommonContributor, revisions []models.EditEvent) []models.TagTeamPattern {
	// TODO: Implement tag-team editing detection
	return []models.TagTeamPattern{}
}

func (cpa *CrossPageAnalyzer) detectCoordinatedReversions(revisions []models.EditEvent) []models.CoordinatedRevert {
	// TODO: Implement coordinated reversion detection
	return []models.CoordinatedRevert{}
}

func (cpa *CrossPageAnalyzer) buildSupportNetworks(pairs []models.MutualSupportPair, contributors []models.CommonContributor) []models.SupportNetwork {
	// TODO: Implement support network building
	return []models.SupportNetwork{}
}

func (cpa *CrossPageAnalyzer) calculateCoordinationScore(patterns models.CoordinatedPatterns) float64 {
	score := 0.0

	// Add points for each suspicious pattern
	score += float64(len(patterns.MutualSupportPairs)) * 10.0
	score += float64(len(patterns.TagTeamEditing)) * 15.0
	score += float64(len(patterns.CoordinatedReversions)) * 20.0
	score += float64(len(patterns.SupportNetworks)) * 25.0

	// Normalize to 0-100 scale
	if score > 100 {
		score = 100
	}

	return score
}

func (cpa *CrossPageAnalyzer) analyzeTemporalPatterns(revisions []models.EditEvent, contributors []models.CommonContributor) models.TemporalPatterns {
	// TODO: Implement temporal pattern analysis
	return models.TemporalPatterns{
		SynchronizedEditing:   []models.SynchronizedEvent{},
		EditingWaves:          []models.EditingWave{},
		TimeZonePatterns:      []models.TimeZonePattern{},
		TemporalCorrelation:   0.0,
		SuspiciousTimeWindows: []models.SuspiciousTimeWindow{},
	}
}

func (cpa *CrossPageAnalyzer) detectSockpuppetNetworks(contributors []models.CommonContributor, revisions []models.EditEvent) []models.SockpuppetNetwork {
	// TODO: Implement sockpuppet network detection
	return []models.SockpuppetNetwork{}
}

func (cpa *CrossPageAnalyzer) calculateCrossPageSuspicion(
	coordinated models.CoordinatedPatterns,
	temporal models.TemporalPatterns,
	sockpuppets []models.SockpuppetNetwork,
	contributors []models.CommonContributor) (int, []string) {

	score := 0
	flags := []string{}

	// Coordination patterns
	if len(coordinated.MutualSupportPairs) > 0 {
		score += 25
		flags = append(flags, "MUTUAL_SUPPORT_DETECTED")
	}

	if coordinated.CoordinationScore > 50 {
		score += 20
		flags = append(flags, "HIGH_COORDINATION_SCORE")
	}

	// Sockpuppet networks
	if len(sockpuppets) > 0 {
		score += 30
		flags = append(flags, "SOCKPUPPET_NETWORK_DETECTED")
	}

	// High overlap of contributors
	multiPageContributors := 0
	for _, contributor := range contributors {
		if len(contributor.PagesEdited) > 1 {
			multiPageContributors++
		}
	}

	if multiPageContributors > len(contributors)/2 {
		score += 15
		flags = append(flags, "HIGH_CONTRIBUTOR_OVERLAP")
	}

	// Limit score to 100
	if score > 100 {
		score = 100
	}

	return score, flags
}
