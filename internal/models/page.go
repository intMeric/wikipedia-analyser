// internal/models/page.go
package models

import (
	"time"
)

// PageProfile represents the complete profile of a Wikipedia page
type PageProfile struct {
	PageTitle       string           `json:"page_title"`
	PageID          int              `json:"page_id"`
	Namespace       int              `json:"namespace"`
	Language        string           `json:"language"`
	CreationDate    *time.Time       `json:"creation_date"`
	LastModified    time.Time        `json:"last_modified"`
	TotalRevisions  int              `json:"total_revisions"`
	PageSize        int              `json:"page_size"`
	Contributors    []TopContributor `json:"top_contributors"`
	RecentRevisions []Revision       `json:"recent_revisions"`
	ConflictStats   ConflictStats    `json:"conflict_stats"`
	QualityMetrics  QualityMetrics   `json:"quality_metrics"`
	SuspicionScore  int              `json:"suspicion_score"`
	SuspicionFlags  []string         `json:"suspicion_flags"`
	RetrievedAt     time.Time        `json:"retrieved_at"`
}

// TopContributor represents a major contributor to the page
type TopContributor struct {
	Username       string    `json:"username"`
	UserID         int       `json:"user_id,omitempty"`
	EditCount      int       `json:"edit_count"`
	FirstEdit      time.Time `json:"first_edit"`
	LastEdit       time.Time `json:"last_edit"`
	TotalSizeDiff  int       `json:"total_size_diff"`
	IsAnonymous    bool      `json:"is_anonymous"`
	IsRegistered   bool      `json:"is_registered"`
	SuspicionScore int       `json:"suspicion_score"`
	SuspicionFlags []string  `json:"suspicion_flags"`
	AnalysisError  string    `json:"analysis_error,omitempty"`
}

// Revision represents a single page revision
type Revision struct {
	RevID       int       `json:"rev_id"`
	ParentID    int       `json:"parent_id"`
	Username    string    `json:"username"`
	UserID      int       `json:"user_id,omitempty"`
	Timestamp   time.Time `json:"timestamp"`
	Comment     string    `json:"comment"`
	SizeDiff    int       `json:"size_diff"`
	NewSize     int       `json:"new_size"`
	IsMinor     bool      `json:"is_minor"`
	IsRevert    bool      `json:"is_revert"`
	IsAnonymous bool      `json:"is_anonymous"`
}

// ConflictStats contains conflict analysis metrics
type ConflictStats struct {
	ReversionsCount  int             `json:"reversions_count"`
	ConflictingUsers []string        `json:"conflicting_users"`
	EditWarPeriods   []EditWarPeriod `json:"edit_war_periods"`
	StabilityScore   float64         `json:"stability_score"`
	ControversyScore float64         `json:"controversy_score"`
	RecentConflicts  int             `json:"recent_conflicts_7_days"`
}

// EditWarPeriod represents a period of intensive editing conflicts
type EditWarPeriod struct {
	StartTime     time.Time `json:"start_time"`
	EndTime       time.Time `json:"end_time"`
	Participants  []string  `json:"participants"`
	RevisionCount int       `json:"revision_count"`
}

// QualityMetrics contains page quality indicators
type QualityMetrics struct {
	AverageEditSize      float64       `json:"average_edit_size"`
	AnonymousEditRatio   float64       `json:"anonymous_edit_ratio"`
	NewEditorRatio       float64       `json:"new_editor_ratio"`
	RecentActivityBurst  bool          `json:"recent_activity_burst"`
	ContributorDiversity float64       `json:"contributor_diversity"`
	EditFrequency        EditFrequency `json:"edit_frequency"`
}

// EditFrequency contains editing frequency analysis
type EditFrequency struct {
	EditsLast7Days   int            `json:"edits_last_7_days"`
	EditsLast30Days  int            `json:"edits_last_30_days"`
	EditsLast90Days  int            `json:"edits_last_90_days"`
	PeakEditingHours []int          `json:"peak_editing_hours"`
	EditsByDay       map[string]int `json:"edits_by_day"`
}

// API Response structures for MediaWiki API

// WikiPageInfo represents page information from the API
type WikiPageInfo struct {
	PageID    int    `json:"pageid"`
	NS        int    `json:"ns"`
	Title     string `json:"title"`
	Touched   string `json:"touched"`
	LastRevID int    `json:"lastrevid"`
	Length    int    `json:"length"`
	Missing   string `json:"missing,omitempty"`
}

// WikiRevision represents a revision from the API
type WikiRevision struct {
	RevID     int    `json:"revid"`
	ParentID  int    `json:"parentid"`
	User      string `json:"user"`
	UserID    int    `json:"userid,omitempty"`
	Timestamp string `json:"timestamp"`
	Size      int    `json:"size"`
	Comment   string `json:"comment"`
	Minor     string `json:"minor,omitempty"`
	Anon      string `json:"anon,omitempty"`
}

// WikiContributor represents a contributor from the API
type WikiContributor struct {
	UserID    int    `json:"userid,omitempty"`
	Name      string `json:"name"`
	EditCount int    `json:"editcount"`
	Anon      string `json:"anon,omitempty"`
}

// PageAnalysisRequest represents the request for page analysis
type PageAnalysisRequest struct {
	PageTitle       string `json:"page_title"`
	Language        string `json:"language"`
	MaxRevisions    int    `json:"max_revisions"`
	MaxContributors int    `json:"max_contributors"`
	AnalyzeDays     int    `json:"analyze_days"`
}

// PageComparisonResult represents comparison between page versions
type PageComparisonResult struct {
	PageTitle      string                 `json:"page_title"`
	Languages      []string               `json:"languages"`
	Profiles       map[string]PageProfile `json:"profiles"`
	Discrepancies  []Discrepancy          `json:"discrepancies"`
	SuspicionLevel string                 `json:"suspicion_level"`
}

// Discrepancy represents differences between language versions
type Discrepancy struct {
	Type        string    `json:"type"`
	Language1   string    `json:"language1"`
	Language2   string    `json:"language2"`
	Description string    `json:"description"`
	Severity    string    `json:"severity"`
	DetectedAt  time.Time `json:"detected_at"`
}
