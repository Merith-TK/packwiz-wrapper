// Package modlist provides modlist generation functionality
package modlist

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Merith-TK/packwiz-wrapper/internal/packwiz"
	"github.com/Merith-TK/packwiz-wrapper/internal/utils"
	"github.com/pelletier/go-toml"
)

// GetCategorizedMods processes mod files and returns them categorized by side
func GetCategorizedMods() ([]packwiz.ModToml, []packwiz.ModToml, []packwiz.ModToml, error) {
	packDir, _ := os.Getwd()

	// Find pack directory
	packLocation := utils.FindPackToml(packDir)
	if packLocation == "" {
		return nil, nil, nil, fmt.Errorf("pack.toml not found")
	}

	// Read index.toml
	indexFile := filepath.Join(packLocation, "index.toml")
	indexFileHandler, err := os.Open(indexFile)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to open index.toml: %w", err)
	}
	defer indexFileHandler.Close()

	var index packwiz.IndexToml
	if err := toml.NewDecoder(indexFileHandler).Decode(&index); err != nil {
		return nil, nil, nil, fmt.Errorf("failed to decode index.toml: %w", err)
	}

	// Process mod files
	var modlist []packwiz.ModToml
	for _, file := range index.Files {
		if !file.Metafile {
			continue
		}

		modFilePath := filepath.Join(packLocation, file.File)
		modFile, err := os.Open(modFilePath)
		if err != nil {
			continue
		}

		var mod packwiz.ModToml
		if err := toml.NewDecoder(modFile).Decode(&mod); err != nil {
			modFile.Close()
			continue
		}
		modFile.Close()

		// Set mod.Parse.ModID to the last part of the path without the .pw.toml extension
		modID := strings.TrimSuffix(filepath.Base(modFilePath), ".pw.toml")
		mod.Parse.ModID = modID

		modlist = append(modlist, mod)
	}

	// Sort mods by side
	var clientMods []packwiz.ModToml
	var serverMods []packwiz.ModToml
	var sharedMods []packwiz.ModToml

	for _, mod := range modlist {
		switch mod.Side {
		case "client":
			clientMods = append(clientMods, mod)
		case "server":
			serverMods = append(serverMods, mod)
		default: // "both" or empty - treat as shared
			sharedMods = append(sharedMods, mod)
		}
	}

	return clientMods, sharedMods, serverMods, nil
}
