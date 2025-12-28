name: "OpenSearch Date-Based Indexing for Log Retention"
description: |

---

## Goal

**Feature Goal**: Enable automatic log retention by using date-based OpenSearch index names, allowing ISM policies to delete old indices.

**Deliverable**: Modified logger that creates daily indices (e.g., `signalmice-logs-2024-12-28`) instead of a single static index.

**Success Definition**:
- Logger creates indices with date suffix in format `{index}-YYYY-MM-DD`
- New configuration option to control date-based indexing (enabled by default)
- All existing tests pass, new tests cover date-based functionality
- Documentation updated with ISM policy example for 90-day retention

## User Persona

**Target User**: DevOps/Infrastructure engineer managing signalmice deployment

**Use Case**: Implementing log retention policy to automatically delete logs older than 90 days

**User Journey**:
1. Deploy signalmice with date-based indexing (default behavior)
2. Configure OpenSearch ISM policy with 90-day retention
3. ISM automatically deletes old daily indices

**Pain Points Addressed**:
- Logs accumulating indefinitely in OpenSearch
- Manual cleanup of old logs required
- Storage costs growing continuously

## Why

- Enables standard OpenSearch Index State Management (ISM) policies for automatic log retention
- Follows industry best practices for log management
- Allows configurable retention periods without application changes
- Improves storage management and cost control

## What

### Functional Requirements

1. Logger creates daily indices with format `{base-index}-YYYY-MM-DD`
2. New config option `OPENSEARCH_USE_DAILY_INDEX` (default: `true`)
3. Backward compatible - can disable date suffix via config
4. Date uses UTC timezone for consistency

### Success Criteria

- [ ] Logger creates index with date suffix when enabled
- [ ] Date format is `YYYY-MM-DD` in UTC
- [ ] Config option `OPENSEARCH_USE_DAILY_INDEX` controls behavior
- [ ] Setting to `false` uses static index name (backward compatible)
- [ ] All tests pass
- [ ] README documents new config and ISM policy example

## All Needed Context

### Documentation & References

```yaml
- file: internal/logger/logger.go
  why: Current logger implementation - needs date suffix logic
  pattern: sendToOpensearch method uses l.index directly
  gotcha: Index name must be valid (lowercase, no special chars except hyphen)

- file: internal/config/config.go
  why: Add new configuration option
  pattern: Follow existing getEnv pattern with defaults

- file: internal/logger/logger_test.go
  why: Add tests for date-based indexing
  pattern: Follow existing test patterns with mocks

- file: README.md
  why: Document new configuration and ISM policy
  section: Configuration table and new Log Retention section
```

### Current Codebase tree

```bash
signalmice/
├── cmd/signalmice/
│   └── main.go
├── internal/
│   ├── config/
│   │   ├── config.go
│   │   └── config_test.go
│   ├── logger/
│   │   ├── logger.go
│   │   └── logger_test.go
│   ├── redis/
│   │   ├── client.go
│   │   └── client_test.go
│   └── shutdown/
│       ├── shutdown.go
│       └── shutdown_test.go
├── PRPs/
│   └── features/
│       ├── prp-signalmice-core.md
│       └── prp-opensearch-date-based-indexing.md  # NEW
├── README.md
└── ...
```

### Desired Changes

```bash
internal/config/config.go          # ADD: OpensearchUseDailyIndex field
internal/config/config_test.go     # ADD: Test for new config field
internal/logger/logger.go          # MODIFY: Use date-based index name
internal/logger/logger_test.go     # ADD: Tests for date-based indexing
README.md                          # UPDATE: Config table + ISM policy section
```

## Implementation Blueprint

### Data models and structure

```go
// Config struct addition
type Config struct {
    // ... existing fields ...
    OpensearchUseDailyIndex bool  // NEW: Enable date-based index names
}

// Logger will compute index name dynamically
func (l *Logger) getIndexName() string {
    if l.useDailyIndex {
        return fmt.Sprintf("%s-%s", l.baseIndex, time.Now().UTC().Format("2006-01-02"))
    }
    return l.baseIndex
}
```

### Implementation Tasks (ordered by dependencies)

```yaml
Task 1: MODIFY internal/config/config.go
  - ADD: OpensearchUseDailyIndex bool field to Config struct
  - ADD: getEnvBool helper function for boolean env vars
  - LOAD: OPENSEARCH_USE_DAILY_INDEX with default "true"
  - NAMING: OpensearchUseDailyIndex (Go naming convention)

Task 2: MODIFY internal/config/config_test.go
  - ADD: Test for OpensearchUseDailyIndex default value (true)
  - ADD: Test for OpensearchUseDailyIndex when set to false
  - FOLLOW: Existing test patterns in file

Task 3: MODIFY internal/logger/logger.go
  - ADD: useDailyIndex bool and baseIndex string fields to Logger struct
  - ADD: getIndexName() method that returns date-suffixed index when enabled
  - MODIFY: NewLogger to store useDailyIndex from config
  - MODIFY: sendToOpensearch to use getIndexName() instead of l.index

Task 4: MODIFY internal/logger/logger_test.go
  - ADD: Test for getIndexName with daily index enabled
  - ADD: Test for getIndexName with daily index disabled
  - ADD: Test that index name format is correct (YYYY-MM-DD)
  - FOLLOW: Existing mock patterns

Task 5: MODIFY README.md
  - ADD: OPENSEARCH_USE_DAILY_INDEX to configuration table
  - ADD: New "Log Retention" section with ISM policy example
  - UPDATE: Opensearch index documentation to explain date-based naming
```

### Implementation Patterns & Key Details

```go
// Config pattern - add boolean env var helper
func getEnvBool(key string, defaultValue bool) bool {
    if value, exists := os.LookupEnv(key); exists {
        return value == "true" || value == "1" || value == "yes"
    }
    return defaultValue
}

// Logger pattern - dynamic index name
func (l *Logger) getIndexName() string {
    if l.useDailyIndex {
        return fmt.Sprintf("%s-%s", l.baseIndex, time.Now().UTC().Format("2006-01-02"))
    }
    return l.baseIndex
}

// CRITICAL: Use UTC for consistent index names across timezones
// CRITICAL: Format "2006-01-02" is Go's reference date for YYYY-MM-DD
```

## Validation Loop

### Level 1: Syntax & Style

```bash
go fmt ./...
go vet ./...
golangci-lint run
```

### Level 2: Unit Tests

```bash
go test ./internal/config/... -v
go test ./internal/logger/... -v
go test ./... -v
```

### Level 3: Integration Testing

```bash
# Build and verify
go build -o signalmice ./cmd/signalmice

# Run with date-based indexing (default)
OPENSEARCH_URL=http://localhost:9200 ./signalmice &
# Check OpenSearch for index with date suffix

# Run with date-based indexing disabled
OPENSEARCH_USE_DAILY_INDEX=false OPENSEARCH_URL=http://localhost:9200 ./signalmice &
# Check OpenSearch for static index name
```

## Final Validation Checklist

### Technical Validation
- [ ] All tests pass: `go test ./... -v`
- [ ] No linting errors: `golangci-lint run`
- [ ] No vet errors: `go vet ./...`

### Feature Validation
- [ ] Date-based index name generated correctly
- [ ] Static index name works when disabled
- [ ] UTC timezone used consistently
- [ ] Index name format valid for OpenSearch

### Documentation
- [ ] README updated with new configuration
- [ ] ISM policy example included
- [ ] Log retention section added

---

## Anti-Patterns to Avoid

- ❌ Don't use local timezone - always use UTC
- ❌ Don't hardcode date format - use Go's reference date
- ❌ Don't break backward compatibility - default should work like before (with date suffix, but configurable)
