# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Language Policy

**All project content must be written in English**, including:
- Documentation (README.md, SPEC.md, comments, etc.)
- Code comments
- Git commit messages
- Variable and function names
- Error messages

This ensures consistency and accessibility for the global development community.

## Project Overview

fenv-fvm is a single-binary CLI tool written in Go that provides FENV-compatible functionality for Flutter version management by directly referencing Flutter SDKs from FVM's cache directory. The project enables CI environments (like Codemagic) to build repositories that use fenv locally without requiring Dart/Flutter runtimes or fvm command execution.

## Key Architecture

The tool operates in two execution modes determined by `argv[0]`:
- **CLI mode**: When executed as `fenv-fvm` - provides subcommands (`init`, `local`, `version`)
- **Shim mode**: When executed as `flutter` or `dart` - acts as a transparent proxy to the appropriate SDK binaries

### Core Components

1. **Version Resolution**: Reads `.flutter-version` files to determine the requested Flutter SDK version
2. **Project Root Discovery**: Searches upward from current directory to find `.flutter-version` file
3. **FVM Cache Discovery**: Locates FVM's cache directory in priority order:
   - `$FVM_CACHE_PATH` environment variable
   - `$HOME/fvm/versions`
   - `$HOME/Library/Application Support/fvm/versions`
4. **Binary Resolution**: Resolves SDK binaries via `<cache>/<version>/bin/{flutter,dart}`
5. **Shim System**: Creates symlinks/copies in `$FENV_FVM_ROOT/shims/` for transparent binary execution

## Project Structure

The project follows Go's standard project layout:

```
fenv-fvm/
├── cmd/
│   └── fenv-fvm/          # Main application entry point
│       └── main.go
├── internal/              # Internal packages (future extensions)
├── go.mod                 # Go module definition
├── .mise.toml            # Go version management with mise
├── .gitignore            # Git ignore rules
├── README.md             # Basic project description
├── SPEC.md               # Complete technical specification
└── CLAUDE.md             # This file
```

## Development Setup

This project uses [mise](https://github.com/jdx/mise) for Go version management.

```bash
# Install and use the specified Go version
mise install
mise use

# Verify Go version
go version
```

## Development Commands

```bash
# Build the binary
go build -o fenv-fvm ./cmd/fenv-fvm

# Run tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests with coverage
go test -cover ./...

# Format code
go fmt ./...

# Static analysis
go vet ./...

# Build for multiple platforms
GOOS=linux GOARCH=amd64 go build -o fenv-fvm-linux-amd64 ./cmd/fenv-fvm
GOOS=darwin GOARCH=amd64 go build -o fenv-fvm-darwin-amd64 ./cmd/fenv-fvm
GOOS=darwin GOARCH=arm64 go build -o fenv-fvm-darwin-arm64 ./cmd/fenv-fvm
```

## Implementation Requirements

- **Language**: Must be implemented in Go (per specification)
- **Dependencies**: Single binary with no external runtime dependencies
- **Distribution**: Statically linked binary for Linux x86_64/aarch64 and macOS x86_64/arm64
- **Runtime Requirements**: Flutter SDKs must be pre-installed in FVM's cache directory

## Critical Behaviors

- **No Global Fallback**: `.flutter-version` file is mandatory - no global version support
- **No Command Execution**: All operations are filesystem-based - no fvm commands are executed
- **Process Replacement**: Uses `syscall.Exec` to replace current process with SDK binaries in shim mode
- **Filesystem-only Operations**: SDK resolution is purely based on directory and file existence checks
- **Error Handling**: Standardized error messages as defined in SPEC.md section 7

## Key Files

- `cmd/fenv-fvm/main.go`: Main application entry point
- `internal/fvm/fvm.go`: FVM cache directory discovery and SDK path resolution
- `internal/version/version.go`: Version file reading and project root discovery
- `internal/shim/shim.go`: Shim initialization and PATH setup
- `SPEC.md`: Complete technical specification (primary reference for implementation)
- `README.md`: User-facing documentation
- `.flutter-version`: Project-level Flutter version specification (created by users)

## Security Considerations

- Trusts that SDKs in FVM's cache directory are valid
- No checksum verification performed by fenv-fvm itself
- Does not download or modify SDKs - only references pre-installed ones