// internal/models/contribution.go
package models

import (
	"time"
)

// ContributionProfile represents the complete analysis of a Wikipedia contribution
type ContributionProfile struct {
	RevisionID      int                 `json:"revision_id"`
	PageTitle       string              `json:"page_title"`
	PageID          int                 `json:"page_id"`
	Language        string              `json:"language"`
	Timestamp       time.Time           `json:"timestamp"`
	Comment         string              `json:"comment"`
	Size            int                 `json:"size"`
	IsMinor         bool                `json:"is_minor"`
	IsRevert        bool                `json:"is_revert"`
	Author          ContributionAuthor  `json:"author"`
	ContentAnalysis ContributionContent `json:"content_analysis"`
	ContextAnalysis ContributionContext `json:"context_analysis"`
	QualityMetrics  ContributionQuality `json:"quality_metrics"`
	SuspicionScore  int                 `json:"suspicion_score"`
	SuspicionFlags  []string            `json:"suspicion_flags"`
	RetrievedAt     time.Time           `json:"retrieved_at"`
}

// ContributionAuthor represents the author of a contribution
type ContributionAuthor struct {
	Username         string             `json:"username"`
	UserID           int                `json:"user_id"`
	IsAnonymous      bool               `json:"is_anonymous"`
	IsRegistered     bool               `json:"is_registered"`
	IsBlocked        bool               `json:"is_blocked"`
	EditCount        int                `json:"edit_count"`
	Groups           []string           `json:"groups"`
	RegistrationDate *time.Time         `json:"registration_date"`
	RecentActivity   RecentUserActivity `json:"recent_activity"`
	SuspicionScore   int                `json:"suspicion_score"`
}

// RecentUserActivity represents recent activity patterns
type RecentUserActivity struct {
	EditsLast24h int        `json:"edits_last_24h"`
	EditsLast7d  int        `json:"edits_last_7d"`
	EditsLast30d int        `json:"edits_last_30d"`
	PagesEdited  int        `json:"pages_edited"`
	Namespaces   []int      `json:"namespaces"`
	LastEditTime *time.Time `json:"last_edit_time"`
}

// ContributionContent represents content analysis
type ContributionContent struct {
	ContentType      string             `json:"content_type"`
	TextChanges      TextChangeAnalysis `json:"text_changes"`
	LinksAnalysis    LinksAnalysis      `json:"links_analysis"`
	SourcesAnalysis  SourcesAnalysis    `json:"sources_analysis"`
	LanguageAnalysis LanguageAnalysis   `json:"language_analysis"`
}

// TextChangeAnalysis represents analysis of text changes
type TextChangeAnalysis struct {
	CharsAdded       int      `json:"chars_added"`
	CharsRemoved     int      `json:"chars_removed"`
	WordsAdded       int      `json:"words_added"`
	WordsRemoved     int      `json:"words_removed"`
	SectionsAffected []string `json:"sections_affected"`
	IsStructural     bool     `json:"is_structural"`
	IsTrivial        bool     `json:"is_trivial"`
}

// LinksAnalysis represents analysis of link changes
type LinksAnalysis struct {
	LinksAdded    []LinkChange `json:"links_added"`
	LinksRemoved  []LinkChange `json:"links_removed"`
	InternalLinks int          `json:"internal_links"`
	ExternalLinks int          `json:"external_links"`
}

// LinkChange represents a link change
type LinkChange struct {
	Type string `json:"type"` // "internal" or "external"
	URL  string `json:"url"`
	Text string `json:"text"`
}

// SourcesAnalysis represents analysis of sources and citations
type SourcesAnalysis struct {
	CitationsAdded   int     `json:"citations_added"`
	CitationsRemoved int     `json:"citations_removed"`
	SourcesQuality   float64 `json:"sources_quality"`
}

// LanguageAnalysis represents language and POV analysis
type LanguageAnalysis struct {
	Language     string   `json:"language"`
	POVWords     []string `json:"pov_words"`
	BiasScore    float64  `json:"bias_score"`
	ToneAnalysis string   `json:"tone_analysis"`
}

// ContributionContext represents contextual information
type ContributionContext struct {
	PageContext     PageContextInfo     `json:"page_context"`
	TimingContext   TimingContextInfo   `json:"timing_context"`
	AuthorContext   AuthorContextInfo   `json:"author_context"`
	RelatedEdits    []RelatedEdit       `json:"related_edits"`
	ConflictContext ConflictContextInfo `json:"conflict_context"`
}

// PageContextInfo represents page context
type PageContextInfo struct {
	PageAge          int      `json:"page_age"`
	Categories       []string `json:"categories"`
	Controversiality float64  `json:"controversiality"`
}

// TimingContextInfo represents timing context
type TimingContextInfo struct {
	EditHour          int  `json:"edit_hour"`
	EditDayOfWeek     int  `json:"edit_day_of_week"`
	IsWeekend         bool `json:"is_weekend"`
	TimeSinceLastEdit int  `json:"time_since_last_edit"`
}

// AuthorContextInfo represents author context
type AuthorContextInfo struct {
	EditFrequency EditFrequencyInfo `json:"edit_frequency"`
	EditPattern   EditPatternInfo   `json:"edit_pattern"`
	PageFocus     PageFocusInfo     `json:"page_focus"`
}

// EditFrequencyInfo represents edit frequency analysis
type EditFrequencyInfo struct {
	EditsPerDay  float64 `json:"edits_per_day"`
	EditsPerHour float64 `json:"edits_per_hour"`
}

// EditPatternInfo represents edit pattern analysis
type EditPatternInfo struct {
	FavoriteNamespaces []int `json:"favorite_namespaces"`
}

// PageFocusInfo represents page focus analysis
type PageFocusInfo struct {
	PagesEditedCount    int     `json:"pages_edited_count"`
	TopPageEditRatio    float64 `json:"top_page_edit_ratio"`
	IsSpecializedEditor bool    `json:"is_specialized_editor"`
}

// RelatedEdit represents a related edit
type RelatedEdit struct {
	RevisionID int       `json:"revision_id"`
	Author     string    `json:"author"`
	Timestamp  time.Time `json:"timestamp"`
	Relation   string    `json:"relation"`
	Similarity float64   `json:"similarity"`
}

// ConflictContextInfo represents conflict context
type ConflictContextInfo struct {
	IsContested      bool    `json:"is_contested"`
	ConflictSeverity float64 `json:"conflict_severity"`
}

// ContributionQuality represents quality metrics
type ContributionQuality struct {
	ContentQuality   ContentQualityInfo   `json:"content_quality"`
	SourceQuality    SourceQualityInfo    `json:"source_quality"`
	StructureQuality StructureQualityInfo `json:"structure_quality"`
	ComplianceScore  ComplianceInfo       `json:"compliance_score"`
	OverallQuality   float64              `json:"overall_quality"`
}

// ContentQualityInfo represents content quality metrics
type ContentQualityInfo struct {
	Accuracy     float64 `json:"accuracy"`
	Completeness float64 `json:"completeness"`
	Neutrality   float64 `json:"neutrality"`
	Clarity      float64 `json:"clarity"`
	Relevance    float64 `json:"relevance"`
}

// SourceQualityInfo represents source quality metrics
type SourceQualityInfo struct {
	ReliabilityScore float64 `json:"reliability_score"`
	DiversityScore   float64 `json:"diversity_score"`
	RecencyScore     float64 `json:"recency_score"`
	AuthorityScore   float64 `json:"authority_score"`
}

// StructureQualityInfo represents structural quality metrics
type StructureQualityInfo struct {
	Formatting      float64 `json:"formatting"`
	Organization    float64 `json:"organization"`
	WikimarkupScore float64 `json:"wikimarkup_score"`
	LinkingQuality  float64 `json:"linking_quality"`
	CategoryUsage   float64 `json:"category_usage"`
	TemplateUsage   float64 `json:"template_usage"`
}

// ComplianceInfo represents policy compliance information
type ComplianceInfo struct {
	PolicyCompliance    float64  `json:"policy_compliance"`
	GuidelineCompliance float64  `json:"guideline_compliance"`
	COI_Risk            float64  `json:"coi_risk"`
	AdvertisingRisk     float64  `json:"advertising_risk"`
	VandalismRisk       float64  `json:"vandalism_risk"`
	ViolatedPolicies    []string `json:"violated_policies"`
}
