package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Merith-TK/packwiz-wrapper/cmd/pw/internal/packwiz"
	"github.com/pelletier/go-toml"
)

// CmdReinstall provides mod reinstallation functionality
func CmdReinstall() (names []string, shortHelp, longHelp string, execute func([]string) error) {
	return []string{"reinstall", "refresh-mods"},
		"Refresh and reinstall all mods in the pack",
		`Reinstall Commands:
  pw reinstall            - Refresh pack and reinstall all mods (latest versions)
  pw reinstall versions   - Reinstall all mods preserving exact versions

Examples:
  pw reinstall            - Standard refresh and reinstall with latest versions
  pw reinstall versions   - Reinstall preserving current mod versions`,
		func(args []string) error {
			showVersions := false
			
			// Parse arguments
			for _, arg := range args {
				if arg == "versions" {
					showVersions = true
				}
			}
			
			return reinstallMods(showVersions)
		}
}

func reinstallMods(showVersions bool) error {
	packDir, _ := os.Getwd()
	client := packwiz.NewClient(packDir)
	
	fmt.Println("Refreshing pack...")
	if err := client.Execute([]string{"refresh"}); err != nil {
		return fmt.Errorf("failed to refresh pack: %w", err)
	}
	
	// Find pack directory
	packLocation := client.GetPackDir()
	if packLocation == "" {
		return fmt.Errorf("pack.toml not found")
	}
	
	// Read index.toml
	indexFile := filepath.Join(packLocation, "index.toml")
	indexFileHandler, err := os.Open(indexFile)
	if err != nil {
		return fmt.Errorf("failed to open index.toml: %w", err)
	}
	defer indexFileHandler.Close()
	
	var index packwiz.IndexToml
	if err := toml.NewDecoder(indexFileHandler).Decode(&index); err != nil {
		return fmt.Errorf("failed to decode index.toml: %w", err)
	}
	
	// Build mod list for reinstallation
	var modlist []packwiz.ModToml
	var errors []string
	
	fmt.Println("Reading mod metadata...")
	for _, file := range index.Files {
		if !file.Metafile {
			continue
		}
		
		modFilePath := filepath.Join(packLocation, file.File)
		modFile, err := os.Open(modFilePath)
		if err != nil {
			errors = append(errors, fmt.Sprintf("Failed to open %s: %v", file.File, err))
			continue
		}
		
		var mod packwiz.ModToml
		if err := toml.NewDecoder(modFile).Decode(&mod); err != nil {
			errors = append(errors, fmt.Sprintf("Failed to decode %s: %v", file.File, err))
			modFile.Close()
			continue
		}
		modFile.Close()
		
		// Set parse information for mod identification
		mod.Parse.ModID = strings.TrimSuffix(filepath.Base(modFilePath), ".pw.toml")
		if mod.Update.Modrinth.ModID == "" && mod.Update.Curseforge.ProjectID == 0 {
			// For URL files (no modrinth or curseforge)
			mod.Parse.Path = filepath.Dir(modFilePath)
		}
		
		modlist = append(modlist, mod)
	}
	
	if len(errors) > 0 {
		fmt.Printf("Encountered %d error(s) reading mod files:\n", len(errors))
		for _, err := range errors {
			fmt.Printf("  - %s\n", err)
		}
	}
	
	if len(modlist) == 0 {
		fmt.Println("No mods found to reinstall")
		return nil
	}
	
	fmt.Printf("Found %d mod(s) to reinstall\n", len(modlist))
	
	// Remove all mods first
	fmt.Println("Removing existing mods...")
	for _, mod := range modlist {
		fmt.Printf("Removing: %s\n", mod.Name)
		if err := client.Execute([]string{"remove", mod.Parse.ModID}); err != nil {
			fmt.Printf("Warning: failed to remove %s: %v\n", mod.Name, err)
		}
	}
	
	// Re-add all mods
	fmt.Println("Re-adding mods...")
	for _, mod := range modlist {
		if showVersions {
			version := getModVersionForReinstall(mod)
			if version != "" {
				fmt.Printf("Reinstalling: %s (v%s)\n", mod.Name, version)
			} else {
				fmt.Printf("Reinstalling: %s\n", mod.Name)
			}
		} else {
			fmt.Printf("Reinstalling: %s\n", mod.Name)
		}
		
		if err := reinstallSingleMod(client, mod, showVersions); err != nil {
			fmt.Printf("Warning: failed to reinstall %s: %v\n", mod.Name, err)
		} else {
			fmt.Printf("Successfully reinstalled: %s\n", mod.Name)
		}
	}
	
	fmt.Printf("\nReinstallation completed for %d mod(s)\n", len(modlist))
	return nil
}

func reinstallSingleMod(client *packwiz.Client, mod packwiz.ModToml, withVersions bool) error {
	var arguments []string
	
	if mod.Update.Modrinth.ModID != "" {
		// Modrinth mod
		arguments = append(arguments, "mr", "add", "--project-id", mod.Update.Modrinth.ModID)
		if withVersions && mod.Update.Modrinth.Version != "" {
			arguments = append(arguments, "--version-id", mod.Update.Modrinth.Version)
		}
	} else if mod.Update.Curseforge.ProjectID != 0 {
		// CurseForge mod
		arguments = append(arguments, "cf", "add", "--addon-id", fmt.Sprint(mod.Update.Curseforge.ProjectID))
		if withVersions && mod.Update.Curseforge.FileID != 0 {
			arguments = append(arguments, "--file-id", fmt.Sprint(mod.Update.Curseforge.FileID))
		}
	} else {
		// URL mod
		arguments = append(arguments, "url", "add", mod.Parse.ModID, mod.Download.URL)
		if mod.Parse.Path != "" {
			arguments = append(arguments, "--meta-folder", mod.Parse.Path)
		}
	}
	
	return client.Execute(arguments)
}

func getModVersionForReinstall(mod packwiz.ModToml) string {
	// Try to get version from update sources
	if mod.Update.Modrinth.Version != "" {
		return mod.Update.Modrinth.Version
	}
	
	// For CurseForge, we could try to extract from filename or other metadata
	if mod.Update.Curseforge.FileID != 0 {
		return fmt.Sprintf("CF:%d", mod.Update.Curseforge.FileID)
	}
	
	return ""
}
