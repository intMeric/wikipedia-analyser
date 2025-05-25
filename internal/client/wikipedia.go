// internal/client/wikipedia.go
package client

import (
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/intMeric/wikipedia-analyser/internal/models"
	"github.com/tidwall/gjson"
)

const (
	defaultUserAgent = "WikiOSINT/1.0 (https://github.com/votre-username/wikiosint)"
	defaultTimeout   = 30 * time.Second
	maxRetries       = 3
)

// WikipediaClient encapsulates interactions with the MediaWiki API
type WikipediaClient struct {
	client   *resty.Client
	baseURL  string
	language string
}

// NewWikipediaClient creates a new client for the Wikipedia API
func NewWikipediaClient(language string) *WikipediaClient {
	client := resty.New()
	client.SetTimeout(defaultTimeout)
	client.SetRetryCount(maxRetries)
	client.SetRetryWaitTime(1 * time.Second)
	client.SetRetryMaxWaitTime(5 * time.Second)

	// User-Agent required by Wikipedia
	client.SetHeader("User-Agent", defaultUserAgent)

	baseURL := fmt.Sprintf("https://%s.wikipedia.org/w/api.php", language)

	return &WikipediaClient{
		client:   client,
		baseURL:  baseURL,
		language: language,
	}
}

// GetUserInfo retrieves basic user information
func (w *WikipediaClient) GetUserInfo(username string) (*models.WikiUserInfo, error) {
	params := map[string]string{
		"action":  "query",
		"list":    "users",
		"ususers": username,
		"usprop":  "blockinfo|groups|implicitgroups|rights|editcount|registration",
		"format":  "json",
	}

	resp, err := w.client.R().
		SetQueryParams(params).
		Get(w.baseURL)

	if err != nil {
		return nil, fmt.Errorf("API request error: %w", err)
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("non-200 API response: %d", resp.StatusCode())
	}

	// Parse with gjson for fast extraction
	body := string(resp.Body())
	users := gjson.Get(body, "query.users")

	if !users.Exists() || len(users.Array()) == 0 {
		return nil, fmt.Errorf("user not found: %s", username)
	}

	userInfo := users.Array()[0]

	// Check if user exists
	if gjson.Get(userInfo.String(), "missing").Exists() {
		return nil, fmt.Errorf("user not found: %s", username)
	}

	// Extract data
	wikiUser := &models.WikiUserInfo{
		UserID:       int(gjson.Get(userInfo.String(), "userid").Int()),
		Name:         gjson.Get(userInfo.String(), "name").String(),
		EditCount:    int(gjson.Get(userInfo.String(), "editcount").Int()),
		Registration: gjson.Get(userInfo.String(), "registration").String(),
	}

	// Extract groups
	groups := gjson.Get(userInfo.String(), "groups")
	for _, group := range groups.Array() {
		wikiUser.Groups = append(wikiUser.Groups, group.String())
	}

	// Extract implicit groups
	implicitGroups := gjson.Get(userInfo.String(), "implicitgroups")
	for _, group := range implicitGroups.Array() {
		wikiUser.ImplicitGroups = append(wikiUser.ImplicitGroups, group.String())
	}

	// Extract rights
	rights := gjson.Get(userInfo.String(), "rights")
	for _, right := range rights.Array() {
		wikiUser.Rights = append(wikiUser.Rights, right.String())
	}

	// Block info if present
	if gjson.Get(userInfo.String(), "blockexpiry").Exists() {
		wikiUser.BlockExpiry = gjson.Get(userInfo.String(), "blockexpiry").String()
		wikiUser.BlockReason = gjson.Get(userInfo.String(), "blockreason").String()
		wikiUser.BlockedBy = gjson.Get(userInfo.String(), "blockedby").String()
	}

	return wikiUser, nil
}

// GetUserContributions retrieves recent user contributions
func (w *WikipediaClient) GetUserContributions(username string, limit int) ([]models.WikiContribution, error) {
	params := map[string]string{
		"action":  "query",
		"list":    "usercontribs",
		"ucuser":  username,
		"uclimit": fmt.Sprintf("%d", limit),
		"ucprop":  "ids|title|timestamp|comment|size|sizediff|flags",
		"format":  "json",
	}

	resp, err := w.client.R().
		SetQueryParams(params).
		Get(w.baseURL)

	if err != nil {
		return nil, fmt.Errorf("API request error: %w", err)
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("non-200 API response: %d", resp.StatusCode())
	}

	body := string(resp.Body())
	contribs := gjson.Get(body, "query.usercontribs")

	if !contribs.Exists() {
		return []models.WikiContribution{}, nil
	}

	var contributions []models.WikiContribution
	for _, contrib := range contribs.Array() {
		contribution := models.WikiContribution{
			UserID:    int(gjson.Get(contrib.String(), "userid").Int()),
			User:      gjson.Get(contrib.String(), "user").String(),
			PageID:    int(gjson.Get(contrib.String(), "pageid").Int()),
			RevID:     int(gjson.Get(contrib.String(), "revid").Int()),
			ParentID:  int(gjson.Get(contrib.String(), "parentid").Int()),
			NS:        int(gjson.Get(contrib.String(), "ns").Int()),
			Title:     gjson.Get(contrib.String(), "title").String(),
			Timestamp: gjson.Get(contrib.String(), "timestamp").String(),
			Comment:   gjson.Get(contrib.String(), "comment").String(),
			Size:      int(gjson.Get(contrib.String(), "size").Int()),
			SizeDiff:  int(gjson.Get(contrib.String(), "sizediff").Int()),
		}

		// Optional flags
		if gjson.Get(contrib.String(), "minor").Exists() {
			contribution.Minor = "true"
		}
		if gjson.Get(contrib.String(), "top").Exists() {
			contribution.Top = "true"
		}

		contributions = append(contributions, contribution)
	}

	return contributions, nil
}

// GetUserEditsByNamespace retrieves edit statistics by namespace
func (w *WikipediaClient) GetUserEditsByNamespace(username string) (map[int]int, error) {
	// This query requires special privileges or extensions
	// For now, we use an approach based on recent contributions
	// In a future version, we could use external tools or Toolforge API

	contributions, err := w.GetUserContributions(username, 500) // Higher limit for analysis
	if err != nil {
		return nil, err
	}

	namespaceStats := make(map[int]int)
	for _, contrib := range contributions {
		namespaceStats[contrib.NS]++
	}

	return namespaceStats, nil
}

// SetUserAgent allows customizing the User-Agent
func (w *WikipediaClient) SetUserAgent(userAgent string) {
	w.client.SetHeader("User-Agent", userAgent)
}

// SetTimeout allows customizing the timeout
func (w *WikipediaClient) SetTimeout(timeout time.Duration) {
	w.client.SetTimeout(timeout)
}

// Language returns the configured language of the client
func (w *WikipediaClient) Language() string {
	return w.language
}

// GetPageInfo retrieves basic page information
func (w *WikipediaClient) GetPageInfo(title string) (*models.WikiPageInfo, error) {
	params := map[string]string{
		"action": "query",
		"titles": title,
		"prop":   "info",
		"format": "json",
	}

	resp, err := w.client.R().
		SetQueryParams(params).
		Get(w.baseURL)

	if err != nil {
		return nil, fmt.Errorf("API request error: %w", err)
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("non-200 API response: %d", resp.StatusCode())
	}

	// Parse with gjson for fast extraction
	body := string(resp.Body())
	pages := gjson.Get(body, "query.pages")

	if !pages.Exists() {
		return nil, fmt.Errorf("page not found: %s", title)
	}

	// Get the first (and only) page from the response
	var pageInfo models.WikiPageInfo
	pages.ForEach(func(key, value gjson.Result) bool {
		// Check if page exists
		if gjson.Get(value.String(), "missing").Exists() {
			return false // Page doesn't exist
		}

		pageInfo = models.WikiPageInfo{
			PageID:    int(gjson.Get(value.String(), "pageid").Int()),
			NS:        int(gjson.Get(value.String(), "ns").Int()),
			Title:     gjson.Get(value.String(), "title").String(),
			Touched:   gjson.Get(value.String(), "touched").String(),
			LastRevID: int(gjson.Get(value.String(), "lastrevid").Int()),
			Length:    int(gjson.Get(value.String(), "length").Int()),
		}
		return false // Break after first iteration
	})

	if pageInfo.PageID == 0 {
		return nil, fmt.Errorf("page not found: %s", title)
	}

	return &pageInfo, nil
}

// GetPageRevisions retrieves recent page revisions
func (w *WikipediaClient) GetPageRevisions(title string, limit int) ([]models.WikiRevision, error) {
	params := map[string]string{
		"action":  "query",
		"titles":  title,
		"prop":    "revisions",
		"rvlimit": fmt.Sprintf("%d", limit),
		"rvprop":  "ids|timestamp|user|userid|size|comment|flags",
		"format":  "json",
	}

	resp, err := w.client.R().
		SetQueryParams(params).
		Get(w.baseURL)

	if err != nil {
		return nil, fmt.Errorf("API request error: %w", err)
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("non-200 API response: %d", resp.StatusCode())
	}

	body := string(resp.Body())
	pages := gjson.Get(body, "query.pages")

	if !pages.Exists() {
		return []models.WikiRevision{}, nil
	}

	var revisions []models.WikiRevision
	pages.ForEach(func(key, value gjson.Result) bool {
		// Check if page exists
		if gjson.Get(value.String(), "missing").Exists() {
			return false
		}

		// Get revisions array
		revisionsArray := gjson.Get(value.String(), "revisions")
		for _, rev := range revisionsArray.Array() {
			revision := models.WikiRevision{
				RevID:     int(gjson.Get(rev.String(), "revid").Int()),
				ParentID:  int(gjson.Get(rev.String(), "parentid").Int()),
				User:      gjson.Get(rev.String(), "user").String(),
				Timestamp: gjson.Get(rev.String(), "timestamp").String(),
				Size:      int(gjson.Get(rev.String(), "size").Int()),
				Comment:   gjson.Get(rev.String(), "comment").String(),
			}

			// Optional fields
			if gjson.Get(rev.String(), "userid").Exists() {
				revision.UserID = int(gjson.Get(rev.String(), "userid").Int())
			}
			if gjson.Get(rev.String(), "minor").Exists() {
				revision.Minor = "true"
			}
			if gjson.Get(rev.String(), "anon").Exists() {
				revision.Anon = "true"
			}

			revisions = append(revisions, revision)
		}
		return false // Break after first page
	})

	return revisions, nil
}

// GetPageContributors retrieves top contributors to a page
func (w *WikipediaClient) GetPageContributors(title string, limit int) ([]models.WikiContributor, error) {
	// First get the page ID
	pageInfo, err := w.GetPageInfo(title)
	if err != nil {
		return nil, fmt.Errorf("unable to get page info: %w", err)
	}

	params := map[string]string{
		"action":         "query",
		"pageids":        fmt.Sprintf("%d", pageInfo.PageID),
		"prop":           "contributors",
		"pclimit":        fmt.Sprintf("%d", limit),
		"pcexcludegroup": "bot",
		"format":         "json",
	}

	resp, err := w.client.R().
		SetQueryParams(params).
		Get(w.baseURL)

	if err != nil {
		return nil, fmt.Errorf("API request error: %w", err)
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("non-200 API response: %d", resp.StatusCode())
	}

	body := string(resp.Body())
	pages := gjson.Get(body, "query.pages")

	if !pages.Exists() {
		return []models.WikiContributor{}, nil
	}

	var contributors []models.WikiContributor
	pages.ForEach(func(key, value gjson.Result) bool {
		// Get contributors array
		contributorsArray := gjson.Get(value.String(), "contributors")
		for _, contrib := range contributorsArray.Array() {
			contributor := models.WikiContributor{
				Name: gjson.Get(contrib.String(), "name").String(),
			}

			// Optional fields
			if gjson.Get(contrib.String(), "userid").Exists() {
				contributor.UserID = int(gjson.Get(contrib.String(), "userid").Int())
			}
			if gjson.Get(contrib.String(), "anon").Exists() {
				contributor.Anon = "true"
			}

			contributors = append(contributors, contributor)
		}
		return false // Break after first page
	})

	return contributors, nil
}

// GetPageHistory retrieves detailed edit history for analysis
func (w *WikipediaClient) GetPageHistory(title string, days int) ([]models.WikiRevision, error) {
	// Calculate start date
	startDate := time.Now().AddDate(0, 0, -days).Format("2006-01-02T15:04:05Z")

	params := map[string]string{
		"action":  "query",
		"titles":  title,
		"prop":    "revisions",
		"rvlimit": "500", // Maximum allowed
		"rvprop":  "ids|timestamp|user|userid|size|comment|flags",
		"rvstart": startDate,
		"rvdir":   "newer",
		"format":  "json",
	}

	resp, err := w.client.R().
		SetQueryParams(params).
		Get(w.baseURL)

	if err != nil {
		return nil, fmt.Errorf("API request error: %w", err)
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("non-200 API response: %d", resp.StatusCode())
	}

	body := string(resp.Body())
	pages := gjson.Get(body, "query.pages")

	if !pages.Exists() {
		return []models.WikiRevision{}, nil
	}

	var revisions []models.WikiRevision
	pages.ForEach(func(key, value gjson.Result) bool {
		// Check if page exists
		if gjson.Get(value.String(), "missing").Exists() {
			return false
		}

		// Get revisions array
		revisionsArray := gjson.Get(value.String(), "revisions")
		for _, rev := range revisionsArray.Array() {
			revision := models.WikiRevision{
				RevID:     int(gjson.Get(rev.String(), "revid").Int()),
				ParentID:  int(gjson.Get(rev.String(), "parentid").Int()),
				User:      gjson.Get(rev.String(), "user").String(),
				Timestamp: gjson.Get(rev.String(), "timestamp").String(),
				Size:      int(gjson.Get(rev.String(), "size").Int()),
				Comment:   gjson.Get(rev.String(), "comment").String(),
			}

			// Optional fields
			if gjson.Get(rev.String(), "userid").Exists() {
				revision.UserID = int(gjson.Get(rev.String(), "userid").Int())
			}
			if gjson.Get(rev.String(), "minor").Exists() {
				revision.Minor = "true"
			}
			if gjson.Get(rev.String(), "anon").Exists() {
				revision.Anon = "true"
			}

			revisions = append(revisions, revision)
		}
		return false // Break after first page
	})

	return revisions, nil
}
