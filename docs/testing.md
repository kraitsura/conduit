# Testing Guide

This document describes the testing patterns and conventions used in the Conduit CLI project.

## Quick Start

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test ./... -v

# Run tests with race detector
go test ./... -race

# Run tests with coverage
go test ./... -cover

# Run specific package tests
go test ./internal/store/...

# Run specific test
go test ./internal/git -run TestIsAICommit

# Run benchmarks
go test ./... -bench=. -benchtime=1s
```

## Testing Patterns

### 1. Table-Driven Tests

All tests use table-driven patterns with named subtests for clarity:

```go
func TestMyFunction(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
        wantErr  bool
    }{
        {
            name:     "valid input",
            input:    "hello",
            expected: "HELLO",
        },
        {
            name:    "empty input returns error",
            input:   "",
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := MyFunction(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if got != tt.expected {
                t.Errorf("got %q, want %q", got, tt.expected)
            }
        })
    }
}
```

### 2. Test Utilities (`internal/testutil`)

The `testutil` package provides shared testing infrastructure.

#### Filesystem Helpers

```go
// Create temp directory (auto-cleaned up)
dir := testutil.TempDir(t)

// Create temp file with content
path := testutil.TempFile(t, dir, "config.json", `{"key": "value"}`)

// Create nested directories
testutil.MkdirAll(t, filepath.Join(dir, "a", "b", "c"))

// Create fake git repo for testing
repoPath := testutil.CreateGitRepo(t, dir, "my-project")
```

#### Builder Pattern

Builders provide a fluent API for creating test fixtures:

```go
// Project builder
project := testutil.NewProject().
    WithPath("/home/user/project").
    WithName("my-project").
    WithGitRemote("https://github.com/user/repo.git").
    WithLastActivity(time.Now()).
    Build()

// Agent builder
agent := testutil.NewAgent().
    WithPID(1234).
    WithType("claude").
    WithProjectPath("/project").
    Build()

// Commit builder
commit := testutil.NewCommit().
    WithHash("abc123").
    WithAuthor("John Doe").
    WithIsAI(false).
    Build()

// Activity builder
activity := testutil.NewActivity().
    WithType(types.ActivityCommit).
    WithProject("/project").
    Build()
```

#### Assertion Helpers

```go
testutil.AssertEqual(t, got, want, "values should match")
testutil.AssertTrue(t, condition, "condition should be true")
testutil.AssertFalse(t, condition, "condition should be false")
testutil.AssertNoError(t, err, "operation should succeed")
testutil.AssertError(t, err, "operation should fail")
testutil.AssertLen(t, slice, 3, "slice should have 3 elements")
```

#### Time Helpers

```go
weekAgo := testutil.DaysAgo(7)
hourAgo := testutil.HoursAgo(1)
recent := testutil.MinutesAgo(30)
```

### 3. Database Tests

For SQLite tests, create an isolated test database:

```go
func testDB(t *testing.T) *Store {
    t.Helper()

    tmpDir, err := os.MkdirTemp("", "conduit-test-*")
    if err != nil {
        t.Fatalf("failed to create temp dir: %v", err)
    }

    dbPath := filepath.Join(tmpDir, "test.db")
    store, err := Open(dbPath)
    if err != nil {
        os.RemoveAll(tmpDir)
        t.Fatalf("failed to open store: %v", err)
    }

    t.Cleanup(func() {
        store.Close()
        os.RemoveAll(tmpDir)
    })

    return store
}

func TestDatabaseOperation(t *testing.T) {
    store := testDB(t)
    // Use store for testing...
}
```

### 4. Benchmarks

Add benchmarks for performance-critical functions:

```go
func BenchmarkMyFunction(b *testing.B) {
    // Setup outside the timer
    data := prepareTestData()

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        MyFunction(data)
    }
}
```

## Test Organization

### File Naming

- Test files: `<source>_test.go` (e.g., `store_test.go`)
- Place in same package as source code
- Use `_test` suffix for test-only packages if needed

### Test Naming

- Test functions: `Test<FunctionName>` or `Test<Feature>`
- Subtests: descriptive snake_case names
- Benchmarks: `Benchmark<FunctionName>`

### Package Structure

```
internal/
├── agent/
│   ├── detect.go
│   └── detect_test.go      # Agent detection tests
├── config/
│   ├── config.go
│   └── config_test.go      # Configuration tests
├── daemon/
│   ├── daemon.go
│   └── daemon_test.go      # Daemon IPC, idle timeout, auto-start tests
├── git/
│   ├── git.go
│   └── git_test.go         # Git operations tests
├── project/
│   ├── discover.go
│   └── discover_test.go    # Project discovery tests
├── store/
│   ├── store.go
│   └── store_test.go       # SQLite persistence tests
├── testutil/
│   ├── testutil.go         # Test utilities
│   └── testutil_test.go    # Self-tests for utilities
└── types/
    ├── types.go
    └── types_test.go       # Type serialization tests
```

## Current Test Coverage

| Package | Coverage | Description |
|---------|----------|-------------|
| `store` | 89.5% | SQLite operations |
| `testutil` | 84.5% | Test utilities |
| `project` | 72.9% | Project discovery |
| `config` | 31.8% | Configuration |
| `agent` | 28.9% | Process detection |
| `git` | 6.4% | Git operations |
| `types` | n/a | Type definitions only |
| `daemon` | ~0%* | IPC protocol, idle timeout logic |

*Note: Daemon tests focus on types, protocol, and behavioral logic rather than executing the actual daemon (which requires process spawning and socket management).

## Test Inventory

### agent/detect_test.go

| Test | Description |
|------|-------------|
| `TestNewDetector` | Detector initialization with various patterns |
| `TestMatchAgentType` | Pattern matching against process commands |
| `TestMatchAgentTypeCaseInsensitive` | Case-insensitive matching verification |
| `TestExtractProcessName` | Process name extraction from commands |
| `TestGroupAgentsByProject` | Agent grouping by project path |
| `TestDetectorPatternEscaping` | Regex special character handling |
| `TestParseProcessOutput` | ps command output parsing |
| `BenchmarkMatchAgentType` | Pattern matching performance |
| `BenchmarkGroupAgentsByProject` | Grouping performance |

### config/config_test.go

| Test | Description |
|------|-------------|
| `TestDefaultConfig` | Default configuration values |
| `TestConfigDir` | Config directory path |
| `TestConfigPath` | Config file path |
| `TestLoadNonExistent` | Loading when file doesn't exist |
| `TestSaveAndLoad` | Config persistence roundtrip |
| `TestConfigJSONFormat` | JSON field names |
| `TestConfigPartialJSON` | Partial JSON updates |
| `TestEnsureDirectories` | Directory creation |
| `TestConfigValidation` | Config validation scenarios |
| `TestConfigPaths` | Platform path separators |
| `BenchmarkDefaultConfig` | Factory performance |
| `BenchmarkConfigMarshal` | JSON serialization performance |
| `BenchmarkConfigUnmarshal` | JSON deserialization performance |

### daemon/daemon_test.go

| Test | Description |
|------|-------------|
| `TestIdleTimeoutConstant` | Idle timeout value verification |
| `TestRequestJSON` | Request type JSON serialization |
| `TestRequestOmitempty` | Empty field omission in Request |
| `TestResponseJSON` | Response type JSON serialization |
| `TestResponseOmitempty` | Empty field omission in Response |
| `TestIsRunningWithNoSocket` | IsRunning behavior without socket |
| `TestSocketPathCreation` | Unix socket creation and cleanup |
| `TestDaemonRequestResponseProtocol` | Full IPC protocol roundtrip |
| `TestIdleTimeoutBehavior` | Auto-shutdown logic scenarios |
| `TestActivityTrackingScenarios` | Activity tracking for agents |
| `TestCommandHandlerRouting` | Command validation and routing |
| `TestDefaultActivityLimit` | Default limit for activities query |
| `BenchmarkRequestMarshal` | Request serialization performance |
| `BenchmarkResponseMarshal` | Response serialization performance |
| `BenchmarkResponseUnmarshal` | Response deserialization performance |

### git/git_test.go

| Test | Description |
|------|-------------|
| `TestIsAICommit` | AI commit detection patterns |
| `TestAIPatternsComprehensive` | All AI pattern coverage |
| `TestParseDiffStatLine` | git diff --stat parsing |
| `TestCommitLogParsing` | git log format parsing |
| `BenchmarkIsAICommit` | Detection performance |

### project/discover_test.go

| Test | Description |
|------|-------------|
| `TestIsSubPath` | Path containment detection |
| `TestSortByActivity` | Activity-based sorting |
| `TestFilterActive` | Active project filtering |
| `TestGetProjectByPath` | Project lookup by path |
| `TestGetProjectByName` | Project lookup by name |
| `TestDiscoverProjects` | Git repo discovery |
| `TestDiscoverProjectsSkipsHidden` | Hidden directory filtering |
| `TestDiscoverProjectsSkipsNonGit` | Non-git directory filtering |
| `TestDiscoverProjectsSkipsFiles` | File filtering |
| `TestDiscoverProjectsInvalidRoot` | Invalid path handling |
| `TestDetectCurrentProject` | Current project detection |
| `BenchmarkSortByActivity` | Sorting performance |
| `BenchmarkFilterActive` | Filtering performance |
| `BenchmarkIsSubPath` | Path check performance |

### store/store_test.go

| Test | Description |
|------|-------------|
| `TestOpen` | Database creation and opening |
| `TestSaveProject` | Project upsert operations |
| `TestGetProjects` | Project retrieval and ordering |
| `TestLogActivity` | Activity logging |
| `TestGetActivities` | Activity retrieval and filtering |
| `TestAgentSessions` | Agent session tracking |
| `TestGetProjectStats` | Stats aggregation |
| `TestActivityData` | JSON encoding helper |
| `TestStoreClose` | Database cleanup |
| `BenchmarkSaveProject` | Write performance |
| `BenchmarkGetProjects` | Read performance |
| `BenchmarkLogActivity` | Logging performance |

### testutil/testutil_test.go

| Test | Description |
|------|-------------|
| `TestTempDir` | Temp directory creation |
| `TestTempFile` | Temp file creation |
| `TestMkdirAll` | Directory tree creation |
| `TestCreateGitRepo` | Git repo scaffolding |
| `TestProjectBuilder` | Project builder API |
| `TestAgentBuilder` | Agent builder API |
| `TestCommitBuilder` | Commit builder API |
| `TestActivityBuilder` | Activity builder API |
| `TestAssertions` | Assertion helpers |
| `TestTimeHelpers` | Time utility functions |
| `TestBuilderImmutability` | Builder state isolation |
| `BenchmarkProjectBuilder` | Builder performance |
| `BenchmarkTempDir` | Temp dir overhead |

### types/types_test.go

| Test | Description |
|------|-------------|
| `TestProjectJSON` | Project serialization |
| `TestAgentJSON` | Agent serialization |
| `TestActivityJSON` | Activity serialization |
| `TestActivityTypeConstants` | Activity type values |
| `TestCommitJSON` | Commit serialization |
| `TestProjectStatsJSON` | Stats serialization |
| `TestDaemonStatusJSON` | Status serialization |
| `TestDaemonStatusOmitempty` | Omitempty behavior |
| `TestTypeZeroValues` | Zero value behavior |
| `BenchmarkProjectMarshal` | Serialization performance |
| `BenchmarkProjectUnmarshal` | Deserialization performance |

## Adding New Tests

1. Create test file alongside source: `myfile_test.go`
2. Use table-driven tests with `t.Run()` subtests
3. Use `testutil` helpers for fixtures and assertions
4. Add benchmarks for performance-critical code
5. Aim for descriptive test names that explain the scenario
6. Include edge cases and error conditions

## CI Integration

Recommended CI test commands:

```yaml
# Run all tests
- go test ./...

# Run with race detector
- go test ./... -race

# Run with coverage threshold
- go test ./... -coverprofile=coverage.out
- go tool cover -func=coverage.out

# Run benchmarks (optional)
- go test ./... -bench=. -benchtime=100ms -run='^$'
```
