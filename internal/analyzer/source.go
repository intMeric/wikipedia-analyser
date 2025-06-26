package analyzer

import (
	"net/url"
	"regexp"
	"strings"

	"github.com/intMeric/wikipedia-analyser/internal/models"
)

type SourceAnalyzer struct {
	refPattern      *regexp.Regexp
	namedRefPattern *regexp.Regexp
	urlPattern      *regexp.Regexp
	templatePattern *regexp.Regexp
	reliableDomains map[string]string
}

func NewSourceAnalyzer() *SourceAnalyzer {
	return &SourceAnalyzer{
		refPattern:      regexp.MustCompile(`<ref[^>]*>([^<]+)</ref>`),
		namedRefPattern: regexp.MustCompile(`<ref\s+name\s*=\s*["']([^"']+)["'][^>]*>([^<]*)</ref>`),
		urlPattern:      regexp.MustCompile(`https?://[^\s\]]+`),
		templatePattern: regexp.MustCompile(`\{\{cite\s+(\w+)`),
		reliableDomains: getReliableDomains(),
	}
}

func (sa *SourceAnalyzer) AnalyzePageSources(wikitext string) *models.SourceAnalysis {
	references := sa.extractReferences(wikitext)
	if len(references) == 0 {
		return &models.SourceAnalysis{
			TotalReferences:    0,
			UniqueReferences:   0,
			DomainDistribution: make(map[string]int),
			TemplateUsage:      make(map[string]int),
			ReliabilityScore:   0.0,
			UnreliableSources:  []models.UnreliableSource{},
			DeadLinks:          []models.DeadLink{},
		}
	}

	domainDist := sa.analyzeDomains(references)
	templateUsage := sa.analyzeTemplates(references)
	reliabilityScore := sa.calculateReliabilityScore(domainDist)
	unreliableSources := sa.identifyUnreliableSources(references)

	return &models.SourceAnalysis{
		TotalReferences:    len(references),
		UniqueReferences:   sa.countUniqueReferences(references),
		DomainDistribution: domainDist,
		TemplateUsage:      templateUsage,
		ReliabilityScore:   reliabilityScore,
		UnreliableSources:  unreliableSources,
		DeadLinks:          []models.DeadLink{},
	}
}

func (sa *SourceAnalyzer) extractReferences(wikitext string) []models.Reference {
	var references []models.Reference
	namedRefs := make(map[string]int)

	namedMatches := sa.namedRefPattern.FindAllStringSubmatch(wikitext, -1)
	for _, match := range namedMatches {
		if len(match) >= 3 {
			refName := match[1]
			content := match[2]
			namedRefs[refName]++

			if content != "" {
				ref := models.Reference{
					Content:    strings.TrimSpace(content),
					IsNamed:    true,
					Name:       refName,
					UsageCount: 1,
				}
				sa.enrichReference(&ref)
				references = append(references, ref)
			}
		}
	}

	regularMatches := sa.refPattern.FindAllStringSubmatch(wikitext, -1)
	for _, match := range regularMatches {
		if len(match) >= 2 {
			content := strings.TrimSpace(match[1])
			if content != "" && !sa.isNamedRefReuse(match[0]) {
				ref := models.Reference{
					Content:    content,
					IsNamed:    false,
					UsageCount: 1,
				}
				sa.enrichReference(&ref)
				references = append(references, ref)
			}
		}
	}

	for refName, count := range namedRefs {
		for i := range references {
			if references[i].Name == refName {
				references[i].UsageCount = count
				break
			}
		}
	}

	return references
}

func (sa *SourceAnalyzer) enrichReference(ref *models.Reference) {
	urls := sa.urlPattern.FindAllString(ref.Content, -1)
	if len(urls) > 0 {
		ref.URL = urls[0]
		if parsedURL, err := url.Parse(ref.URL); err == nil {
			ref.Domain = parsedURL.Host
		}
	}

	templateMatches := sa.templatePattern.FindStringSubmatch(ref.Content)
	if len(templateMatches) >= 2 {
		ref.Template = strings.ToLower(templateMatches[1])
	}
}

func (sa *SourceAnalyzer) isNamedRefReuse(refTag string) bool {
	return strings.Contains(refTag, `name=`) && strings.Contains(refTag, `/>`)
}

func (sa *SourceAnalyzer) countUniqueReferences(references []models.Reference) int {
	seen := make(map[string]bool)
	for _, ref := range references {
		key := ref.Content
		if ref.URL != "" {
			key = ref.URL
		}
		seen[key] = true
	}
	return len(seen)
}

func (sa *SourceAnalyzer) analyzeDomains(references []models.Reference) map[string]int {
	domainCounts := make(map[string]int)
	for _, ref := range references {
		if ref.Domain != "" {
			domain := strings.ToLower(ref.Domain)
			domain = strings.TrimPrefix(domain, "www.")
			domainCounts[domain]++
		}
	}
	return domainCounts
}

func (sa *SourceAnalyzer) analyzeTemplates(references []models.Reference) map[string]int {
	templateCounts := make(map[string]int)
	for _, ref := range references {
		if ref.Template != "" {
			templateCounts[ref.Template]++
		}
	}
	return templateCounts
}

func (sa *SourceAnalyzer) calculateReliabilityScore(domainDist map[string]int) float64 {
	totalSources := 0
	reliableSources := 0

	for domain, count := range domainDist {
		totalSources += count
		if reliability, exists := sa.reliableDomains[domain]; exists {
			if reliability == "reliable" {
				reliableSources += count
			}
		}
	}

	if totalSources == 0 {
		return 0.0
	}

	return float64(reliableSources) / float64(totalSources) * 100.0
}

func (sa *SourceAnalyzer) identifyUnreliableSources(references []models.Reference) []models.UnreliableSource {
	var unreliable []models.UnreliableSource
	domainCounts := make(map[string]int)

	for _, ref := range references {
		if ref.Domain != "" {
			domain := strings.ToLower(strings.TrimPrefix(ref.Domain, "www."))
			domainCounts[domain] += ref.UsageCount

			if reliability, exists := sa.reliableDomains[domain]; exists && reliability != "reliable" {
				unreliable = append(unreliable, models.UnreliableSource{
					URL:              ref.URL,
					Domain:           domain,
					ReliabilityLevel: reliability,
					Reason:           getUnreliabilityReason(reliability),
					UsageCount:       ref.UsageCount,
				})
			}
		}
	}

	return unreliable
}

func getReliableDomains() map[string]string {
	return map[string]string{
		"pubmed.ncbi.nlm.nih.gov": "reliable",
		"doi.org":                 "reliable",
		"nature.com":              "reliable",
		"science.org":             "reliable",
		"bbc.com":                 "reliable",
		"reuters.com":             "reliable",
		"gov":                     "reliable",
		"edu":                     "reliable",
		"lemonde.fr":              "reliable",
		"lefigaro.fr":             "reliable",
		"liberation.fr":           "reliable",
		"wikipedia.org":           "questionable",
		"blog":                    "unreliable",
		"blogspot.com":            "unreliable",
		"wordpress.com":           "questionable",
		"youtube.com":             "questionable",
		"facebook.com":            "unreliable",
		"twitter.com":             "unreliable",
		"reddit.com":              "unreliable",
	}
}

func getUnreliabilityReason(level string) string {
	switch level {
	case "unreliable":
		return "Source généralement considérée comme non fiable"
	case "questionable":
		return "Fiabilité à vérifier selon le contexte"
	case "deprecated":
		return "Source obsolète ou dépréciée"
	default:
		return "Niveau de fiabilité indéterminé"
	}
}