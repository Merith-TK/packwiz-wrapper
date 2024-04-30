package main

import (
	"fmt"
	"log"
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
type ModToml struct {
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
	} `toml:"update,omitempty"`
	// Parse is specific to this program
	Parse struct {
		ModID string `toml:"mod-id"`
		Path  string `toml:"path"`
	} `toml:"parse"`
}

// run packwiz with args
func packwiz(dir string, args []string) {
	dir = findPackToml(dir)
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

func executeArb(dir string, args []string) {
	fmt.Println("[PackWrap] Arbitrary: ["+dir+"]", strings.Join(args, " "))
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = filepath.Dir(dir)
	if _, err := os.Stat(cmd.Dir); err != nil {
		fmt.Println("[PackWrap] [ERROR] arbitrary directory not found, creating...")
		os.Mkdir(cmd.Dir, 0755)
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Run()
	fmt.Print("\n")
}

func findPackToml(dir string) string {
	// all this, just to find the pack.toml
	_, err := os.Stat(filepath.Join(dir, "pack.toml"))
	if err != nil {
		_, err = os.Stat(filepath.Join(dir, ".minecraft", "pack.toml"))
		if err != nil {
			fmt.Println("[PackWrap] [ERROR] pack.toml not found")
			return ""
		}
		fmt.Println("[PackWrap] Using pack.toml from .minecraft")
		dir = filepath.Join(dir, ".minecraft")
		dir = filepath.ToSlash(dir)
		if !strings.HasSuffix(dir, "/") {
			dir = dir + "/"
		}
		return dir
	}
	dir = filepath.ToSlash(dir)
	if !strings.HasSuffix(dir, "/") {
		dir = dir + "/"
	}
	log.Println("[PackWrap] Found Pack Directory:", dir)
	return dir
}
