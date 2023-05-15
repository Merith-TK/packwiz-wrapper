package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type IndexToml struct {
	HashFormat string `toml:"hash-format"`
	Files      []struct {
		File     string `toml:"file"`
		Hash     string `toml:"hash"`
		Metafile bool   `toml:"metafile,omitempty"`
	} `toml:"files"`
}
type PackToml struct {
	Name     string `toml:"name"`
	Filename string `toml:"filename"`
	Side     string `toml:"side"`
	Download struct {
		URL        string `toml:"url"`
		HashFormat string `toml:"hash-format"`
		Hash       string `toml:"hash"`
	} `toml:"download"`
	Update struct {
		Modrinth struct {
			ModID   string `toml:"mod-id"`
			Version string `toml:"version"`
		} `toml:"modrinth"`
		Curseforge struct {
			FileID    int `toml:"file-id"`
			ProjectID int `toml:"project-id"`
		} `toml:"curseforge"`
	} `toml:"update"`
	// Parse is specific to this program
	Parse struct {
		ModID string `toml:"mod-id"`
		Path  string `toml:"path"`
	} `toml:"parse"`
}

// run packwiz with args
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
