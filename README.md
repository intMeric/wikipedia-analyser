## 📖 Commands

### User Analysis

```bash
# Basic user profile analysis
wikiosint user profile "Username" [options]

Options:
  --lang string     Wikipedia language (default "en")
  --output string   Output format: table, json, yaml (default "table")
  --save string     Save results to file
  -v, --verbose     Verbose output
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
  --max-revisions int        Max revisions to analyze (default 100)
  --max-contributors int     Max contributors to analyze (default 20)
  --max-history int          Days of detailed history (default 30)
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
wikiosint page analyze "Same Topic" --lang en
wikiosint page analyze "Same Topic" --lang fr
wikiosint page analyze "Same Topic" --lang de
```
