package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	toml "github.com/pelletier/go-toml"
)

var (
	Version = "0.0.1"

	// flags
	flagHelp      = flag.Bool("h", false, "show help")
	flagVersion   = flag.Bool("v", false, "show version")
	flagPackDir   = flag.String("d", ".", "pack directory")
	flatOuputFile = flag.String("o", "modlist.md", "output file")
	flagRaw       = flag.Bool("r", false, "raw output")
)

func main() {
	flag.Parse()
	//args := flag.Args()

	if *flagVersion {
		println(Version)
		return
	}
	if *flagHelp {
		flag.Usage()
		return
	}
	if *flatOuputFile == "" {
		println("output file is required")
		return
	}
	if *flagPackDir == "" {
		println("pack directory is required")
		return
	}

	// delte output file if it exists
	os.Remove(*flatOuputFile)
	// open output file
	f, err := os.OpenFile(*flatOuputFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}

	// write header to output file
	if !*flagRaw {
		_, err = f.WriteString("# Modlist\n\n")
		if err != nil {
			log.Fatal(err)
		}
	}

	// find all files in pack directory using filepath.Walk
	err = filepath.Walk(*flagPackDir, func(path string, info os.FileInfo, err error) error {
		// check if file is a .pw.toml file
		if strings.HasSuffix(path, ".pw.toml") {
			// read file
			file, err := os.Open(path)
			if err != nil {
				log.Fatal(err)
			}
			defer file.Close()
			// decode file
			var mod PackwizToml
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
	var clientMods []PackwizToml
	var serverMods []PackwizToml
	var sharedMods []PackwizToml
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

func writeSection(header string, mods []PackwizToml, f *os.File) {
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

func writeMod(mod PackwizToml, f *os.File) {
	var modURL string
	if mod.Update.Modrinth.ModID != "" {
		modURL = "https://modrinth.com/mod/" + mod.Update.Modrinth.ModID + "/version/" + mod.Update.Modrinth.Version
	} else if mod.Update.Curseforge.ProjectID != 0 {
		modURL = "https://www.curseforge.com/minecraft/mc-mods/" + mod.Parse.ModID + "/files/" + strconv.Itoa(mod.Update.Curseforge.FileID)
	} else {
		modURL = mod.Download.URL
	}
	var err error
	if *flagRaw {
		_, err = f.WriteString(mod.Name + "\n" + modURL + "\n\n")
	} else {
		_, err = f.WriteString("- [" + mod.Name + "](" + modURL + ")\n")
	}
	if err != nil {
		log.Fatal(err)
	}
}
