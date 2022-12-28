package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

var (
	Version = "0.0.1"

	// flags
	flagHelp      = flag.Bool("h", false, "show help")
	flagVersion   = flag.Bool("v", false, "show version")
	flagPackDir   = flag.String("d", ".", "pack directory")
	flatOuputFile = flag.String("o", "modlist.md", "output file")
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

	// list files in flagPackDir/mods
	mods, _ := ioutil.ReadDir(*flagPackDir + "/mods")
	var modList []string
	for _, mod := range mods {
		// read the file contents
		contents, _ := ioutil.ReadFile(*flagPackDir + "/mods/" + mod.Name())
		// trim .pw.toml from the filename
		modName := strings.TrimSuffix(mod.Name(), ".pw.toml")

		lines := strings.Split(string(contents), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "[update.modrinth]") {
				modList = append(modList, "["+modName+"](https://modrinth.com/mod/"+modName+")")
			}
			if strings.HasPrefix(line, "[update.curseforge]") {
				modList = append(modList, "["+modName+"](https://www.curseforge.com/minecraft/mc-mods/"+modName+")")
			}

		}
	}
	generateModList(modList, *flatOuputFile)
}

func generateModList(modList []string, outputFile string) {
	// if the file exists, delete it
	if _, err := os.Stat(outputFile); err == nil {
		err = os.Remove(outputFile)
		if err != nil {
			log.Fatal(err)
		}
	}
	// create the file
	file, err := os.Create(outputFile)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// write the header
	_, err = file.WriteString("# Mod List\n")
	_, err = file.WriteString("\n")
	if err != nil {
		log.Fatal(err)
	}

	// write the mod list
	for _, mod := range modList {
		_, err = file.WriteString(fmt.Sprintf("- %s\n", mod))
		if err != nil {
			log.Fatal(err)
		}

	}
}
