# fenv-fvm

fenv-fvm is an [fenv](https://github.com/fenv-org/fenv)-compatible Flutter version management tool. It references Flutter SDKs from [FVM (Flutter Version Management)](https://fvm.app/)'s cache directory and enables the same workflow as fenv in CI environments.

## Features

- **Single Binary**: No Dart or Flutter runtime required
- **fenv-compatible**: Version management via `.flutter-version` files
- **FVM Cache Integration**: Directly references Flutter SDKs installed by FVM
- **No Command Execution**: Pure filesystem-based operations, no subprocess overhead
- **CI/CD Optimized**: Perfect for environments where Flutter SDKs are pre-installed via FVM
- **Multi-platform**: Supports Linux (x86_64/aarch64) and macOS (x86_64/arm64)

## Requirements

- Flutter SDKs must be pre-installed in FVM's cache directory (typically via `fvm install`)
- No runtime dependency on fvm commands

## Installation

### Homebrew (macOS/Linux)

The easiest way to install fenv-fvm is via Homebrew:

```bash
brew install shunirr/fenv-fvm/fenv-fvm
```

### Manual Installation from GitHub Releases

Alternatively, download the binary for your platform from the latest release:

```bash
# Example: macOS arm64
curl -L -o fenv-fvm.tar.gz https://github.com/shunirr/fenv-fvm/releases/latest/download/fenv-fvm-darwin-arm64.tar.gz
tar -xzf fenv-fvm.tar.gz
chmod +x fenv-fvm
sudo mv fenv-fvm /usr/local/bin/
```

Available binaries:
- `fenv-fvm-linux-amd64.tar.gz`
- `fenv-fvm-linux-arm64.tar.gz`
- `fenv-fvm-darwin-amd64.tar.gz` (Intel Mac)
- `fenv-fvm-darwin-arm64.tar.gz` (Apple Silicon)

## Setup

### 1. Install Flutter SDKs via FVM

First, install the Flutter SDK versions you need using FVM:

```bash
# Install fvm if not already installed
dart pub global activate fvm

# Install Flutter SDK versions
fvm install 3.13.9
fvm install stable
```

### 2. Initialize Shims

Set up fenv-fvm shims and add them to your PATH:

```bash
eval "$(fenv-fvm init)"
```

To make this permanent, add it to your shell configuration file (`~/.bashrc`, `~/.zshrc`, etc.):

```bash
echo 'eval "$(fenv-fvm init)"' >> ~/.zshrc
```

### 3. Specify Flutter Version in Your Project

Create a `.flutter-version` file in your project root:

```bash
# Using fenv-fvm local command (verifies SDK exists)
fenv-fvm local 3.13.9

# Or create manually
echo "3.13.9" > .flutter-version
```

### 4. Use Flutter Commands as Normal

From now on, normal `flutter` and `dart` commands will automatically run with the correct version:

```bash
flutter --version
flutter pub get
flutter build apk
dart --version
```

## Command Reference

### `fenv-fvm init`

Sets up the shim directory and outputs a shell script for PATH configuration.

```bash
eval "$(fenv-fvm init)"
```

### `fenv-fvm local <version>`

Creates a `.flutter-version` file in the current directory and verifies the specified Flutter version exists in FVM's cache.

```bash
fenv-fvm local 3.13.9
fenv-fvm local stable
```

**Note**: If the SDK is not installed, you'll receive an error message instructing you to run `fvm install <version>`.

### `fenv-fvm version`

Displays the Flutter version configured for the current project.

```bash
fenv-fvm version
# Output: 3.13.9 (set by /path/to/project/.flutter-version)
```

## Usage in CI/CD Environments

### Typical CI Configuration Example

```yaml
# Example: .github/workflows/build.yml
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Install FVM and Flutter SDKs
        run: |
          dart pub global activate fvm
          echo "$HOME/.pub-cache/bin" >> $GITHUB_PATH
          fvm install 3.13.9

      - name: Install fenv-fvm
        run: |
          curl -L -o fenv-fvm.tar.gz https://github.com/shunirr/fenv-fvm/releases/latest/download/fenv-fvm-linux-amd64.tar.gz
          tar -xzf fenv-fvm.tar.gz
          chmod +x fenv-fvm
          sudo mv fenv-fvm /usr/local/bin/

      - name: Setup Flutter
        run: eval "$(fenv-fvm init)"

      - name: Build
        run: |
          eval "$(fenv-fvm init)"
          flutter pub get
          flutter build apk
```

### Codemagic

Even simpler on Codemagic since FVM comes pre-installed:

```yaml
workflows:
  build:
    environment:
      flutter: fvm  # Use FVM to manage Flutter versions
    scripts:
      - name: Install Flutter SDK for fenv-fvm
        script: |
          # Read version from .flutter-version file
          VERSION=$(cat .flutter-version)
          fvm install $VERSION

      - name: Setup fenv-fvm
        script: |
          curl -L -o fenv-fvm.tar.gz https://github.com/shunirr/fenv-fvm/releases/latest/download/fenv-fvm-linux-amd64.tar.gz
          tar -xzf fenv-fvm.tar.gz
          chmod +x fenv-fvm
          export PATH="$PWD:$PATH"
          eval "$(fenv-fvm init)"

      - name: Build
        script: |
          eval "$(fenv-fvm init)"
          flutter build apk
```

## How It Works

fenv-fvm operates as follows:

1. Reads the Flutter version from the `.flutter-version` file
2. Locates FVM's cache directory in the following priority order:
   - `$FVM_CACHE_PATH` environment variable (if set)
   - `$HOME/fvm/versions` (common default)
   - `$HOME/Library/Application Support/fvm/versions` (official macOS default)
3. Resolves the path to `<cache>/<version>/bin/flutter` (or `dart`)
4. Replaces the current process with the resolved binary (`syscall.Exec`)

This allows normal `flutter`/`dart` commands to transparently execute with the appropriate version, without any subprocess overhead or infinite loop issues.

## Troubleshooting

### `Flutter SDK 'x.x.x' is not installed`

The requested Flutter SDK version is not installed in FVM's cache:

```bash
# Install the SDK using fvm
fvm install 3.13.9
```

### `fvm cache directory not found`

FVM's cache directory doesn't exist. Install FVM and at least one Flutter SDK:

```bash
# Install fvm
dart pub global activate fvm

# Install a Flutter SDK
fvm install stable
```

### `.flutter-version not found`

The `.flutter-version` file doesn't exist in the project root:

```bash
fenv-fvm local 3.13.9
```

## Architecture

fenv-fvm is designed to be completely independent of FVM commands:

- **No subprocess execution**: All operations are filesystem-based
- **Direct SDK access**: Reads directly from FVM's cache directory structure
- **Fast and reliable**: No command parsing or process management overhead
- **CI-friendly**: Works as long as SDKs are pre-installed, regardless of how

## License

MIT License

## Related Projects

- [fenv](https://github.com/fenv-org/fenv) - Original fenv
- [FVM](https://fvm.app/) - Flutter Version Management
