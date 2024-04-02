package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	toml "github.com/pelletier/go-toml"
)

var (
	modlistRaw         = false
	modlistShowVersion = false
)

func modlist() {
	var modlist []PackToml

	outputFile := filepath.Join(*flagPackDir, "modlist.md")

	// if args contain raw or versions, set modlistRaw or modlistShowVersion to true
	for _, arg := range args {
		switch arg {
		case "raw":
			modlistRaw = true
		case "versions":
			modlistShowVersion = true
		case "help":
			fmt.Println("Usage: pw modlist <args>")
			fmt.Println("Args:")
			fmt.Println("  raw      - output raw modlist")
			fmt.Println("  versions - show mod versions")
		}
	}
	os.Remove(outputFile)
	f, err := os.OpenFile(outputFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}

	// write header to output file
	if !modlistRaw {
		_, err = f.WriteString("# Modlist\n\n")
		if err != nil {
			log.Fatal(err)
		}
	}

	// find all files in pack directory using filepath.Walk
	filepath.Walk(*flagPackDir, func(path string, info os.FileInfo, _ error) error {
		// check if file is a .pw.toml file
		if strings.HasSuffix(path, ".pw.toml") {
			// read file
			file, err := os.Open(path)
			if err != nil {
				log.Fatal(err)
			}
			defer file.Close()
			// decode file
			var mod PackToml
			err = toml.NewDecoder(file).Decode(&mod)
			if err != nil {
				log.Fatal(err)
			}
			// set mod.Parse.ModID to the last part of the path without the .pw.toml extension
			mod.Parse.ModID = strings.TrimSuffix(filepath.Base(path), ".pw.toml")
			// append mod to modlist
			modlist = append(modlist, mod)
		}
		return nil
	})

	// sort modlist by side
	var clientMods []PackToml
	var serverMods []PackToml
	var sharedMods []PackToml
	for _, mod := range modlist {
		switch mod.Side {
		case "client":
			clientMods = append(clientMods, mod)
		case "server":
			serverMods = append(serverMods, mod)
		case "both":
			sharedMods = append(sharedMods, mod)
		}
	}

	// write client mods to output file
	var clientHeader = "## Client Mods\n\n"
	var sharedHeader = "## Shared Mods\n\n"
	var serverHeader = "## Server Mods\n\n"
	writeSection(clientHeader, clientMods, f)
	writeSection(sharedHeader, sharedMods, f)
	writeSection(serverHeader, serverMods, f)

}

func writeSection(header string, mods []PackToml, f *os.File) {
	if len(mods) > 0 {
		_, err := f.WriteString(header)
		if err != nil {
			log.Fatal(err)
		}
		for _, mod := range mods {
			writeMod(mod, f)
		}
	}
	// write newline
	_, err := f.WriteString("\n")
	if err != nil {
		log.Fatal(err)
	}
}

func writeMod(mod PackToml, f *os.File) {
	var modURL string
	if mod.Update.Modrinth.ModID != "" {
		modURL = "https://modrinth.com/mod/" + mod.Update.Modrinth.ModID
		if modlistShowVersion {
			modURL += "/version/" + mod.Update.Modrinth.Version
		}
	} else if mod.Update.Curseforge.ProjectID != 0 {
		modURL = "https://www.curseforge.com/minecraft/mc-mods/" + mod.Parse.ModID
		if modlistShowVersion {
			modURL += "/files/" + strconv.Itoa(mod.Update.Curseforge.FileID)
		}
	} else {
		modURL = mod.Download.URL
	}
	var err error
	if modlistRaw {
		_, err = f.WriteString(mod.Name + "\n" + modURL + "\n\n")
	} else {
		_, err = f.WriteString("- [" + mod.Name + "](" + modURL + ")\n")
	}
	if err != nil {
		log.Fatal(err)
	}
}
