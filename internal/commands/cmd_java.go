package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"

	"github.com/Merith-TK/packwiz-wrapper/internal/utils"
	"github.com/Merith-TK/packwiz-wrapper/internal/utils/java"
)

// CmdJava provides Java installation management
func CmdJava() (names []string, shortHelp, longHelp string, execute func([]string) error) {
	return []string{"java"},
		"Manage Java installations for Minecraft servers",
		`Java Management Commands:
  pw java list            - List all available Java installations
  pw java status          - Check Java compatibility for current pack
  pw java install <ver>   - Install specific Java version (8, 17, 21)
  pw java remove <ver>    - Remove managed Java installation
  pw java path <ver>      - Show path to managed Java installation

Java Passthrough:
  pw java <ver> [args...]  - Execute Java commands with specific version
  
Examples:
  pw java list            - Show system and managed Java versions
  pw java status          - Check what Java is needed for current pack
  pw java install 21      - Pre-install Java 21
  pw java path 17         - Get path to Java 17 executable
  
  pw java 21 -version     - Run java -version with managed Java 21
  pw java 17 -jar app.jar - Run JAR file with managed Java 17
  pw java 8 -Xmx2G MyApp   - Run application with Java 8 and 2GB memory

Managed Java installations are stored in the application data directory
and are automatically used by server commands when appropriate.`,
		func(args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("java command requires a subcommand. Use 'pw help java' for available commands")
			}

			subcommand := args[0]
			subArgs := args[1:]

			// Check if first argument is a version number for passthrough
			if isVersionNumber(subcommand) {
				return javaPassthrough(subcommand, subArgs)
			}

			// Handle regular subcommands
			switch subcommand {
			case "list":
				return javaList(subArgs)
			case "status":
				return javaStatus(subArgs)
			case "install":
				return javaInstall(subArgs)
			case "remove":
				return javaRemove(subArgs)
			case "path":
				return javaPath(subArgs)
			default:
				return fmt.Errorf("unknown java subcommand: %s\nUse 'pw help java' for available commands", subcommand)
			}
		}
}

// javaList shows all available Java installations
func javaList(args []string) error {
	fmt.Println("☕ Java Installations")
	fmt.Println("===================")

	installations, err := java.FindJavaInstallations()
	if err != nil {
		return fmt.Errorf("failed to find Java installations: %w", err)
	}

	if len(installations) == 0 {
		fmt.Println("No Java installations found")
		fmt.Println("\nYou can install Java automatically with:")
		fmt.Println("  pw java install 8   # For Minecraft ≤ 1.12.2")
		fmt.Println("  pw java install 17  # For Minecraft 1.13-1.20.4")
		fmt.Println("  pw java install 21  # For Minecraft ≥ 1.20.5")
		return nil
	}

	for i, java := range installations {
		marker := ""
		if i == 0 {
			marker = " (default)"
		}

		// Determine if this is managed or system Java
		source := "system"
		if isManagerJava(java.Path) {
			source = "managed"
		}

		fmt.Printf("Java %d: %s%s\n", java.Major, java.Version, marker)
		fmt.Printf("  Path: %s\n", java.Path)
		fmt.Printf("  Source: %s\n", source)
		fmt.Println()
	}

	return nil
}

// javaStatus checks Java compatibility for current pack
func javaStatus(args []string) error {
	fmt.Println("☕ Java Compatibility Status")
	fmt.Println("===========================")

	// Try to find pack.toml
	packDir, _ := os.Getwd()
	packToml, _, err := utils.LoadPackConfig(packDir)
	if err != nil {
		fmt.Printf("❌ No pack found: %v\n", err)
		fmt.Println("Run this command from a directory containing pack.toml")
		return nil
	}

	mcVersion := utils.GetMinecraftVersion(packToml)
	required := java.GetRequiredJavaVersion(mcVersion)
	strict := java.GetMinimumJavaVersion(mcVersion)

	fmt.Printf("Pack: %s\n", packToml.Name)
	fmt.Printf("Minecraft Version: %s\n", mcVersion)
	fmt.Printf("Required Java: %d (minimum: %d)\n", required, strict)
	fmt.Println()

	// Check current Java compatibility
	if javaVer, err := java.FindCompatibleJava(mcVersion); err == nil {
		fmt.Printf("✅ Compatible Java found: %s (version %d)\n", javaVer.Version, javaVer.Major)
		fmt.Printf("   Path: %s\n", javaVer.Path)
		if javaVer.Major == required {
			fmt.Println("   Status: Perfect match! 🎯")
		} else {
			fmt.Println("   Status: Compatible but not optimal")
		}
	} else {
		fmt.Printf("❌ No compatible Java found\n")
		fmt.Printf("   Problem: %v\n", err)
		fmt.Printf("   Solution: Run 'pw java install %d' to install Java %d\n", required, required)
	}

	return nil
}

// javaInstall installs a specific Java version
func javaInstall(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("java install requires a version number (8, 17, or 21)")
	}

	version := args[0]
	var majorVersion int

	switch version {
	case "8":
		majorVersion = 8
	case "17":
		majorVersion = 17
	case "21":
		majorVersion = 21
	default:
		return fmt.Errorf("unsupported Java version: %s (supported: 8, 17, 21)", version)
	}

	fmt.Printf("Installing Java %d...\n", majorVersion)

	// Check if already installed
	dataDir := java.GetDataDirectory()
	javaDir := filepath.Join(dataDir, "java", fmt.Sprintf("java-%d", majorVersion))
	javaExe := filepath.Join(javaDir, "bin", "java")
	if runtime.GOOS == "windows" {
		javaExe = filepath.Join(javaDir, "bin", "java.exe")
	}

	if _, err := os.Stat(javaExe); err == nil {
		// Get version info of existing installation
		javaVer, err := java.DetectJavaVersion(javaExe)
		if err == nil {
			fmt.Printf("✅ Java %d is already installed!\n", majorVersion)
			fmt.Printf("   Version: %s\n", javaVer.Version)
			fmt.Printf("   Path: %s\n", javaVer.Path)
			return nil
		}
	}

	// Force download and install the specific version
	_, err := java.InstallJava(majorVersion, nil)
	if err != nil {
		return fmt.Errorf("failed to install Java %d: %w", majorVersion, err)
	}

	// Detect the installed Java version
	javaVer, err := java.DetectJavaVersion(javaExe)
	if err != nil {
		return fmt.Errorf("failed to validate installed Java: %w", err)
	}

	fmt.Printf("✅ Java %d installed successfully!\n", majorVersion)
	fmt.Printf("   Version: %s\n", javaVer.Version)
	fmt.Printf("   Path: %s\n", javaVer.Path)

	return nil
}

// javaRemove removes a managed Java installation
func javaRemove(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("java remove requires a version number (8, 17, or 21)")
	}

	version := args[0]
	var majorVersion int

	switch version {
	case "8":
		majorVersion = 8
	case "17":
		majorVersion = 17
	case "21":
		majorVersion = 21
	default:
		return fmt.Errorf("unsupported Java version: %s (supported: 8, 17, 21)", version)
	}

	// Get managed Java directory
	dataDir := java.GetDataDirectory()
	javaDir := filepath.Join(dataDir, "java", fmt.Sprintf("java-%d", majorVersion))

	if _, err := os.Stat(javaDir); os.IsNotExist(err) {
		fmt.Printf("Java %d is not installed (no managed installation found)\n", majorVersion)
		return nil
	}

	fmt.Printf("Removing Java %d installation...\n", majorVersion)
	if err := os.RemoveAll(javaDir); err != nil {
		return fmt.Errorf("failed to remove Java %d: %w", majorVersion, err)
	}

	fmt.Printf("✅ Java %d removed successfully\n", majorVersion)
	return nil
}

// javaPath shows the path to a managed Java installation
func javaPath(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("java path requires a version number (8, 17, or 21)")
	}

	version := args[0]
	var majorVersion int

	switch version {
	case "8":
		majorVersion = 8
	case "17":
		majorVersion = 17
	case "21":
		majorVersion = 21
	default:
		return fmt.Errorf("unsupported Java version: %s (supported: 8, 17, 21)", version)
	}

	// Get managed Java directory
	dataDir := java.GetDataDirectory()
	javaDir := filepath.Join(dataDir, "java", fmt.Sprintf("java-%d", majorVersion))

	javaExe := filepath.Join(javaDir, "bin", "java")
	if runtime.GOOS == "windows" {
		javaExe = filepath.Join(javaDir, "bin", "java.exe")
	}

	if _, err := os.Stat(javaExe); os.IsNotExist(err) {
		return fmt.Errorf("Java %d is not installed (run 'pw java install %d' first)", majorVersion, majorVersion)
	}

	fmt.Println(javaExe)
	return nil
}

// Helper functions

func isManagerJava(javaPath string) bool {
	dataDir := java.GetDataDirectory()
	managedJavaDir := filepath.Join(dataDir, "java")
	return filepath.HasPrefix(javaPath, managedJavaDir)
}

// isVersionNumber checks if the argument is a Java version number (8, 11, 17, 21, etc.)
func isVersionNumber(arg string) bool {
	_, err := strconv.Atoi(arg)
	return err == nil
}

// getJavaExecutablePath returns the path to the Java executable for the specified version
func getJavaExecutablePath(version string) (string, error) {
	var majorVersion int

	switch version {
	case "8":
		majorVersion = 8
	case "17":
		majorVersion = 17
	case "21":
		majorVersion = 21
	default:
		return "", fmt.Errorf("unsupported Java version: %s (supported: 8, 17, 21)", version)
	}

	// Get managed Java directory
	dataDir := java.GetDataDirectory()
	javaDir := filepath.Join(dataDir, "java", fmt.Sprintf("java-%d", majorVersion))

	javaExe := filepath.Join(javaDir, "bin", "java")
	if runtime.GOOS == "windows" {
		javaExe = filepath.Join(javaDir, "bin", "java.exe")
	}

	if _, err := os.Stat(javaExe); os.IsNotExist(err) {
		return "", fmt.Errorf("Java %d is not installed. Use 'pw java install %d' to install it", majorVersion, majorVersion)
	}

	return javaExe, nil
}

// javaPassthrough executes the provided command using the specified Java version
func javaPassthrough(javaVersion string, args []string) error {
	// Get the Java executable path for the specified version
	javaPath, err := getJavaExecutablePath(javaVersion)
	if err != nil {
		return err
	}

	// Prepare the command with java executable and remaining arguments
	cmd := exec.Command(javaPath, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	// Execute the command
	return cmd.Run()
}
