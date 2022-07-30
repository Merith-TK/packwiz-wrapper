package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// list of strings
var list = []string{
	"completion",
	"cf", "curseforge",
	"mr", "modrinth",
	"init",
	"list",
	"refresh",
	"update",
	"serve",
	"utils",
}

var (
	Version = "0.0.1"

	// flags
	flagHelp    = flag.Bool("h", false, "show help")
	flagVersion = flag.Bool("v", false, "show version")
	flagPackDir = flag.String("d", ".", "pack directory")

	// import file
	flagImport = flag.String("i", "", "import links from file")
)

func main() {
	flag.Parse()
	args := flag.Args()

	if _, err := exec.LookPath("packwiz"); err != nil {
		fmt.Println("[PackWrap] \n[ERROR] packwiz is not installed,\nplease install it with 'go install github.com/packwiz/packwiz@latest'")
		return
	}

	if !strings.HasSuffix(*flagPackDir, "/") {
		*flagPackDir += "/"
		fmt.Println("[PackWrap] PackDir:", *flagPackDir)

	}
	if *flagVersion {
		fmt.Println("[PackWrap] version:", Version)
		return
	}
	if *flagHelp {
		fmt.Println("[PackWrap]")
		flag.Usage()
		fmt.Println("")
		packwiz([]string{"help"})
		return
	}

	// check for pack.toml in flagPackDir
	if _, err := os.Stat(*flagPackDir + "pack.toml"); err != nil {
		fmt.Println("[PackWrap] \n[ERROR] pack.toml not found in", *flagPackDir)
		return
	}

	if *flagImport != "" {
		importFromFile()
		return
	}

	if len(args) > 0 {
		packwiz(args)
		return
	}
	packwiz([]string{"refresh"})

}

func packwiz(args []string) {
	fmt.Println("[PackWrap] Handoff: packwiz", strings.Join(args, " "))
	cmd := exec.Command("packwiz", args...)
	cmd.Dir = filepath.Dir(*flagPackDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Run()
	fmt.Print("\n")
}

func importFromFile() {
	var file, err = ioutil.ReadFile(*flagImport)
	if err != nil {
		fmt.Println("[ERROR]\n", err)
		os.Exit(1)
	}
	// print file contents
	fileContent := string(file)

	data := strings.Split(fileContent, "\n")
	for _, line := range data {
		if !strings.HasPrefix(line, "https://") {
			continue
		} else {
			var modHost = ""

			if strings.Contains(line, "modrinth.com/mod/") {
				modHost = "mr"
			}
			if strings.Contains(line, "curseforge.com/minecraft/mc-mods/") {
				modHost = "cf"
			}
			if modHost == "" {
				fmt.Println("[ERROR] unknown host", line)
				continue
			} else {
				packwiz([]string{modHost, "install", line})
			}
		}
	}
}
