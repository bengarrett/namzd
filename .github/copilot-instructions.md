# Copilot Instructions for namzd

## Project Overview
**namzd** is a Go CLI tool that quickly finds files by name or extension across directories and archives. It supports pattern matching, archive searching (ZIP, TAR), filtering by modification dates, and copying matched files.

## Architecture
The project uses a simple CLI architecture with three main components:

- **main.go**: Command-line interface using [kong](https://github.com/alecthomas/kong). Handles CLI parsing, version info, and delegates to the core logic.
- **ls package**: Core search logic (ls.go). The `Config` struct and `Walks` method orchestrate pattern matching and file enumeration using [fastwalk](https://github.com/charlievieth/fastwalk) for fast directory traversal.
- **cp package**: File copying utilities (cp.go). Handles destination validation and copying matched files.

All file matching is case-insensitive by default. Pattern matching supports glob-style wildcards (`*`, `?`).

## Build & Test Commands

### Build
```bash
task build          # Build binary (outputs: namzd on Unix, namzd.exe on Windows)
task buildr         # Build with race detection enabled
```

### Tests
```bash
task test           # Run full test suite once
task testr          # Run tests 1000x with race detection (intensive)
go test -v ./...    # Verbose test run
go test ./ls        # Test specific package
go test -run TestConfig_Copier ./ls   # Run single test
```

### Lint & Format
```bash
task lint           # Run all linters (gci, gofumpt, golangci-lint)
task nil            # Run nilaway static analysis for nil dereference detection
```

### Dependencies
```bash
task pkg-patch      # Update patch-level versions
task pkg-update     # Update to latest versions
go mod verify       # Verify module integrity
```

### Documentation
```bash
task doc            # Generate and browse pkgsite documentation at localhost:8090
```

### Release
```bash
task release        # Build release artifacts with goreleaser
```

## Key Conventions

### Testing Patterns
- Tests use subtests with `t.Run()` for organized test cases
- Subtests are often named with patterns like `CopyFile`, `CopyFile#01`, `CopyFile#02` (numbered variants)
- Test data lives in `../testdata` directory (see cp_test.go)
- Use `t.TempDir()` for temporary test directories to avoid cleanup issues

### Configuration Pattern
- Core logic is wrapped in a `Config` struct (see ls.Config)
- `Config` is initialized in main.go from CLI flags and passed to methods like `Walks()`
- All configuration comes from `Cmd` struct in main.go

### CLI Framework
- Uses `kong` for parameter parsing with structured tags
- CLI flags have descriptive `help:` tags
- Flags are organized into logical groups (`zip`, `copy`, `errs`)
- Single-letter short flags are used (e.g., `-c`, `-n`, `-a`)

### Error Handling
- Return errors wrapped with `fmt.Errorf()` for context
- Use two error handling modes: `--errors` (display) and `--panic` (exit on error)
- Commands support both Unix and Windows via platform-specific Task directives

### Code Quality
- golangci-lint is configured with multiple formatters enabled (gci, gofmt, gofumpt, goimports)
- Disabled linters: exhaustruct, nlreturn, noinlineerr, wsl, wsl_v5
- Test files excluded from some strict linting rules
- nilaway tool enabled for nil dereference detection (tool dependency in go.mod)

### Dependency Management
- Minimal dependencies: kong (CLI), charlievieth/fastwalk (fast directory walk)
- Go 1.24.6 required
- klauspost/compress (indirect, for archive support)

## Typical Workflow
1. Feature changes go in **ls/ls.go** or **cp/cp.go**
2. CLI interface changes go in **main.go** (Cmd struct and group definitions)
3. Add tests to the corresponding `*_test.go` file
4. Run `task lint` to auto-format and check for issues
5. Run `task test` to verify functionality
6. Use `task doc` to verify documentation renders correctly
