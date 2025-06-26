## 📖 Commands

### User Analysis

```bash
# Basic user profile analysis
wikiosint user profile "Username" [options]

Options:
  --lang string              Wikipedia language (default "en")
  --output string            Output format: table, json, yaml (default "table")
  --save string              Save results to file
  -v, --verbose              Verbose output

  Revoked Contributions Analysis Options:
  --max-pages-analyze int    Maximum number of pages to analyze for revoked contributions (default 10)
  --max-revisions-page int   Maximum number of revisions to check per page for revoked contributions (default 50)
  --enable-deep-analysis     Enable thorough analysis for revoked contributions (slower but more accurate) (default false)
  --recent-days-only int     Only analyze revoked contributions from the last N days (default 90)
  --skip-revoked-analysis    Skip the entire revoked contributions analysis (default false)
```

### Page Analysis

```bash
# Comprehensive page analysis
wikiosint page analyze "Page Title" [options]

# Focus on edit history
wikiosint page history "Page Title" [options]

# Detect conflicts and edit wars
wikiosint page conflicts "Page Title" [options]

Options:
  --lang string              Wikipedia language (default "en")
  --output string            Output format: table, json, yaml (default "table")
  --save string              Save results to file
  --days int                 Number of days to analyze (default 30)
  --max-revisions int        Max revisions to analyze (default 100)
  --max-contributors int     Max contributors to analyze (default 20)
  --max-history int          Days of detailed history (default 30)
  --analyse-sources          Analyze page sources and references (default false)
```

### Cross-Page Analysis

```bash
# Analyze coordination patterns across multiple pages
wikiosint pages "Page 1" "Page 2" "Page 3" [options]

Options:
  --lang string              Wikipedia language (default "en")
  --output string            Output format: table, json, yaml (default "table")
  --save string              Save results to file
  --max-revisions int        Max revisions per page (default 200)
  --max-contributors int     Max contributors per page (default 50)
  --max-history int          Days of detailed history (default 90)
  --min-common-edits int     Min edits to be considered common contributor (default 3)
  --max-reaction-time int    Max minutes for suspicious reaction time (default 60)
  --min-support-ratio float Min ratio for mutual support detection (default 0.3)
  --enable-deep-analysis     Enable resource-intensive analysis (default false)
```

### Contribution Analysis

```bash
# Analyze a specific contribution/revision
wikiosint contribution analyze [revision_id] [page_title] [options]

# Analyze latest contribution to a page
wikiosint contribution analyze latest "Page Title" [options]

# Analyze recent contributions to a page
wikiosint contribution recent "Page Title" [options]

# Find suspicious contributions to a page
wikiosint contribution suspicious "Page Title" [options]

Options for 'analyze':
  --lang string              Wikipedia language (default "en")
  --output string            Output format: table, json, yaml (default "table")
  --save string              Save results to file
  --depth string             Analysis depth: basic, standard, deep (default "standard")
  --include-content          Include detailed content analysis (default true)
  --include-context          Include contextual analysis (default false, auto-enabled for deep)

Options for 'recent':
  --lang string              Wikipedia language (default "en")
  --output string            Output format: table, json, yaml (default "table")
  --save string              Save results to file
  --depth string             Analysis depth: basic, standard (default "basic")
  --limit int                Number of recent contributions to analyze (5-50) (default 10)

Options for 'suspicious':
  --lang string              Wikipedia language (default "en")
  --output string            Output format: table, json, yaml (default "table")
  --save string              Save results to file
  --threshold int            Minimum suspicion score threshold (0-100) (default 40)
  --days int                 Number of days to scan back (default 30)
  --limit int                Maximum suspicious contributions to show (default 20)
```

## 🎯 Use Cases

### Detect Suspicious Users

```bash
# Check if a user shows signs of coordinated manipulation
wikiosint user profile "Potential_Sockpuppet" --output json
```

### Analyze Controversial Pages

```bash
# Deep analysis of a potentially manipulated page
wikiosint page analyze "Controversial Topic" --max-revisions 500 --max-history 90

# Analyze page with source reliability checking
wikiosint page analyze "Scientific Article" --analyse-sources
```

### Cross-Page Coordination Detection

```bash
# Detect coordinated campaigns across related pages
wikiosint pages "Bitcoin" "Ethereum" "Cryptocurrency" --enable-deep-analysis

# Compare political topics for coordination
wikiosint pages "Politician A" "Politician B" "Election 2024" --max-history 180
```

### Monitor Recent Activity

```bash
# Focus on recent conflicts
wikiosint page conflicts "Current Events Page" --max-history 7
```

### Multi-language Investigation

```bash
# Compare activity across language editions
wikiosint page analyze "Sa
me Topic" --lang en
wikiosint page analyze "Same Topic" --lang fr
wikiosint page analyze "Same Topic" --lang de
```

### Source Reliability Analysis

```bash
# Comprehensive source analysis for academic topics
wikiosint page analyze "Climate Change" --analyse-sources --output json

# Quick source check for recent articles
wikiosint page analyze "Breaking News Topic" --analyse-sources --max-history 7

# Combined analysis: contributors and sources
wikiosint page analyze "Medical Article" --analyse-sources --max-contributors 50
```

### Contribution Analysis

```bash
# Analyze a specific suspicious edit
wikiosint contribution analyze 123456789 "Page Title" --depth deep

# Quick scan for recent problematic edits
wikiosint contribution recent "Controversial Page" --limit 20 --depth standard

# Find all suspicious activity on a page
wikiosint contribution suspicious "Target Page" --threshold 30 --days 60
```
