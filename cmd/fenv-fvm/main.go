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
	case "install":
		handleInstall(os.Args[2:])
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

	// 2. Read requested version
	ver, _, err := version.ReadVersion(projectRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	// 3. Check fvm availability
	if err := fvm.CheckFvmAvailable(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	// 4. Synchronize with fvm
	if err := fvm.Prepare(ver, projectRoot); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	// 5. Resolve executable
	binaryPath, err := fvm.ResolveBinary(projectRoot, binaryName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	// 6. Exec - replace current process with the resolved binary
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

// handleLocal implements "fenv-fvm local [version]"
func handleLocal(args []string) {
	// Check if fvm is available
	if err := fvm.CheckFvmAvailable(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	if len(args) > 0 {
		// Case: fenv-fvm local <version>
		// Write version to current directory and synchronize
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

		// Prepare Flutter via fvm
		if err := fvm.Prepare(newVersion, cwd); err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}

		// Output version
		fmt.Println(newVersion)
	} else {
		// Case: fenv-fvm local (no args)
		// Find project root and synchronize
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

		// Prepare Flutter via fvm
		if err := fvm.Prepare(ver, projectRoot); err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}

		// Output version and path
		fmt.Printf("%s (set by %s)\n", ver, versionFilePath)
	}
}

// handleInstall implements "fenv-fvm install <version>"
func handleInstall(args []string) {
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "fenv-fvm: install requires a version argument\n")
		os.Exit(1)
	}

	// Check if fvm is available
	if err := fvm.CheckFvmAvailable(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	versionToInstall := args[0]

	// Run fvm install only (no fvm use)
	if err := fvm.Install(versionToInstall); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
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
	fmt.Fprintf(os.Stderr, "  init               Setup shims and PATH\n")
	fmt.Fprintf(os.Stderr, "  local [version]    Set or sync Flutter version\n")
	fmt.Fprintf(os.Stderr, "  install <version>  Pre-download Flutter version\n")
	fmt.Fprintf(os.Stderr, "  version            Show current Flutter version\n")
}

// determineProgramMode returns the execution mode based on program name
func determineProgramMode(programName string) string {
	switch programName {
	case "fenv-fvm":
		return "cli"
	case "flutter", "dart":
		return "shim"
	default:
		return "unknown"
	}
}

// parseCommand parses the command from CLI arguments
func parseCommand(args []string) string {
	if len(args) < 2 {
		return ""
	}
	return args[1]
}

// validateInstallArgs validates arguments for install command
func validateInstallArgs(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("install requires a version argument")
	}
	return nil
}