package main

import (
	"fmt"
	"log"
	"os"
	"strings"
)

func importFromFile(importFile string) {
	log.Println("IMPORTING", importFile)
	// check if file exists, if not, check if ./import.txt exists and notify user, if not, exit with error
	if _, err := os.Stat(importFile); os.IsNotExist(err) {
		if _, err := os.Stat("./import.txt"); os.IsNotExist(err) {
			fmt.Println("[ERROR]\n", "File not found")
			os.Exit(1)
		} else {
			importFile = "./import.txt"
		}
	}
	// open file and read contents
	file, err := os.ReadFile(importFile)
	if err != nil {
		fmt.Println("[ERROR]\n", err)
		os.Exit(1)
	}
	// print file contents
	fileContent := string(file)

	data := strings.Split(fileContent, "\n")
	previousLine := ""
	for _, line := range data {
		line = strings.ReplaceAll(line, " ", "")
		line = strings.ReplaceAll(line, "	", "")
		if !strings.HasPrefix(line, "https://") {
			if strings.HasPrefix(line, "/") {
				previousLine = line
			}
			continue
		} else {
			installArgs := []string{}
			// use a switch statement to determine which command to use
			switch {
			case strings.HasPrefix(line, "https://www.curseforge.com/"):
				installArgs = append(installArgs, "cf", "add", line)
			case strings.HasPrefix(line, "https://modrinth.com/"):
				installArgs = append(installArgs, "mr", "add", line)
			default:
				installArgs = append(installArgs, "url", "add")
				filepath := strings.Split(line, "/")       // split the url by /
				filename := filepath[len(filepath)-1]      // get the last element of the url
				filename = strings.Split(filename, ".")[0] // remove the file extensions
				if previousLine != "" {
					installArgs = append(installArgs, "--meta-folder", previousLine)
				}
				installArgs = append(installArgs, filename)
				installArgs = append(installArgs, line)
			}
			packwiz(*flagPackDir, installArgs)
			previousLine = ""
		}
	}
}
