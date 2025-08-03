package packwiz

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Client handles interactions with the packwiz command-line tool
type Client struct {
	PackDir string
}

// NewClient creates a new packwiz client
func NewClient(packDir string) *Client {
	return &Client{
		PackDir: packDir,
	}
}

// Execute runs packwiz with the given arguments
func (c *Client) Execute(args []string) error {
	dir := c.findPackToml()
	if dir == "" {
		return fmt.Errorf("pack.toml not found")
	}
	
	fmt.Println("[PackWrap] Handoff: ["+dir+"] packwiz", strings.Join(args, " "))
	cmd := exec.Command("packwiz", args...)
	cmd.Dir = filepath.Dir(dir)
	
	if _, err := os.Stat(cmd.Dir); err != nil {
		fmt.Println("[PackWrap] [ERROR] packwiz directory not found, creating...")
		if err := os.Mkdir(cmd.Dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	}
	
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("packwiz command failed: %w", err)
	}
	
	fmt.Print("\n")
	return nil
}

// ExecuteInDir runs packwiz with the given arguments in the specified directory
func (c *Client) ExecuteInDir(args []string, dir string) error {
	fmt.Println("[PackWrap] Handoff: ["+dir+"] packwiz", strings.Join(args, " "))
	cmd := exec.Command("packwiz", args...)
	cmd.Dir = dir
	
	if _, err := os.Stat(cmd.Dir); err != nil {
		fmt.Println("[PackWrap] [ERROR] packwiz directory not found, creating...")
		if err := os.Mkdir(cmd.Dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	}
	
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("packwiz command failed: %w", err)
	}
	
	fmt.Print("\n")
	return nil
}

// ExecuteArbitrary runs an arbitrary command in the pack directory
func (c *Client) ExecuteArbitrary(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("no command provided")
	}
	
	dir := c.PackDir
	fmt.Println("[PackWrap] Arbitrary: ["+dir+"]", strings.Join(args, " "))
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = filepath.Dir(dir)
	
	if _, err := os.Stat(cmd.Dir); err != nil {
		fmt.Println("[PackWrap] [ERROR] arbitrary directory not found, creating...")
		if err := os.Mkdir(cmd.Dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	}
	
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("arbitrary command failed: %w", err)
	}
	
	fmt.Print("\n")
	return nil
}

// GetPackDir returns the pack directory containing pack.toml
func (c *Client) GetPackDir() string {
	return c.findPackToml()
}

// findPackToml locates the pack.toml file and returns the directory containing it
func (c *Client) findPackToml() string {
	dir := c.PackDir
	
	// Check for pack.toml in current directory
	_, err := os.Stat(filepath.Join(dir, "pack.toml"))
	if err == nil {
		dir = filepath.ToSlash(dir)
		if !strings.HasSuffix(dir, "/") {
			dir = dir + "/"
		}
		log.Println("[PackWrap] Found Pack Directory:", dir)
		return dir
	}
	
	// Check for pack.toml in .minecraft subdirectory
	_, err = os.Stat(filepath.Join(dir, ".minecraft", "pack.toml"))
	if err == nil {
		fmt.Println("[PackWrap] Using pack.toml from .minecraft")
		dir = filepath.Join(dir, ".minecraft")
		dir = filepath.ToSlash(dir)
		if !strings.HasSuffix(dir, "/") {
			dir = dir + "/"
		}
		return dir
	}
	
	fmt.Println("[PackWrap] [ERROR] pack.toml not found")
	return ""
}

// IsPackwizInstalled checks if the packwiz command is available
func IsPackwizInstalled() bool {
	_, err := exec.LookPath("packwiz")
	return err == nil
}
