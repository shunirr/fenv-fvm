package main

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	"fenv-fvm/internal/fvm"
	"fenv-fvm/internal/shim"
	"fenv-fvm/internal/version"
)

func main() {
	// Determine execution mode based on argv[0]
	programName := filepath.Base(os.Args[0])
	
	switch programName {
	case "fenv-fvm":
		runCLIMode()
	case "flutter", "dart":
		runShimMode(programName)
	default:
		fmt.Fprintf(os.Stderr, "fenv-fvm: unexpected program name '%s'\n", programName)
		os.Exit(1)
	}
}

// runCLIMode handles CLI mode when executed as "fenv-fvm"
func runCLIMode() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}
	
	command := os.Args[1]
	
	switch command {
	case "init":
		handleInit()
	case "local":
		handleLocal(os.Args[2:])
	case "version":
		handleVersion()
	default:
		fmt.Fprintf(os.Stderr, "fenv-fvm: unknown command '%s'\n", command)
		printUsage()
		os.Exit(1)
	}
}

// runShimMode handles shim mode when executed as "flutter" or "dart"
func runShimMode(binaryName string) {
	// 1. Locate project root
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "fenv-fvm: failed to get current directory\n")
		os.Exit(1)
	}

	projectRoot, err := version.FindProjectRoot(cwd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	// 2. Read version from .flutter-version
	ver, _, err := version.ReadVersion(projectRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	// 3. Resolve executable using version (filesystem-based, no fvm commands)
	binaryPath, err := fvm.ResolveBinary(ver, binaryName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	// 4. Exec - replace current process with the resolved binary
	// Prepare arguments (binary name + original args, excluding argv[0])
	args := append([]string{binaryName}, os.Args[1:]...)
	env := os.Environ()

	if err := syscall.Exec(binaryPath, args, env); err != nil {
		fmt.Fprintf(os.Stderr, "fenv-fvm: failed to exec resolved Flutter SDK binary: %v\n", err)
		os.Exit(1)
	}
}

// handleInit implements "fenv-fvm init"
func handleInit() {
	// Get current executable path
	exePath, err := os.Executable()
	if err != nil {
		fmt.Fprintf(os.Stderr, "fenv-fvm: failed to initialize shims directory\n")
		os.Exit(1)
	}

	// Initialize shims
	if err := shim.InitializeShims(exePath); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	// Print PATH setup snippet
	if err := shim.PrintPathSetup(); err != nil {
		fmt.Fprintf(os.Stderr, "fenv-fvm: failed to initialize shims directory\n")
		os.Exit(1)
	}
}

// handleLocal implements "fenv-fvm local <version>"
func handleLocal(args []string) {
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "fenv-fvm: local command requires a version argument\n")
		os.Exit(1)
	}

	// Case: fenv-fvm local <version>
	// Write version to current directory and verify SDK exists
	newVersion := args[0]

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "fenv-fvm: failed to get current directory\n")
		os.Exit(1)
	}

	// Write version file
	if err := version.WriteVersion(cwd, newVersion); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	// Verify SDK exists in fvm cache (does not call fvm commands)
	if err := fvm.VerifySDKExists(newVersion); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	// Output version
	fmt.Println(newVersion)
}

// handleVersion implements "fenv-fvm version"
func handleVersion() {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "fenv-fvm: failed to get current directory\n")
		os.Exit(1)
	}

	projectRoot, err := version.FindProjectRoot(cwd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	ver, versionFilePath, err := version.ReadVersion(projectRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	fmt.Printf("%s (set by %s)\n", ver, versionFilePath)
}

// printUsage prints usage information
func printUsage() {
	fmt.Fprintf(os.Stderr, "Usage: fenv-fvm <command> [args]\n\n")
	fmt.Fprintf(os.Stderr, "Commands:\n")
	fmt.Fprintf(os.Stderr, "  init             Setup shims and PATH\n")
	fmt.Fprintf(os.Stderr, "  local <version>  Set Flutter version for project\n")
	fmt.Fprintf(os.Stderr, "  version          Show current Flutter version\n")
}