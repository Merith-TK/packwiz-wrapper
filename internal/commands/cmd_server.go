package commands

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/Merith-TK/packwiz-wrapper/internal/packwiz"
	"github.com/Merith-TK/packwiz-wrapper/internal/utils"
)

// CmdServer provides comprehensive server management functionality
func CmdServer() (names []string, shortHelp, longHelp string, execute func([]string) error) {
	return []string{"server", "test-server"},
		"Manage Minecraft test servers for the current pack",
		`Server Management Commands:
  pw server setup         - Download and deploy all server files
  pw server start         - Start the server (foreground)
  pw server stop          - Stop the running server (if managed by pw)
  pw server reset         - Delete and redeploy all server files
  pw server delete        - Delete all server files
  pw server status        - Show server status and information

Examples:
  pw server setup         - Set up server for the first time
  pw server start         - Start the configured server
  pw server reset         - Clean and reconfigure server
  pw server delete        - Remove all server files

The server command automatically detects pack versions and downloads
the appropriate server JAR and Java requirements.`,
		func(args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("server command requires a subcommand. Use 'pw help server' for available commands")
			}

			subcommand := args[0]
			subArgs := args[1:]

			switch subcommand {
			case "setup":
				return serverSetup(subArgs)
			case "start":
				return serverStart(subArgs)
			case "stop":
				return serverStop(subArgs)
			case "reset":
				return serverReset(subArgs)
			case "delete":
				return serverDelete(subArgs)
			case "status":
				return serverStatus(subArgs)
			default:
				return fmt.Errorf("unknown server subcommand: %s", subcommand)
			}
		}
}

// serverSetup downloads and deploys all server files
func serverSetup(args []string) error {
	packDir, _ := os.Getwd()
	
	// Find and parse pack.toml
	packToml, packLocation, err := loadPackConfig(packDir)
	if err != nil {
		return fmt.Errorf("failed to load pack configuration: %w", err)
	}

	// Determine Minecraft version
	mcVersion := getMinecraftVersion(packToml)
	if mcVersion == "" {
		return fmt.Errorf("could not determine Minecraft version from pack.toml")
	}

	fmt.Printf("Setting up server for Minecraft %s...\n", mcVersion)

	// Ensure Java is available (download if necessary)
	java, err := utils.EnsureJava(mcVersion)
	if err != nil {
		fmt.Printf("Java setup warning: %v\n", err)
		fmt.Println("Server may not start correctly without compatible Java")
		// Continue anyway - user might have Java in PATH
	} else {
		fmt.Printf("Using Java %s (version %d)\n", java.Version, java.Major)
	}

	// Create server run directory
	runDir := filepath.Join(packDir, ".run")
	if err := os.MkdirAll(runDir, 0755); err != nil {
		return fmt.Errorf("failed to create .run directory: %w", err)
	}

	// Create server configuration files
	if err := createServerConfig(runDir, packLocation); err != nil {
		return fmt.Errorf("failed to create server configuration: %w", err)
	}

	// Install mods using packwiz installer
	if err := installMods(runDir, packLocation); err != nil {
		return fmt.Errorf("failed to install mods: %w", err)
	}

	// Download appropriate server JAR
	if err := downloadServerJar(runDir, packToml, mcVersion); err != nil {
		return fmt.Errorf("failed to download server JAR: %w", err)
	}

	fmt.Println("✅ Server setup completed successfully!")
	fmt.Println("Use 'pw server start' to launch the server")
	return nil
}

// serverStart starts the configured server
func serverStart(args []string) error {
	packDir, _ := os.Getwd()
	runDir := filepath.Join(packDir, ".run")

	// Verify server is set up
	if err := verifyServerSetup(runDir); err != nil {
		return fmt.Errorf("server not properly set up: %w\nRun 'pw server setup' first", err)
	}

	// Load pack config for Java validation
	packToml, _, err := loadPackConfig(packDir)
	var javaCmd = "java" // Default fallback
	
	if err != nil {
		fmt.Printf("Warning: could not load pack config: %v\n", err)
	} else {
		mcVersion := getMinecraftVersion(packToml)
		if java, err := utils.FindCompatibleJava(mcVersion); err == nil {
			fmt.Printf("Using Java %s for Minecraft %s\n", java.Version, mcVersion)
			javaCmd = java.Path
		} else {
			fmt.Printf("Java warning: %v\n", err)
		}
	}

	fmt.Println("🚀 Starting Minecraft server...")
	fmt.Println("Press Ctrl+C to stop the server")

	// Start server with appropriate memory allocation
	cmd := exec.Command(javaCmd, "-Xmx2G", "-Xms1G", "-jar", "server.jar", "nogui")
	cmd.Dir = runDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}

// serverStop stops the running server (placeholder for future implementation)
func serverStop(args []string) error {
	fmt.Println("⚠️  Server stop functionality not yet implemented")
	fmt.Println("Use Ctrl+C in the server terminal to stop the server")
	return nil
}

// serverReset deletes and redeploys all server files
func serverReset(args []string) error {
	fmt.Println("🔄 Resetting server...")
	
	if err := serverDelete(args); err != nil {
		return fmt.Errorf("failed to delete server files: %w", err)
	}
	
	return serverSetup(args)
}

// serverDelete removes all server files
func serverDelete(args []string) error {
	packDir, _ := os.Getwd()
	runDir := filepath.Join(packDir, ".run")

	if _, err := os.Stat(runDir); os.IsNotExist(err) {
		fmt.Println("ℹ️  No server directory found")
		return nil
	}

	fmt.Println("🗑️  Deleting server files...")
	if err := os.RemoveAll(runDir); err != nil {
		return fmt.Errorf("failed to delete server directory: %w", err)
	}

	fmt.Println("✅ Server files deleted")
	return nil
}

// serverStatus shows current server status and information
func serverStatus(args []string) error {
	packDir, _ := os.Getwd()
	runDir := filepath.Join(packDir, ".run")

	fmt.Println("📊 Server Status")
	fmt.Println("================")

	// Check if server directory exists
	if _, err := os.Stat(runDir); os.IsNotExist(err) {
		fmt.Println("Status: ❌ Not set up")
	} else {
		fmt.Println("Status: ✅ Set up")
	}

	// Load pack information
	if packToml, _, err := loadPackConfig(packDir); err == nil {
		mcVersion := getMinecraftVersion(packToml)
		fmt.Printf("Minecraft Version: %s\n", mcVersion)
		
		if packToml.Versions.Fabric != "" {
			fmt.Printf("Fabric Version: %s\n", packToml.Versions.Fabric)
		}
		
		// Check Java compatibility
		if java, err := utils.FindCompatibleJava(mcVersion); err == nil {
			fmt.Printf("Java: %s (compatible)\n", java.Version)
		} else {
			fmt.Printf("Java: ❌ %v\n", err)
		}
	} else {
		fmt.Printf("Pack: ❌ %v\n", err)
		fmt.Println("Run 'pw server setup' from a directory containing pack.toml")
		return nil
	}

	// Only check server files if we have a server directory
	if _, err := os.Stat(runDir); err == nil {
		// Check server files
		serverJar := filepath.Join(runDir, "server.jar")
		if _, err := os.Stat(serverJar); err == nil {
			fmt.Println("Server JAR: ✅ Present")
		} else {
			fmt.Println("Server JAR: ❌ Missing")
		}

		eulaFile := filepath.Join(runDir, "eula.txt")
		if _, err := os.Stat(eulaFile); err == nil {
			fmt.Println("EULA: ✅ Accepted")
		} else {
			fmt.Println("EULA: ❌ Not accepted")
		}
	} else {
		fmt.Println("\nTo set up the server, run: pw server setup")
	}

	return nil
}

// Helper functions

func loadPackConfig(packDir string) (*packwiz.PackToml, string, error) {
	packLocation := utils.FindPackToml(packDir)
	if packLocation == "" {
		return nil, "", fmt.Errorf("pack.toml not found in current directory or parent directories")
	}

	packTomlPath := filepath.Join(packLocation, "pack.toml")
	data, err := os.ReadFile(packTomlPath)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read pack.toml: %w", err)
	}

	var packToml packwiz.PackToml
	if err := toml.Unmarshal(data, &packToml); err != nil {
		return nil, "", fmt.Errorf("failed to parse pack.toml: %w", err)
	}

	return &packToml, packLocation, nil
}

func getMinecraftVersion(packToml *packwiz.PackToml) string {
	// Prefer versions.minecraft, fallback to mc-version
	if packToml.Versions.Minecraft != "" {
		return packToml.Versions.Minecraft
	}
	return packToml.McVersion
}

func createServerConfig(runDir, packLocation string) error {
	// Create eula.txt
	eulaPath := filepath.Join(runDir, "eula.txt")
	if err := os.WriteFile(eulaPath, []byte("eula=true\n"), 0644); err != nil {
		return fmt.Errorf("failed to create eula.txt: %w", err)
	}
	fmt.Println("✅ Created eula.txt")

	// Copy server icon if available
	iconSrc := filepath.Join(packLocation, "icon.png")
	if _, err := os.Stat(iconSrc); err == nil {
		iconDst := filepath.Join(runDir, "server-icon.png")
		if err := copyFile(iconSrc, iconDst); err != nil {
			fmt.Printf("⚠️  Warning: failed to copy server icon: %v\n", err)
		} else {
			fmt.Println("✅ Copied server icon")
		}
	}

	return nil
}

func installMods(runDir, packLocation string) error {
	fmt.Println("📦 Installing mods using packwiz installer...")
	
	packTomlPath := filepath.Join(packLocation, "pack.toml")

	// Download packwiz installer if needed
	installerPath := filepath.Join(packLocation, "packwiz-installer-bootstrap.jar")
	if _, err := os.Stat(installerPath); os.IsNotExist(err) {
		fmt.Println("⬇️  Downloading packwiz installer...")
		if err := downloadPackwizInstaller(installerPath); err != nil {
			return fmt.Errorf("failed to download packwiz installer: %w", err)
		}
	}

	// Determine which Java to use
	javaCmd := "java" // Default fallback
	
	// Try to read pack.toml to get MC version for Java selection
	if data, err := os.ReadFile(packTomlPath); err == nil {
		var packToml packwiz.PackToml
		if err := toml.Unmarshal(data, &packToml); err == nil {
			mcVersion := getMinecraftVersion(&packToml)
			if java, err := utils.FindCompatibleJava(mcVersion); err == nil {
				javaCmd = java.Path
				fmt.Printf("Using Java %s for packwiz installer\n", java.Version)
			}
		}
	}

	// Run installer
	cmd := exec.Command(javaCmd, "-jar", installerPath, packTomlPath, "-s", "server")
	cmd.Dir = runDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("packwiz installer failed: %w", err)
	}

	fmt.Println("✅ Mods installed successfully")
	return nil
}

func downloadServerJar(runDir string, packToml *packwiz.PackToml, mcVersion string) error {
	serverJarPath := filepath.Join(runDir, "server.jar")
	
	// Skip if already exists
	if _, err := os.Stat(serverJarPath); err == nil {
		fmt.Println("✅ Server JAR already exists")
		return nil
	}

	fmt.Println("⬇️  Downloading server JAR...")

	// Determine server type and version
	if packToml.Versions.Fabric != "" {
		return downloadFabricServer(serverJarPath, mcVersion, packToml.Versions.Fabric)
	}
	
	// TODO: Add support for Forge, Quilt, Vanilla
	return fmt.Errorf("unsupported server type - only Fabric is currently supported")
}

func downloadFabricServer(serverJarPath, mcVersion, fabricVersion string) error {
	// Use Fabric API to get the appropriate server JAR
	url := fmt.Sprintf("https://meta.fabricmc.net/v2/versions/loader/%s/%s/1.1.0/server/jar", 
		mcVersion, fabricVersion)

	fmt.Printf("Downloading Fabric server for MC %s with Fabric %s...\n", mcVersion, fabricVersion)

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download server JAR: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download server JAR: HTTP %d", resp.StatusCode)
	}

	file, err := os.Create(serverJarPath)
	if err != nil {
		return fmt.Errorf("failed to create server JAR file: %w", err)
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write server JAR: %w", err)
	}

	fmt.Println("✅ Server JAR downloaded successfully")
	return nil
}

func verifyServerSetup(runDir string) error {
	// Check if server.jar exists
	serverJarPath := filepath.Join(runDir, "server.jar")
	if _, err := os.Stat(serverJarPath); os.IsNotExist(err) {
		return fmt.Errorf("server.jar not found")
	}

	// Check if eula.txt exists
	eulaPath := filepath.Join(runDir, "eula.txt")
	if _, err := os.Stat(eulaPath); os.IsNotExist(err) {
		return fmt.Errorf("eula.txt not found")
	}

	return nil
}

func downloadPackwizInstaller(path string) error {
	url := "https://github.com/packwiz/packwiz-installer/releases/latest/download/packwiz-installer-bootstrap.jar"

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	return err
}

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}
