# fenv-fvm

fenv-fvm is an [fenv](https://github.com/fenv-org/fenv)-compatible Flutter version management tool. It uses [FVM (Flutter Version Management)](https://fvm.app/) as the backend and enables the same workflow as fenv in CI environments.

## Features

- **Single Binary**: No Dart or Flutter runtime required
- **fenv-compatible**: Version management via `.flutter-version` files
- **FVM Backend**: Leverages FVM's powerful SDK management capabilities
- **CI/CD Optimized**: Works immediately in fvm-ready CI environments like Codemagic
- **Multi-platform**: Supports Linux (x86_64/aarch64) and macOS (x86_64/arm64)

## Requirements

- `fvm` must be available in PATH
- Network connection (for downloading Flutter SDK)

## Installation

### Download from GitHub Releases

Download the binary for your platform from the latest release:

```bash
# Example: macOS arm64
curl -L -o fenv-fvm.tar.gz https://github.com/shunirr/fenv-fvm/releases/latest/download/fenv-fvm-darwin-arm64.tar.gz
tar -xzf fenv-fvm.tar.gz
chmod +x fenv-fvm
sudo mv fenv-fvm /usr/local/bin/
```

Available binaries:

- `fenv-fvm-linux-amd64.tar.gz`
- `fenv-fvm-linux-aarch64.tar.gz`
- `fenv-fvm-darwin-amd64.tar.gz` (Intel Mac)
- `fenv-fvm-darwin-arm64.tar.gz` (Apple Silicon)

## Setup

### 1. Initialize Shims

Set up fenv-fvm shims and add them to your PATH:

```bash
eval "$(fenv-fvm init)"
```

To make this permanent, add it to your shell configuration file (`~/.bashrc`, `~/.zshrc`, etc.):

```bash
echo 'eval "$(fenv-fvm init)"' >> ~/.zshrc
```

### 2. Specify Flutter Version in Your Project

Create a `.flutter-version` file in your project root:

```bash
# Method 1: Using fenv-fvm local command
fenv-fvm local 3.13.9

# Method 2: Create manually
echo "3.13.9" > .flutter-version
```

### 3. Use Flutter Commands as Normal

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

### `fenv-fvm local [version]`

#### With version argument

Creates a `.flutter-version` file in the current directory and installs the specified Flutter version.

```bash
fenv-fvm local 3.13.9
fenv-fvm local stable
```

#### Without version argument

Reads the existing `.flutter-version` file and synchronizes the SDK via FVM (primarily for CI).

```bash
fenv-fvm local
```

### `fenv-fvm install <version>`

Pre-downloads the specified Flutter version. Does not modify the `.flutter-version` file.

```bash
fenv-fvm install 3.13.9
```

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

      - name: Install FVM
        run: |
          dart pub global activate fvm
          echo "$HOME/.pub-cache/bin" >> $GITHUB_PATH

      - name: Install fenv-fvm
        run: |
          curl -L -o fenv-fvm.tar.gz https://github.com/YOUR_USERNAME/fenv-fvm/releases/latest/download/fenv-fvm-linux-amd64.tar.gz
          tar -xzf fenv-fvm.tar.gz
          chmod +x fenv-fvm
          sudo mv fenv-fvm /usr/local/bin/

      - name: Setup Flutter
        run: |
          eval "$(fenv-fvm init)"
          fenv-fvm local

      - name: Build
        run: |
          eval "$(fenv-fvm init)"
          flutter pub get
          flutter build apk
```

### Codemagic

Even simpler on Codemagic since fvm comes pre-installed:

```yaml
workflows:
  build:
    environment:
      flutter: stable
    scripts:
      - name: Setup fenv-fvm
        script: |
          curl -L -o fenv-fvm.tar.gz https://github.com/YOUR_USERNAME/fenv-fvm/releases/latest/download/fenv-fvm-linux-amd64.tar.gz
          tar -xzf fenv-fvm.tar.gz
          chmod +x fenv-fvm
          export PATH="$PWD:$PATH"
          eval "$(fenv-fvm init)"
          fenv-fvm local

      - name: Build
        script: |
          eval "$(fenv-fvm init)"
          flutter build apk
```

## How It Works

fenv-fvm operates as follows:

1. Reads the Flutter version from the `.flutter-version` file
2. Runs `fvm install <version>` and `fvm use <version>` to prepare the SDK
3. Resolves the path to `<project>/.fvm/flutter_sdk/bin/flutter` (or `dart`)
4. Replaces the current process with the resolved binary (`syscall.Exec`)

This allows normal `flutter`/`dart` commands to transparently execute with the appropriate version.

## Troubleshooting

### `fvm not found in PATH`

fvm is not installed or not in your PATH:

```bash
# Install fvm
dart pub global activate fvm

# Add to PATH
export PATH="$PATH:$HOME/.pub-cache/bin"
```

### `.flutter-version not found`

The `.flutter-version` file doesn't exist in the project root:

```bash
fenv-fvm local 3.13.9
```

## License

MIT License

## Related Projects

- [fenv](https://github.com/fenv-org/fenv) - Original fenv
- [FVM](https://fvm.app/) - Flutter Version Management
