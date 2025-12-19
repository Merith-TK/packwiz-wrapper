package modformat

import (
	"fmt"
	"strconv"

	"github.com/Merith-TK/packwiz-wrapper/internal/packwiz"
	"github.com/Merith-TK/packwiz-wrapper/internal/utils"
)

const (
	PlatformModrinth   = "Modrinth"
	PlatformCurseForge = "CurseForge"
	SideClient         = "client"
	SideServer         = "server"
	SideBoth           = "both"
)

// ModlistOptions configures modlist generation behavior
type ModlistOptions struct {
	ShowVersions bool
	ShowAuthors  bool
	ShowPlatform bool
	RawOutput    bool
	OnlyPrint    bool
}

// GetPlatform determines the platform from the mod's TOML structure
func GetPlatform(mod packwiz.ModToml) string {
	if mod.Update.Modrinth.ModID != "" {
		return PlatformModrinth
	} else if mod.Update.Curseforge.ProjectID != 0 {
		return PlatformCurseForge
	}
	return ""
}

// GetModURL generates the URL for a mod based on its platform and version settings
func GetModURL(mod packwiz.ModToml, showVersions bool) string {
	var modURL string

	if mod.Update.Modrinth.ModID != "" {
		modURL = "https://modrinth.com/mod/" + mod.Update.Modrinth.ModID
		if showVersions && mod.Update.Modrinth.Version != "" {
			modURL += "/version/" + mod.Update.Modrinth.Version
		}
	} else if mod.Update.Curseforge.ProjectID != 0 {
		modURL = "https://www.curseforge.com/minecraft/mc-mods/"
		if mod.Parse.ModID != "" {
			modURL += mod.Parse.ModID
		} else {
			modURL += strconv.Itoa(mod.Update.Curseforge.ProjectID)
		}
		if showVersions && mod.Update.Curseforge.FileID != 0 {
			modURL += "/files/" + strconv.Itoa(mod.Update.Curseforge.FileID)
		}
	} else if mod.Download.URL != "" {
		modURL = mod.Download.URL
	} else {
		modURL = "#"
	}

	return modURL
}

// GetModAuthor fetches author info based on platform
func GetModAuthor(mod packwiz.ModToml, platform string) string {
	switch platform {
	case PlatformModrinth:
		if info, err := utils.GetModrinthInfo(mod.Update.Modrinth.ModID); err == nil {
			return info.Author
		}
	case PlatformCurseForge:
		if info, err := utils.GetCurseForgeInfo(mod.Update.Curseforge.ProjectID); err == nil {
			return info.Author
		}
	}
	return ""
}

// FormatModLine generates a formatted markdown line for a mod
// If author is provided (non-empty), it will be used; otherwise, it will be fetched if showAuthors is true
func FormatModLine(mod packwiz.ModToml, opts ModlistOptions, author string) string {
	modURL := GetModURL(mod, opts.ShowVersions)

	// Early return for simple case
	if author == "" && !opts.ShowPlatform {
		return fmt.Sprintf("- [%s](%s)\n", mod.Name, modURL)
	}

	// Build the line progressively
	line := fmt.Sprintf("- [%s](%s)", mod.Name, modURL)

	// Add author if provided
	if author != "" {
		line += fmt.Sprintf(" - *by %s*", author)
	}

	// Add platform if requested
	if opts.ShowPlatform {
		platform := GetPlatform(mod)
		if platform != "" {
			line += fmt.Sprintf(" [%s]", platform)
		}
	}

	return line + "\n"
}
