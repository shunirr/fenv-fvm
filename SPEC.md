# fenv-fvm Specification v1

This document defines all requirements necessary for the initial implementation of **fenv-fvm**. Implementers must write the program in **Go** and must not add behavior outside this specification.

## 1. Overview

### 1.1 Purpose

fenv-fvm is a single-binary CLI designed to reproduce the same workflow as **fenv** in CI environments. Its main purposes are:

- Manage Flutter SDK versions through `.flutter-version` files.
- Allow commands such as `flutter build ...`, `flutter pub get`, or `dart ...` to be executed without prefixes.
- Delegate Flutter SDK installation and management to **fvm**.
- Operate without dependencies on Dart or Flutter runtimes.
- Allow CI systems like Codemagic (where `fvm` is pre-installed) to build repositories that use `fenv` locally.

### 1.2 Typical CI Workflow

1. The repository includes a `.flutter-version` file.
2. The CI environment already has `fvm` installed.
3. The CI step downloads or includes the `fenv-fvm` binary.
4. Run `eval "$(fenv-fvm init)"` to initialize the PATH.
5. Run `fenv-fvm local` to synchronize `.flutter-version` with fvm.
6. Execute `flutter build ...` or `flutter pub get` as usual.

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

- `fvm` must be available on PATH at runtime.
- `fvm install` must have network access.
- Flutter and Dart SDKs are not required beforehand.

## 3. Core Model

### 3.1 Source of Truth

- The Flutter SDK version for a project is defined by the `.flutter-version` file.
- There is **no global fallback**. The `.flutter-version` file is mandatory.

### 3.2 Project Root Definition

- The “project root” is defined as the nearest ancestor directory (including the current one) that contains a `.flutter-version` file.
- All fvm commands (`use`, `.fvm` directory creation, SDK binary resolution) are based on this directory.

### 3.3 Requested Version

- The first line of `.flutter-version` defines the requested Flutter version.
- This string is passed directly to `fvm install <version>` and `fvm use <version>`.

  - Examples: `3.13.9`, `stable`.

### 3.4 fvm Synchronization

For the requested version, fenv-fvm performs:

1. `fvm install <version>` to download the SDK if not cached.
2. Execute `fvm use <version>` in the project root.

   - This creates `<project>/.fvm/flutter_sdk` pointing to the active SDK.

### 3.5 Binary Resolution

- Flutter binary: `<project>/.fvm/flutter_sdk/bin/flutter`
- Dart binary: `<project>/.fvm/flutter_sdk/bin/dart`
- fenv-fvm never parses fvm output or internal JSON; it only trusts `.fvm/flutter_sdk`.

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

3. **Check fvm Availability**

   - If `fvm` not found in PATH, exit with error.

4. **Synchronize**

   - Run `fvm install <version>` and `fvm use <version>` in the project root.
   - On failure, exit with error.

5. **Resolve Executable**

   - If `argv[0] == flutter`, use `<project>/.fvm/flutter_sdk/bin/flutter`.
   - If `argv[0] == dart`, use `<project>/.fvm/flutter_sdk/bin/dart`.
   - If missing, exit with error.

6. **Exec**

   - Replace the current process with the resolved binary using `syscall.Exec`.
   - Pass through all arguments and environment variables.
   - The resulting exit code is the underlying Flutter/Dart process code.

**Idempotency:** If the same version is already installed and active, fvm commands should complete quickly without redundant work.

## 6. CLI Mode Specification

Only the following subcommands are implemented: `init`, `local`, `install`, `version`.
No `global`, `which`, `list`, `list-remote`, `doctor`, or `workspace`.

### 6.1 `fenv-fvm init`

**Purpose:** Setup shims and PATH for use in CI or shell.

**Process:**

1. Determine `$FENV_FVM_ROOT` (default: `~/.fenv-fvm`).
2. Create `$FENV_FVM_ROOT` and `$FENV_FVM_ROOT/shims` if missing.
3. Create shims for `flutter` and `dart` pointing to the main binary.
4. Output the PATH setup snippet to stdout.

**Error:**

- On failure, print `fenv-fvm: failed to initialize shims directory` to stderr and exit non-zero.

### 6.2 `fenv-fvm local`

#### `fenv-fvm local <version>`

**Purpose:** Declare the Flutter SDK version for the current project and synchronize immediately.

**Steps:**

1. Treat current directory as project root (no ancestor search).
2. Write `<version>` to `.flutter-version`.
3. Execute:

   - `fvm install <version>`
   - `fvm use <version>`

4. On success: output `<version>` and exit code 0.

**Errors:**

- `fvm` not found → `fenv-fvm: fvm not found in PATH`
- install/use fails → `fenv-fvm: failed to prepare Flutter '<version>' via fvm`

#### `fenv-fvm local` (no args)

**Purpose:** Synchronize with existing `.flutter-version` (mainly for CI).

**Steps:**

1. Search for `.flutter-version` upward.
2. If not found → error.
3. Read the version.
4. Execute `fvm install` and `fvm use` in project root.
5. Output `<version> (set by <path>)`.

**Errors:**

- Missing `.flutter-version` → `fenv-fvm: no Flutter version configured (.flutter-version not found)`
- Unreadable `.flutter-version` → `fenv-fvm: failed to read .flutter-version`
- `fvm` missing → same as above
- fvm failure → `fenv-fvm: failed to prepare Flutter '<version>' via fvm`

### 6.3 `fenv-fvm install <version>`

**Purpose:** Pre-download the specified Flutter version using fvm only.
Does **not** modify `.flutter-version` or run `fvm use`.

**Errors:**

- `fvm` missing → error.
- fvm failure → `fenv-fvm: failed to install Flutter '<version>' via fvm`

### 6.4 `fenv-fvm version`

**Purpose:** Display the Flutter version resolved from `.flutter-version`.

**Output:**
`<version> (set by <path>)`

## 7. Error Messages and Exit Codes

**Standardized messages:**

1. `fenv-fvm: fvm not found in PATH`
2. `fenv-fvm: no Flutter version configured (.flutter-version not found)`
3. `fenv-fvm: failed to read .flutter-version`
4. `fenv-fvm: failed to prepare Flutter '<version>' via fvm`
5. `fenv-fvm: failed to install Flutter '<version>' via fvm`
6. `fenv-fvm: resolved Flutter SDK is incomplete (missing bin/flutter)`

Exit code is non-zero on any failure.

## 8. Implementation Requirements (Go)

### 8.1 Process Execution

- Use Go’s `os/exec` to run `fvm install` and `fvm use`.
- Set working directory to project root for `fvm use`.
- Pass through stdout/stderr. On failure, print the defined error messages.

### 8.2 Exec Replacement

- Use `syscall.Exec` (or equivalent) to replace the current process with the Flutter/Dart binary.
- On failure: `fenv-fvm: failed to exec resolved Flutter SDK binary`.

### 8.3 Path Handling

- Resolve all paths to absolute canonical paths.
- Verify binary existence before `exec`.
- If missing, return the standard “incomplete SDK” error.

### 8.4 Idempotency

- Repeated calls to `local` or shim execution must be safe and not cause redundant downloads.
- Re-running `fvm use` for the same version is acceptable.

### 8.5 Global State

- No global version files (e.g., `~/.fenv-fvm/version`).
- Only `.flutter-version` defines the active version.

## 9. Security and Distribution

### 9.1 Distribution

- Distributed as a single binary.
- CI users should verify integrity via checksum (e.g., SHA256).
- fenv-fvm itself does not perform checksum verification.

### 9.2 Trust Model

- fenv-fvm trusts fvm. SDK authenticity verification is out of scope.
- Network or certificate errors during fvm operations are not handled by fenv-fvm.

## 10. Finalized Specification

- Implementation language: **Go**.
- `.flutter-version` is the only version source.
- No global fallback.
- Shim structure: single binary + symlinks. Behavior switches via `argv[0]`.
- Shim mode: resolves version, runs `fvm install/use`, and execs SDK binaries.
- CLI subcommands: `init`, `local`, `install`, `version`.
- `local <version>` writes `.flutter-version` and synchronizes.
- `local` (no args) synchronizes from existing file.
- `install <version>` pre-downloads only.
- `version` prints `<version> (set by <path>)`.
- Errors and exit codes follow standardized messages.
- Windows is out of scope.

Implementation must strictly follow this specification.
