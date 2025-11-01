# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

fenv-fvm is a single-binary CLI tool written in Go that provides FENV-compatible functionality for Flutter version management using FVM as the backend. The project enables CI environments (like Codemagic) to build repositories that use fenv locally without requiring Dart/Flutter runtimes.

## Key Architecture

The tool operates in two execution modes determined by `argv[0]`:
- **CLI mode**: When executed as `fenv-fvm` - provides subcommands (`init`, `local`, `install`, `version`)
- **Shim mode**: When executed as `flutter` or `dart` - acts as a transparent proxy to the appropriate SDK binaries

### Core Components

1. **Version Resolution**: Reads `.flutter-version` files to determine the requested Flutter SDK version
2. **Project Root Discovery**: Searches upward from current directory to find `.flutter-version` file
3. **FVM Integration**: Delegates SDK management to `fvm install` and `fvm use` commands
4. **Binary Resolution**: Resolves SDK binaries via `<project>/.fvm/flutter_sdk/bin/{flutter,dart}`
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
- **Runtime Requirements**: `fvm` must be available in PATH

## Critical Behaviors

- **No Global Fallback**: `.flutter-version` file is mandatory - no global version support
- **Process Replacement**: Uses `syscall.Exec` to replace current process with SDK binaries in shim mode
- **Idempotency**: Multiple executions with same version should be safe and fast
- **Error Handling**: Standardized error messages as defined in SPEC.md section 7

## Key Files

- `cmd/fenv-fvm/main.go`: Main application entry point
- `internal/`: Internal packages for future code organization
- `SPEC.md`: Complete technical specification (primary reference for implementation)
- `README.md`: Basic project description
- `.flutter-version`: Project-level Flutter version specification (created by users)

## Security Considerations

- Trusts fvm for SDK authenticity verification
- No checksum verification performed by fenv-fvm itself
- Network and certificate errors during fvm operations are not handled