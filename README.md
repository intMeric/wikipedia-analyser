# WikiOSINT 🔍

[![Go Version](https://img.shields.io/badge/go-1.21+-blue.svg)](https://golang.org)

Open Source Intelligence (OSINT) tool to analyze Wikipedia and detect manipulations, interference, and suspicious behavior.

## ✨ Features

- 🔍 **User Profile Analysis** - Detect suspicious user behavior patterns
- 📄 **Page Analysis** - Identify edit wars, conflicts, and coordinated editing
- 🌍 **Multi-language Support** - Works with all Wikipedia language editions
- 📊 **Multiple Output Formats** - Table, JSON, YAML export options
- 🚨 **Suspicion Scoring** - Automated detection of potential manipulation
- ⚔️ **Conflict Detection** - Identify edit wars and controversial pages

## 🚀 Quick Start

### Installation

```bash
# Install directly from GitHub
go install github.com/intMeric/wikipedia-analyser/cmd/wikiosint@latest

# Or clone and build
git clone github.com/intMeric/wikipedia-analyser.git
cd wikiosint
go build cmd/wikiosint/main.go
```

### Basic Usage

```bash
# Analyze a user
wikiosint user profile "Jimmy Wales"

# Analyze a Wikipedia page
wikiosint page analyze "Bitcoin"

# Use different language (French Wikipedia)
wikiosint page analyze "Napoleon" --lang fr

# Export results as JSON
wikiosint user profile "Suspicious_User" --output json --save results.json
```

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

## 🚨 Suspicion Indicators

### User-Level Indicators

- Recent account with intensive activity
- Currently blocked user
- Excessive focus on single pages
- Editing only in sensitive namespaces
- Frequent empty edit comments
- No special user groups despite high activity

### Page-Level Indicators

- High conflict ratio (many reversions)
- Few contributors for high edit volume
- Recent intensive editing activity
- Heavy anonymous editing
- Dominated by new editor accounts
- Low contributor diversity
- Recent editing conflicts

## 📊 Example Output

```
🔍 Analyzing Wikipedia page: Bitcoin
📊 Analysis parameters: 200 revisions, 30 contributors, 90 days history

╭─────────────────────────────────────────────────────────────╮
│  📄 WIKIPEDIA PAGE ANALYSIS: Bitcoin                        │
╰─────────────────────────────────────────────────────────────╯

🚨 Suspicion Score: MODERATE (45/100)

⚠️  SUSPICION INDICATORS
──────────────────────────────────────────────
🔸 Recent intensive editing activity
🔸 Low contributor diversity

👥 TOP CONTRIBUTORS ANALYSIS
────────────────────────────────────────────────────────────────────────────────
👤 CryptocurrencyExpert    89 edits   +15420 bytes 15/01/25 HIGH (72/100)
   📋 Recent account, active, High page activity

🚨 SUSPICIOUS CONTRIBUTORS DETECTED
──────────────────────────────────────────────
⚠️  CryptocurrencyExpert - HIGH (72/100)
   🔸 Recent account with high overall activity
   🔸 Unusually high activity on this page
```

## ⚙️ Configuration

Create optional config file at `~/.wikiosint.yaml`:

```yaml
default_language: "en"
default_output: "table"
api_timeout: "30s"
user_agent: "WikiOSINT/1.0 (your-email@example.com)"
```

## 🛠️ Development

### Prerequisites

- Go 1.21+
- Internet connection for Wikipedia API access

### Building from Source

```bash
git clone https://github.com/intMeric/wikipedia-analyser.git
cd wikiosint
go mod tidy
go build cmd/wikiosint/main.go
```

### Running Tests

```bash

go test ./...
```

### Project Structure

```
wikiosint/
├── cmd/wikiosint/     # CLI entry point
├── internal/
│   ├── cli/          # Command handlers
│   ├── client/       # Wikipedia API client
│   ├── models/       # Data structures
│   ├── analyzer/     # Analysis engines
│   └── formatter/    # Output formatting
└── docs/             # Documentation
```

## ⚠️ Disclaimer

This tool is intended for research and open source intelligence purposes. Use responsibly:

- ✅ Respect Wikipedia's terms of service and API limits
- ✅ Use for legitimate research and analysis
- ✅ Follow applicable laws and regulations
- ❌ Do not use for harassment or targeted attacks
- ❌ Do not use to circumvent Wikipedia policies

---
