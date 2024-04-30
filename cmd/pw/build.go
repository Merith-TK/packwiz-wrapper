package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func buildPack(packdir string, arguments []string) {
	// ensure the .build directory exists
	if _, err := os.Stat(".build"); os.IsNotExist(err) {
		os.Mkdir(".build", 0755)
	}
	// check if packwiz is installed
	if _, err := exec.LookPath("packwiz"); err != nil {
		fmt.Println("[PackWrap] \n[ERROR] packwiz is not installed,\nplease install it with 'go install github.com/packwiz/packwiz@latest'")
		return
	}
	if len(arguments) == 0 {
		fmt.Println("[PackWrap] \n[ERROR] valid arguments are: modrinth (mr),  curseforge (cf), packwiz (pw)")
		return
	}
	switch arguments[0] {
	case "modrinth", "mr":
		exportmr(packdir)
	case "curseforge", "cf":
		exportcf(packdir)
	case "packwiz", "pw":
		exportpw()
	default:
		fmt.Println("[PackWrap] \n[ERROR] invalid export argument")
	}
}

func moveBuildFiles(extension string, packdir string) {
	actualPackDir := findPackToml(packdir)
	filepath.Walk(actualPackDir, func(path string, info os.FileInfo, err error) error {
		// if directory, skip
		if info.IsDir() {
			return nil
		}
		if filepath.Ext(path) == "."+extension {
			// Move the file to the .build directory with the appended timestamp -Month-Day_Hour-Minute-Second
			timestamp := info.ModTime().Format("_01-02_15-04-05")
			filename := fmt.Sprintf("%s%s.%s", strings.TrimSuffix(filepath.Base(path), filepath.Ext(path)), timestamp, extension)
			newFilePath := filepath.Join(".build", filepath.Base(filename))
			if _, err := os.Stat(newFilePath); err == nil {
				fmt.Println("[PackWrap] [ERROR] file already exists in .build directory")
				return err
			}
			if !strings.HasSuffix(filename, "."+extension) {
				filename += "." + extension
			}

			err := os.Rename(path, filepath.Join(".build", filepath.Base(filename)))
			if err != nil {
				fmt.Println("[PackWrap] [ERROR] failed to move file to .build directory")
				fmt.Println(err)
				return err
			}
		}
		return nil
	})
}

func exportmr(packdir string) {
	packwiz(packdir, []string{"modrinth", "export"})
	moveBuildFiles("mrpack", packdir)

}
func exportcf(packdir string) {
	packwiz(packdir, []string{"curseforge", "export"})
	moveBuildFiles("zip", packdir)
}
func exportpw() {
	// export packwiz modpack
}
