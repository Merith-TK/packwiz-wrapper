package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var (
	packDir = ""
)

func main() {
	// check if either .minecraft/pack.toml or ./pack.toml exists
	// if neither exist, exit with error

	packFound := false
	packPath := ""

	// if .minecraft/pack.toml exists, use that
	if _, err := os.Stat(".minecraft/pack.toml"); err == nil {
		packFound = true
		packPath = ".minecraft/"
	} else if _, err := os.Stat("./pack.toml"); err == nil {
		packFound = true
		packPath = "./"
	}
	if !packFound {
		println("pack.toml not found")
		return
	} else {
		packDir = packPath
	}

	// list files in packDir/mods
	mods, _ := ioutil.ReadDir(packDir + "/mods")
	var modList []string
	for _, mod := range mods {
		// read the file contents
		contents, _ := ioutil.ReadFile(packDir + "/mods/" + mod.Name())
		// trim .pw.toml from the filename
		modName := strings.TrimSuffix(mod.Name(), ".pw.toml")

		lines := strings.Split(string(contents), "\n")

		for _, line := range lines {
			if strings.HasPrefix(line, "[update.modrinth]") {
				modList = append(modList, "https://modrinth.com/mod/"+modName)
			}
			if strings.HasPrefix(line, "[update.curseforge]") {
				modList = append(modList, "https://www.curseforge.com/minecraft/mc-mods/"+modName)
			}
		}
	}
	reinstallMods(modList)
}

func reinstallMods(modList []string) {
	// for each mod in modList, remote the coresonding file in packDir/mods
	for _, mod := range modList {
		// remove "https://www.curseforge.com/minecraft/mc-mods/" or "https://modrinth.com/mod/" from the mod name
		modName := strings.TrimPrefix(mod, "https://www.curseforge.com/minecraft/mc-mods/")
		modName = strings.TrimPrefix(mod, "https://modrinth.com/mod/")
		modName = strings.TrimSuffix(modName, ".pw.toml")
		// remove the file in packDir/mods
		err := os.Remove(packDir + "/mods/" + modName + ".pw.toml")
		if err != nil {
			println(err)
		}
	}
	command := ""
	for _, mod := range modList {
		if strings.HasPrefix(mod, "https://www.curseforge.com/minecraft/mc-mods/") {
			// set command to cf install mod
			command = "cf install " + mod
		} else if strings.HasPrefix(mod, "https://modrinth.com/mod/") {
			// set command to mr install mod
			command = "mr install " + mod
		}
		packwiz(packDir, strings.Split(command, " "))
	}
}

func packwiz(dir string, args []string) {
	fmt.Println("[PackWrap] Handoff: ["+dir+"] packwiz", strings.Join(args, " "))
	cmd := exec.Command("packwiz", args...)
	cmd.Dir = filepath.Dir(dir)
	if _, err := os.Stat(cmd.Dir); err != nil {
		fmt.Println("[PackWrap] [ERROR] packwiz directory not found, creating...")
		os.Mkdir(cmd.Dir, 0755)
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Run()
	fmt.Print("\n")
}
