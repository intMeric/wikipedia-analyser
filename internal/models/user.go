// internal/models/user.go
package models

import (
	"time"
)

type UserProfile struct {
	Username         string                `json:"username"`
	UserID           int                   `json:"user_id"`
	RegistrationDate *time.Time            `json:"registration_date"`
	EditCount        int                   `json:"edit_count"`
	Groups           []string              `json:"groups"`
	ImplicitGroups   []string              `json:"implicit_groups"`
	RightsInfo       []string              `json:"rights_info"`
	BlockInfo        *BlockInfo            `json:"block_info,omitempty"`
	RecentContribs   []Contribution        `json:"recent_contributions"`
	TopPages         []PageEditSummary     `json:"top_edited_pages"`
	ActivityStats    ActivityStats         `json:"activity_stats"`
	RevokedContribs  []RevokedContribution `json:"revoked_contributions"`
	RevokedCount     int                   `json:"revoked_count"`
	RevokedRatio     float64               `json:"revoked_ratio"`
	RevertedByUsers  map[string]int        `json:"reverted_by_users"`
	SuspicionScore   int                   `json:"suspicion_score"`
	SuspicionFlags   []string              `json:"suspicion_flags"`
	Language         string                `json:"language"`
	RetrievedAt      time.Time             `json:"retrieved_at"`
}

type BlockInfo struct {
	Blocked    bool      `json:"blocked"`
	BlockedBy  string    `json:"blocked_by,omitempty"`
	BlockStart time.Time `json:"block_start,omitempty"`
	BlockEnd   time.Time `json:"block_end,omitempty"`
	Reason     string    `json:"reason,omitempty"`
}

type Contribution struct {
	RevID        int       `json:"rev_id"`
	PageTitle    string    `json:"page_title"`
	Namespace    int       `json:"namespace"`
	Timestamp    time.Time `json:"timestamp"`
	Comment      string    `json:"comment"`
	SizeDiff     int       `json:"size_diff"`
	IsMinor      bool      `json:"is_minor"`
	IsTop        bool      `json:"is_top"`
	PageID       int       `json:"page_id"`
	IsRevoked    bool      `json:"is_revoked"`
	RevokedBy    string    `json:"revoked_by,omitempty"`
	RevokedAt    time.Time `json:"revoked_at,omitempty"`
	RevertReason string    `json:"revert_reason,omitempty"`
}

type RevokedContribution struct {
	OriginalContrib Contribution `json:"original_contribution"`
	RevokedBy       string       `json:"revoked_by"`
	RevokedAt       time.Time    `json:"revoked_at"`
	RevertComment   string       `json:"revert_comment"`
	PageTitle       string       `json:"page_title"`
	RevertType      string       `json:"revert_type"` // "undo", "revert", "rollback", etc.
}

type PageEditSummary struct {
	PageTitle     string    `json:"page_title"`
	PageID        int       `json:"page_id"`
	Namespace     int       `json:"namespace"`
	EditCount     int       `json:"edit_count"`
	FirstEdit     time.Time `json:"first_edit"`
	LastEdit      time.Time `json:"last_edit"`
	TotalSizeDiff int       `json:"total_size_diff"`
}

type ActivityStats struct {
	DaysActive         int             `json:"days_active"`
	AverageEditsPerDay float64         `json:"average_edits_per_day"`
	LongestStreak      int             `json:"longest_streak_days"`
	MostActiveHour     int             `json:"most_active_hour"`
	MostActiveDay      string          `json:"most_active_day"`
	NamespaceDistrib   map[string]int  `json:"namespace_distribution"`
	RecentActivity     []DailyActivity `json:"recent_activity"`
}

type DailyActivity struct {
	Date      time.Time `json:"date"`
	EditCount int       `json:"edit_count"`
}

type WikiUserInfo struct {
	UserID         int      `json:"userid"`
	Name           string   `json:"name"`
	Registration   string   `json:"registration"`
	EditCount      int      `json:"editcount"`
	Groups         []string `json:"groups"`
	ImplicitGroups []string `json:"implicitgroups"`
	Rights         []string `json:"rights"`
	BlockExpiry    string   `json:"blockexpiry,omitempty"`
	BlockReason    string   `json:"blockreason,omitempty"`
	BlockedBy      string   `json:"blockedby,omitempty"`
}

type WikiContribution struct {
	UserID    int      `json:"userid"`
	User      string   `json:"user"`
	PageID    int      `json:"pageid"`
	RevID     int      `json:"revid"`
	ParentID  int      `json:"parentid"`
	NS        int      `json:"ns"`
	Title     string   `json:"title"`
	Timestamp string   `json:"timestamp"`
	Comment   string   `json:"comment"`
	Size      int      `json:"size"`
	SizeDiff  int      `json:"sizediff"`
	Minor     string   `json:"minor,omitempty"`
	Top       string   `json:"top,omitempty"`
	Tags      []string `json:"tags,omitempty"`
}

type WikiResponse struct {
	Query struct {
		Users        []WikiUserInfo     `json:"users,omitempty"`
		UserContribs []WikiContribution `json:"usercontribs,omitempty"`
	} `json:"query"`
	Continue struct {
		UCContinue string `json:"uccontinue,omitempty"`
		Continue   string `json:"continue,omitempty"`
	} `json:"continue,omitempty"`
}
