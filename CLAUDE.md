# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

WikiOSINT is a Wikipedia OSINT (Open Source Intelligence) analysis tool built in Go. It analyzes Wikipedia pages, users, and contributions to detect potential manipulations, coordinated editing campaigns, and suspicious behavior patterns.

## Build and Development Commands

### Build the Application
```bash
go build -o wikiosint cmd/wikiosint/main.go
```

### Run the Application
```bash
# After building
./wikiosint [command] [args]

# Or run directly
go run cmd/wikiosint/main.go [command] [args]
```

### Basic Commands
- `wikiosint user profile "Username"` - Analyze user profile and contributions
- `wikiosint page analyze "Page Title"` - Comprehensive page analysis  
- `wikiosint pages "Page1" "Page2" "Page3"` - Cross-page coordination analysis
- `wikiosint contribution analyze [revid]` - Analyze specific contributions

### Development Commands
- `go mod tidy` - Clean up dependencies
- `go fmt ./...` - Format all Go files
- `go vet ./...` - Run Go static analysis
- `go build ./...` - Build all packages

## Architecture

### Core Structure
- **cmd/wikiosint/** - Main application entry point
- **internal/cli/** - Cobra CLI command definitions and handlers
- **internal/models/** - Data structures for Wikipedia entities
- **internal/client/** - Wikipedia API client implementation
- **internal/analyzer/** - Core analysis logic for detecting suspicious patterns
- **internal/formatter/** - Output formatting (table, JSON, YAML)
- **internal/utils/** - Shared utility functions

### Key Models
- **PageProfile** - Complete Wikipedia page analysis including revisions, contributors, conflicts
- **UserProfile** - User analysis with contribution patterns, revoked edits, suspicion scoring
- **Contribution** - Individual edit/revision with metadata and analysis flags
- **ConflictStats** - Edit war detection and controversy metrics
- **SourceAnalysis** - Reference reliability and dead link detection

### Analysis Components
- **Page Analyzer** - Detects edit wars, suspicious patterns, source reliability
- **User Analyzer** - Identifies sockpuppets, coordinated accounts, unusual behavior
- **Contribution Analyzer** - Tracks reverted edits and manipulation attempts
- **Cross-Page Analyzer** - Finds coordination patterns across multiple pages

### API Integration
Uses MediaWiki API for data retrieval with support for:
- Multiple Wikipedia language editions
- Rate limiting and error handling
- Comprehensive revision history
- User contribution tracking
- Page metadata and statistics

## Configuration

The application uses Viper for configuration management:
- Default config file: `$HOME/.wikiosint.yaml`
- Environment variables supported
- CLI flags override config values
- Verbose mode available with `-v` flag

## Output Formats

Supports multiple output formats:
- **table** (default) - Human-readable console tables
- **json** - Machine-readable JSON
- **yaml** - YAML format
- File output with `--save filename` option