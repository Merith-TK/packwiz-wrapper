package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml"
)

func reinstall() {
	packwiz(*flagPackDir, []string{"refresh"})
	var index IndexToml
	var modlist []ModToml

	withVersions := false
	// read all arguemnts
	for _, arg := range args[1:] {
		switch arg {
		case "versions":
			withVersions = true
		}
	}

	// read index.toml
	indexFile := filepath.Join(*flagPackDir, "index.toml")
	// open index.toml
	indexFileHandler, err := os.Open(indexFile)
	if err != nil {
		log.Fatal(err)
	}
	err = toml.NewDecoder(indexFileHandler).Decode(&index)
	if err != nil {
		log.Fatal(err)
	}
	indexFileHandler.Close()

	for _, file := range index.Files {
		if !file.Metafile {
			continue
		}
		// read file.File to ModToml
		var packtoml ModToml
		packtomlFile := filepath.Join(*flagPackDir, file.File)
		packtomlFileHandler, err := os.Open(packtomlFile)
		if err != nil {
			log.Fatal(err)
		}
		err = toml.NewDecoder(packtomlFileHandler).Decode(&packtoml)
		if err != nil {
			log.Fatal(err)
		}
		// close the file
		packtomlFileHandler.Close()

		packtoml.Parse.ModID = strings.TrimSuffix(filepath.Base(packtomlFile), ".pw.toml")
		if packtoml.Update.Modrinth.ModID == "" && packtoml.Update.Curseforge.ProjectID == 0 {
			// for URL files (no modrinth or curseforge)
			packtoml.Parse.Path = filepath.Dir(packtomlFile)
		}
		modlist = append(modlist, packtoml)
	}

	for _, mod := range modlist {
		packwiz(*flagPackDir, []string{"remove", mod.Parse.ModID})
	}
	for _, mod := range modlist {
		arguments := []string{}
		if mod.Update.Modrinth.ModID != "" {
			arguments = append(arguments, "mr", "add", "--project-id", mod.Update.Modrinth.ModID)
			if withVersions {
				arguments = append(arguments, "--version-id", mod.Update.Modrinth.Version)
			}
		} else if mod.Update.Curseforge.ProjectID != 0 {
			arguments = append(arguments, "cf", "add", "--addon-id", fmt.Sprint(mod.Update.Curseforge.ProjectID))
			if withVersions {
				arguments = append(arguments, "--file-id", fmt.Sprint(mod.Update.Curseforge.FileID))
			}
		} else {
			arguments = append(arguments, "url", "add", mod.Parse.ModID, mod.Download.URL, "--meta-folder", mod.Parse.Path)
		}

		packwiz(*flagPackDir, arguments)

	}
}
