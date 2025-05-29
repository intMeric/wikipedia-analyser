// internal/models/cross_page.go
package models

import (
	"time"
)

// CrossPageAnalysis represents the result of analyzing multiple pages together
type CrossPageAnalysis struct {
	Pages               []string                `json:"pages"`
	Language            string                  `json:"language"`
	TotalPages          int                     `json:"total_pages"`
	TotalContributors   int                     `json:"total_contributors"`
	CommonContributors  []CommonContributor     `json:"common_contributors"`
	CoordinatedPatterns CoordinatedPatterns     `json:"coordinated_patterns"`
	TemporalPatterns    TemporalPatterns        `json:"temporal_patterns"`
	SockpuppetNetworks  []SockpuppetNetwork     `json:"sockpuppet_networks"`
	SuspicionScore      int                     `json:"suspicion_score"`
	SuspicionFlags      []string                `json:"suspicion_flags"`
	AnalysisTimestamp   time.Time               `json:"analysis_timestamp"`
	PageProfiles        map[string]*PageProfile `json:"page_profiles"`
}

// CommonContributor represents a user who edited multiple pages
type CommonContributor struct {
	Username            string               `json:"username"`
	UserID              int                  `json:"user_id,omitempty"`
	PagesEdited         []string             `json:"pages_edited"`
	TotalEdits          int                  `json:"total_edits"`
	EditsByPage         map[string]int       `json:"edits_by_page"`
	FirstEdit           time.Time            `json:"first_edit"`
	LastEdit            time.Time            `json:"last_edit"`
	SuspicionScore      int                  `json:"suspicion_score"`
	SuspicionFlags      []string             `json:"suspicion_flags"`
	MutualSupportEvents []MutualSupportEvent `json:"mutual_support_events"`
	IsAnonymous         bool                 `json:"is_anonymous"`
}

// CoordinatedPatterns contains detected coordination patterns
type CoordinatedPatterns struct {
	MutualSupportPairs    []MutualSupportPair `json:"mutual_support_pairs"`
	TagTeamEditing        []TagTeamPattern    `json:"tag_team_editing"`
	CoordinatedReversions []CoordinatedRevert `json:"coordinated_reversions"`
	SupportNetworks       []SupportNetwork    `json:"support_networks"`
	CoordinationScore     float64             `json:"coordination_score"`
}

// MutualSupportPair represents two users who defend each other
type MutualSupportPair struct {
	UserA               string               `json:"user_a"`
	UserB               string               `json:"user_b"`
	SupportEvents       []MutualSupportEvent `json:"support_events"`
	MutualSupportRatio  float64              `json:"mutual_support_ratio"`
	AverageReactionTime int                  `json:"average_reaction_time_minutes"`
	ReciprocityScore    float64              `json:"reciprocity_score"`
	ExclusivityRatio    float64              `json:"exclusivity_ratio"`
	PagesInvolved       []string             `json:"pages_involved"`
	SuspicionLevel      string               `json:"suspicion_level"`
}

// MutualSupportEvent represents a single support event
type MutualSupportEvent struct {
	Timestamp     time.Time `json:"timestamp"`
	PageTitle     string    `json:"page_title"`
	SupportType   string    `json:"support_type"`   // "revert_defense", "content_restoration", "discussion_support"
	ReactionTime  int       `json:"reaction_time"`  // Minutes between attack and defense
	AttackerUser  string    `json:"attacker_user"`  // User who attacked/reverted
	DefenderUser  string    `json:"defender_user"`  // User who defended
	SupportedUser string    `json:"supported_user"` // User who was defended
	RevisionID    int       `json:"revision_id"`
	Comment       string    `json:"comment"`
}

// TagTeamPattern represents coordinated tag-team editing
type TagTeamPattern struct {
	Users            []string    `json:"users"`
	PagesAffected    []string    `json:"pages_affected"`
	EditSequences    []EditEvent `json:"edit_sequences"`
	RotationPattern  string      `json:"rotation_pattern"`
	AvoidanceScore   float64     `json:"avoidance_score"`   // How well they avoid 3RR
	CoordinationTime int         `json:"coordination_time"` // Minutes between handoffs
}

// EditEvent represents a single edit in a sequence
type EditEvent struct {
	Timestamp  time.Time `json:"timestamp"`
	Username   string    `json:"username"`
	PageTitle  string    `json:"page_title"`
	RevisionID int       `json:"revision_id"`
	SizeDiff   int       `json:"size_diff"`
	Comment    string    `json:"comment"`
	IsRevert   bool      `json:"is_revert"`
}

// CoordinatedRevert represents coordinated reversion activity
type CoordinatedRevert struct {
	TargetUser       string      `json:"target_user"`
	RevertingUsers   []string    `json:"reverting_users"`
	PagesAffected    []string    `json:"pages_affected"`
	RevertEvents     []EditEvent `json:"revert_events"`
	CoordinationTime int         `json:"coordination_time"`
	SuspicionLevel   string      `json:"suspicion_level"`
}

// SupportNetwork represents a network of users supporting each other
type SupportNetwork struct {
	NetworkID       string             `json:"network_id"`
	Users           []string           `json:"users"`
	SupportMatrix   map[string]float64 `json:"support_matrix"` // user_a->user_b: support_score
	NetworkDensity  float64            `json:"network_density"`
	CentralUsers    []string           `json:"central_users"`
	PagesControlled []string           `json:"pages_controlled"`
	NetworkScore    float64            `json:"network_score"`
}

// TemporalPatterns contains temporal coordination analysis
type TemporalPatterns struct {
	SynchronizedEditing   []SynchronizedEvent    `json:"synchronized_editing"`
	EditingWaves          []EditingWave          `json:"editing_waves"`
	TimeZonePatterns      []TimeZonePattern      `json:"timezone_patterns"`
	TemporalCorrelation   float64                `json:"temporal_correlation"`
	SuspiciousTimeWindows []SuspiciousTimeWindow `json:"suspicious_time_windows"`
}

// SynchronizedEvent represents synchronized editing activity
type SynchronizedEvent struct {
	Timestamp           time.Time `json:"timestamp"`
	Users               []string  `json:"users"`
	PagesAffected       []string  `json:"pages_affected"`
	TimeWindow          int       `json:"time_window_minutes"`
	SynchronizationType string    `json:"synchronization_type"` // "simultaneous", "sequential", "coordinated"
	SuspicionLevel      string    `json:"suspicion_level"`
}

// EditingWave represents a wave of coordinated editing
type EditingWave struct {
	StartTime     time.Time `json:"start_time"`
	EndTime       time.Time `json:"end_time"`
	PeakTime      time.Time `json:"peak_time"`
	Users         []string  `json:"users"`
	PagesAffected []string  `json:"pages_affected"`
	TotalEdits    int       `json:"total_edits"`
	WaveIntensity float64   `json:"wave_intensity"`
	TriggerEvent  string    `json:"trigger_event,omitempty"`
}

// TimeZonePattern represents suspicious timezone patterns
type TimeZonePattern struct {
	EstimatedTimeZone string   `json:"estimated_timezone"`
	Users             []string `json:"users"`
	Confidence        float64  `json:"confidence"`
	ActivityPeaks     []int    `json:"activity_peaks"` // Hours of day
	SuspiciousHours   []int    `json:"suspicious_hours"`
}

// SuspiciousTimeWindow represents a period of suspicious activity
type SuspiciousTimeWindow struct {
	StartTime       time.Time `json:"start_time"`
	EndTime         time.Time `json:"end_time"`
	ActivityType    string    `json:"activity_type"`
	Users           []string  `json:"users"`
	PagesAffected   []string  `json:"pages_affected"`
	SuspicionReason string    `json:"suspicion_reason"`
	SeverityLevel   string    `json:"severity_level"`
}

// SockpuppetNetwork represents a detected sockpuppet network
type SockpuppetNetwork struct {
	NetworkID             string              `json:"network_id"`
	MasterAccount         string              `json:"master_account,omitempty"`
	SuspectedSocks        []SockpuppetAccount `json:"suspected_socks"`
	SharedCharacteristics []string            `json:"shared_characteristics"`
	BehaviorPatterns      []BehaviorPattern   `json:"behavior_patterns"`
	PagesTargeted         []string            `json:"pages_targeted"`
	ConfidenceScore       float64             `json:"confidence_score"`
	DetectionReasons      []string            `json:"detection_reasons"`
	FirstDetected         time.Time           `json:"first_detected"`
	LastActivity          time.Time           `json:"last_activity"`
}

// SockpuppetAccount represents a suspected sockpuppet account
type SockpuppetAccount struct {
	Username          string    `json:"username"`
	UserID            int       `json:"user_id,omitempty"`
	CreationDate      time.Time `json:"creation_date,omitempty"`
	SuspicionScore    int       `json:"suspicion_score"`
	SuspicionReasons  []string  `json:"suspicion_reasons"`
	EditingPattern    string    `json:"editing_pattern"`
	ActivityTimeframe string    `json:"activity_timeframe"`
	PagesEdited       []string  `json:"pages_edited"`
	SimilarityScore   float64   `json:"similarity_score"`
}

// BehaviorPattern represents a pattern of suspicious behavior
type BehaviorPattern struct {
	PatternType   string    `json:"pattern_type"`
	Description   string    `json:"description"`
	Frequency     int       `json:"frequency"`
	Confidence    float64   `json:"confidence"`
	FirstObserved time.Time `json:"first_observed"`
	LastObserved  time.Time `json:"last_observed"`
	AffectedUsers []string  `json:"affected_users"`
	AffectedPages []string  `json:"affected_pages"`
}

// CrossPageAnalysisOptions contains options for cross-page analysis
type CrossPageAnalysisOptions struct {
	MaxRevisionsPerPage    int     `json:"max_revisions_per_page"`
	MaxContributorsPerPage int     `json:"max_contributors_per_page"`
	HistoryDays            int     `json:"history_days"`
	MinCommonEdits         int     `json:"min_common_edits"`         // Minimum edits to be considered common contributor
	MaxReactionTime        int     `json:"max_reaction_time"`        // Max minutes for support reaction to be suspicious
	MinMutualSupportRatio  float64 `json:"min_mutual_support_ratio"` // Min ratio for mutual support detection
	EnableDeepAnalysis     bool    `json:"enable_deep_analysis"`     // Enable resource-intensive analysis
}

// CrossPageAnalysisRequest represents a request for cross-page analysis
type CrossPageAnalysisRequest struct {
	Pages    []string                 `json:"pages"`
	Language string                   `json:"language"`
	Options  CrossPageAnalysisOptions `json:"options"`
}
