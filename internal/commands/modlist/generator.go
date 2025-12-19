// Package modlist provides modlist generation functionality
package modlist

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Merith-TK/packwiz-wrapper/internal/modformat"
)

// GenerateModlist generates a modlist based on the provided options
func GenerateModlist(opts modformat.ModlistOptions) error {
	clientMods, sharedMods, serverMods, err := GetCategorizedMods()
	if err != nil {
		return err
	}

	// Pre-fetch all author information if needed (batch API calls)
	var authorCache map[int]modformat.ModAuthorInfo
	if opts.ShowAuthors {
		fmt.Println("Fetching mod author information from APIs...")
		allMods := append(append(clientMods, sharedMods...), serverMods...)
		authorCache = modformat.FetchModAuthors(allMods)
		fmt.Printf("Fetched author info for %d mods\n", len(authorCache))
	}

	// Prepare output file (only if not raw output or print only)
	var outputFile *os.File
	packDir, _ := os.Getwd()
	outputPath := filepath.Join(packDir, "modlist.md")

	if (!opts.RawOutput) && (!opts.OnlyPrint) {
		os.Remove(outputPath) // Remove existing file
		outputFile, err = os.OpenFile(outputPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("failed to create modlist.md: %w", err)
		}
		defer outputFile.Close()

		// Write header
		if _, err := outputFile.WriteString("# Modlist\n\n"); err != nil {
			return fmt.Errorf("failed to write header: %w", err)
		}
	}

	// Write sections in order: Client, Shared, Server
	if opts.RawOutput {
		// For raw output, just print all mods with their URLs
		allMods := append(append(clientMods, sharedMods...), serverMods...)
		for i, mod := range allMods {
			modURL := modformat.GetModURL(mod, opts.ShowVersions)

			// Determine platform from TOML structure
			platform := modformat.GetPlatform(mod)

			if opts.ShowAuthors || opts.ShowPlatform {
				// Build output based on flags
				output := mod.Name

				if opts.ShowAuthors {
					// Use cached author info if available
					if authorInfo, ok := authorCache[i]; ok {
						output += " - by " + authorInfo.Author
					}
				}

				if opts.ShowPlatform && platform != "" {
					output += " [" + platform + "]"
				}

				fmt.Printf("%s\n%s\n\n", output, modURL)
			} else {
				fmt.Printf("%s\n%s\n\n", mod.Name, modURL)
			}
		}
	} else {
		totalMods := len(clientMods) + len(serverMods) + len(sharedMods)
		fmt.Printf("Found %d mods (%d client, %d shared, %d server)\n",
			totalMods, len(clientMods), len(sharedMods), len(serverMods))

		WriteSection("## Client Mods\n\n", clientMods, outputFile, opts, authorCache, 0)
		WriteSection("## Shared Mods\n\n", sharedMods, outputFile, opts, authorCache, len(clientMods))
		WriteSection("## Server Mods\n\n", serverMods, outputFile, opts, authorCache, len(clientMods)+len(sharedMods))

		if outputFile != nil {
			fmt.Printf("Modlist written to modlist.md\n")
		}
	}

	return nil
}
