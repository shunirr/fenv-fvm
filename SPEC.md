# fenv-fvm Specification v2

This document defines all requirements necessary for the implementation of **fenv-fvm**. Implementers must write the program in **Go** and must not add behavior outside this specification.

## 1. Overview

### 1.1 Purpose

fenv-fvm is a single-binary CLI designed to reproduce the same workflow as **fenv** in CI environments. Its main purposes are:

- Manage Flutter SDK versions through `.flutter-version` files.
- Allow commands such as `flutter build ...`, `flutter pub get`, or `dart ...` to be executed without prefixes.
- Reference Flutter SDKs directly from **fvm's cache directory** without calling fvm commands.
- Operate without dependencies on Dart, Flutter runtimes, or fvm commands.
- Allow CI systems like Codemagic (where Flutter SDKs are pre-installed via `fvm`) to build repositories that use `fenv` locally.

### 1.2 Typical CI Workflow

1. The repository includes a `.flutter-version` file.
2. The CI environment already has Flutter SDKs installed via `fvm install`.
3. The CI step downloads or includes the `fenv-fvm` binary.
4. Run `eval "$(fenv-fvm init)"` to initialize the PATH.
5. Execute `flutter build ...` or `flutter pub get` as usual (shims automatically resolve SDK).

## 2. Distribution and Build

### 2.1 Implementation Language

- fenv-fvm **must be implemented in Go**.
- Implementations in other languages or wrappers are not allowed.

### 2.2 Binary Distribution

- Distributed as a single binary named `fenv-fvm`.
- Must be statically linked where possible.
- The binary must not depend on external runtimes (e.g., Dart or Flutter).

### 2.3 Supported Platforms

- Linux x86_64
- Linux aarch64
- macOS x86_64
- macOS arm64

Windows is out of scope for the initial release.

### 2.4 Dependencies

- Flutter SDKs must be pre-installed in fvm's cache directory (typically via `fvm install`).
- No runtime dependencies on `fvm`, Dart, or Flutter executables.
- fenv-fvm only reads from fvm's cache directory structure.

## 3. Core Model

### 3.1 Source of Truth

- The Flutter SDK version for a project is defined by the `.flutter-version` file.
- There is **no global fallback**. The `.flutter-version` file is mandatory.

### 3.2 Project Root Definition

- The "project root" is defined as the nearest ancestor directory (including the current one) that contains a `.flutter-version` file.
- All SDK binary resolution is based on this directory.

### 3.3 Requested Version

- The first line of `.flutter-version` defines the requested Flutter version.
- This string is used to locate the SDK in fvm's cache directory.

  - Examples: `3.13.9`, `stable`.

### 3.4 FVM Cache Directory Resolution

fenv-fvm locates fvm's cache directory in the following priority order:

1. `$FVM_CACHE_PATH` environment variable (if set)
2. `$HOME/fvm/versions` (common default)
3. `$HOME/Library/Application Support/fvm/versions` (official default for macOS)

The cache directory contains subdirectories named by version (e.g., `3.13.9/`), each containing a complete Flutter SDK.

### 3.5 Binary Resolution

For a requested version from `.flutter-version`:

- Locate SDK directory: `<cache>/<version>/`
- Flutter binary: `<cache>/<version>/bin/flutter`
- Dart binary: `<cache>/<version>/bin/dart`

If the SDK directory or binaries do not exist, fenv-fvm reports an error instructing the user to run `fvm install <version>`.

## 4. Execution Modes

fenv-fvm determines its mode based on `argv[0]`:

- **CLI mode:** when `argv[0]` == `fenv-fvm`
- **Shim mode:** when `argv[0]` == `flutter` or `dart`

### 4.1 Shim Installation

- `fenv-fvm init` creates `$FENV_FVM_ROOT/shims/flutter` and `$FENV_FVM_ROOT/shims/dart` (default root: `~/.fenv-fvm`).
- Each shim is a **symbolic link, hard link, or copy** of the main `fenv-fvm` binary.
- The command also outputs a shell snippet to prepend the shim directory to PATH:

  ```sh
  export FENV_FVM_ROOT="$HOME/.fenv-fvm"
  export PATH="$FENV_FVM_ROOT/shims:$PATH"
  ```

- Users run `eval "$(fenv-fvm init)"` to activate shims.

## 5. Shim Mode Specification

When executed as `flutter` or `dart`, fenv-fvm performs:

1. **Locate Project Root**

   - Search upward from the current directory for `.flutter-version`.
   - If not found, exit with error.

2. **Read Requested Version**

   - Read `.flutter-version`. If unreadable, exit with error.

3. **Resolve FVM Cache Directory**

   - Check `$FVM_CACHE_PATH`, then `$HOME/fvm/versions`, then `$HOME/Library/Application Support/fvm/versions`.
   - Use the first existing directory.

4. **Resolve SDK Path**

   - Construct SDK path: `<cache>/<version>/`
   - If SDK directory does not exist, exit with error instructing user to run `fvm install <version>`.

5. **Resolve Executable**

   - If `argv[0] == flutter`, use `<cache>/<version>/bin/flutter`.
   - If `argv[0] == dart`, use `<cache>/<version>/bin/dart`.
   - If binary does not exist, exit with error.

6. **Exec**

   - Replace the current process with the resolved binary using `syscall.Exec`.
   - Pass through all arguments and environment variables.
   - The resulting exit code is the underlying Flutter/Dart process code.

**No fvm commands are executed** in shim mode. SDK resolution is purely filesystem-based.

## 6. CLI Mode Specification

Only the following subcommands are implemented: `init`, `local`, `version`.
No `global`, `which`, `list`, `list-remote`, `doctor`, `install`, or `workspace`.

### 6.1 `fenv-fvm init`

**Purpose:** Setup shims and PATH for use in CI or shell.

**Process:**

1. Determine `$FENV_FVM_ROOT` (default: `~/.fenv-fvm`).
2. Create `$FENV_FVM_ROOT` and `$FENV_FVM_ROOT/shims` if missing.
3. Create shims for `flutter` and `dart` pointing to the main binary.
4. Output the PATH setup snippet to stdout.

**Error:**

- On failure, print `fenv-fvm: failed to initialize shims directory` to stderr and exit non-zero.

### 6.2 `fenv-fvm local <version>`

**Purpose:** Declare the Flutter SDK version for the current project.

**Steps:**

1. Treat current directory as project root (no ancestor search).
2. Write `<version>` to `.flutter-version`.
3. Verify SDK exists in fvm cache:
   - Resolve fvm cache directory
   - Check if `<cache>/<version>/` exists
   - If not, output instruction to run `fvm install <version>`
4. On success: output `<version>` and exit code 0.

**Errors:**

- SDK not found → `fenv-fvm: Flutter SDK '<version>' is not installed\nPlease run: fvm install <version>`

### 6.3 `fenv-fvm version`

**Purpose:** Display the Flutter version resolved from `.flutter-version`.

**Steps:**

1. Search for `.flutter-version` upward from current directory.
2. If not found → error.
3. Read the version.
4. Output `<version> (set by <path>)`.

**Errors:**

- Missing `.flutter-version` → `fenv-fvm: no Flutter version configured (.flutter-version not found)`
- Unreadable `.flutter-version` → `fenv-fvm: failed to read .flutter-version`

## 7. Error Messages and Exit Codes

**Standardized messages:**

1. `fenv-fvm: no Flutter version configured (.flutter-version not found)`
2. `fenv-fvm: failed to read .flutter-version`
3. `fenv-fvm: Flutter SDK '<version>' is not installed` (followed by instruction: `Please run: fvm install <version>`)
4. `fenv-fvm: fvm cache directory not found`
5. `fenv-fvm: resolved Flutter SDK is incomplete (missing bin/flutter or bin/dart)`
6. `fenv-fvm: failed to initialize shims directory`
7. `fenv-fvm: failed to exec resolved Flutter SDK binary`

Exit code is non-zero on any failure.

## 8. Implementation Requirements (Go)

### 8.1 FVM Cache Discovery

- Check environment variable `$FVM_CACHE_PATH`.
- If not set, try `$HOME/fvm/versions`.
- If not found, try `$HOME/Library/Application Support/fvm/versions`.
- Use the first directory that exists.
- If none exist, return error: `fenv-fvm: fvm cache directory not found`.

### 8.2 SDK Path Resolution

- Given version string from `.flutter-version`, construct: `<cache>/<version>/`.
- Check if directory exists.
- Verify `<cache>/<version>/bin/flutter` and `<cache>/<version>/bin/dart` exist.
- If directory missing, return: `fenv-fvm: Flutter SDK '<version>' is not installed`.
- If binaries missing, return: `fenv-fvm: resolved Flutter SDK is incomplete (missing bin/flutter or bin/dart)`.

### 8.3 Exec Replacement

- Use `syscall.Exec` (or equivalent) to replace the current process with the Flutter/Dart binary.
- On failure: `fenv-fvm: failed to exec resolved Flutter SDK binary`.

### 8.4 Path Handling

- Resolve all paths to absolute canonical paths.
- Verify binary existence before `exec`.

### 8.5 No External Command Execution

- **fenv-fvm must not execute any fvm commands**.
- All operations are filesystem-based: reading `.flutter-version`, checking directory existence, and executing SDK binaries.

### 8.6 Global State

- No global version files (e.g., `~/.fenv-fvm/version`).
- Only `.flutter-version` defines the active version.
- No `.fvm/` directory is created or managed by fenv-fvm.

## 9. Security and Distribution

### 9.1 Distribution

- Distributed as a single binary.
- CI users should verify integrity via checksum (e.g., SHA256).
- fenv-fvm itself does not perform checksum verification.

### 9.2 Trust Model

- fenv-fvm trusts that SDKs in fvm's cache directory are valid.
- SDK authenticity verification is out of scope.
- fenv-fvm does not download or modify SDKs; it only references pre-installed ones.

## 10. Finalized Specification

- Implementation language: **Go**.
- `.flutter-version` is the only version source.
- No global fallback.
- Shim structure: single binary + symlinks. Behavior switches via `argv[0]`.
- Shim mode: resolves version from `.flutter-version`, locates SDK in fvm cache, and execs SDK binaries directly.
- CLI subcommands: `init`, `local <version>`, `version`.
- `local <version>` writes `.flutter-version` and verifies SDK exists (does not call fvm).
- `version` prints `<version> (set by <path>)`.
- **No fvm commands are executed** by fenv-fvm.
- All SDK resolution is filesystem-based using fvm's cache directory structure.
- Errors and exit codes follow standardized messages.
- Windows is out of scope.

Implementation must strictly follow this specification.
