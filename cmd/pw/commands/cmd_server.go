package commands

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/Merith-TK/packwiz-wrapper/internal/packwiz"
)

// CmdServer provides test server functionality
func CmdServer() (names []string, shortHelp, longHelp string, execute func([]string) error) {
	return []string{"server", "test-server", "start"},
		"Start a test Minecraft server with the current pack",
		`Server Commands:
  pw server               - Start test server with current pack
  pw server setup         - Set up server directory without starting
  pw server start         - Start existing server
  pw server clean         - Clean server directory

Examples:
  pw server               - Quick start test server
  pw server setup         - Prepare server files only
  pw server clean         - Remove server files`,
		func(args []string) error {
			action := "start"
			if len(args) > 0 {
				action = args[0]
			}

			switch action {
			case "setup":
				return setupServer()
			case "start":
				return startServer()
			case "clean":
				return cleanServer()
			default:
				// Default behavior: setup and start
				if err := setupServer(); err != nil {
					return err
				}
				return startServer()
			}
		}
}

func setupServer() error {
	packDir, _ := os.Getwd()
	client := packwiz.NewClient(packDir)

	// Find pack directory
	packLocation := client.GetPackDir()
	if packLocation == "" {
		return fmt.Errorf("pack.toml not found")
	}

	// Create server run directory
	runDir := filepath.Join(packDir, ".run")
	if err := os.MkdirAll(runDir, 0755); err != nil {
		return fmt.Errorf("failed to create .run directory: %w", err)
	}

	fmt.Println("Setting up test server...")

	// Create eula.txt
	eulaPath := filepath.Join(runDir, "eula.txt")
	if err := os.WriteFile(eulaPath, []byte("eula=true\n"), 0644); err != nil {
		return fmt.Errorf("failed to create eula.txt: %w", err)
	}
	fmt.Println("Created eula.txt")

	// Copy server icon if available
	iconSrc := filepath.Join(packLocation, "icon.png")
	if _, err := os.Stat(iconSrc); err == nil {
		iconDst := filepath.Join(runDir, "server-icon.png")
		if err := copyFile(iconSrc, iconDst); err != nil {
			fmt.Printf("Warning: failed to copy server icon: %v\n", err)
		} else {
			fmt.Println("Copied server icon")
		}
	}

	// Run packwiz installer to install mods
	fmt.Println("Installing mods using packwiz installer...")
	packTomlPath := filepath.Join(packLocation, "pack.toml")

	// Check if packwiz-installer-bootstrap.jar exists
	installerPath := filepath.Join(packLocation, "packwiz-installer-bootstrap.jar")
	if _, err := os.Stat(installerPath); os.IsNotExist(err) {
		// Download packwiz installer if not found
		fmt.Println("Downloading packwiz installer...")
		if err := downloadPackwizInstaller(installerPath); err != nil {
			return fmt.Errorf("failed to download packwiz installer: %w", err)
		}
	}

	// Run installer
	cmd := exec.Command("java", "-jar", installerPath, packTomlPath, "-s", "server")
	cmd.Dir = runDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run packwiz installer: %w", err)
	}

	// Download server JAR if not present
	serverJarPath := filepath.Join(runDir, "server.jar")
	if _, err := os.Stat(serverJarPath); os.IsNotExist(err) {
		fmt.Println("Downloading Fabric server JAR...")
		if err := downloadServerJar(serverJarPath); err != nil {
			return fmt.Errorf("failed to download server JAR: %w", err)
		}
	}

	fmt.Println("Server setup completed!")
	return nil
}

func startServer() error {
	packDir, _ := os.Getwd()
	runDir := filepath.Join(packDir, ".run")

	// Check if server is set up
	serverJarPath := filepath.Join(runDir, "server.jar")
	if _, err := os.Stat(serverJarPath); os.IsNotExist(err) {
		return fmt.Errorf("server not set up, run 'pw server setup' first")
	}

	fmt.Println("Starting Minecraft server...")
	fmt.Println("Use Ctrl+C to stop the server")

	// Start server with reasonable memory allocation
	cmd := exec.Command("java", "-Xmx2G", "-Xms2G", "-jar", "server.jar", "nogui")
	cmd.Dir = runDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}

func cleanServer() error {
	packDir, _ := os.Getwd()
	runDir := filepath.Join(packDir, ".run")

	if _, err := os.Stat(runDir); os.IsNotExist(err) {
		fmt.Println("No server directory to clean")
		return nil
	}

	fmt.Println("Cleaning server directory...")
	if err := os.RemoveAll(runDir); err != nil {
		return fmt.Errorf("failed to clean server directory: %w", err)
	}

	fmt.Println("Server directory cleaned")
	return nil
}

func downloadPackwizInstaller(path string) error {
	url := "https://github.com/packwiz/packwiz-installer/releases/latest/download/packwiz-installer-bootstrap.jar"

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download: HTTP %d", resp.StatusCode)
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	return err
}

func downloadServerJar(path string) error {
	// Download latest Fabric server JAR
	// This URL should be updated based on the latest Fabric version
	url := "https://meta.fabricmc.net/v2/versions/loader/1.21.4/0.16.9/1.0.1/server/jar"

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download server JAR: HTTP %d", resp.StatusCode)
	}

	file, err := os.Create(path)
	if err != nil {
		return err
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
