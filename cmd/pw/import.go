package main

import (
	"fmt"
	"os"
	"strings"
)

func importcmd(args []string) {
	// if there are no arguments, show help
	if len(args) == 0 {
		fmt.Println("> import -i <file> -y")
		fmt.Println("  import mods from file or string")
		fmt.Println("  -i <file>  import mods from a file")
		fmt.Println("  -y         auto confirm")

		return
	}
	autoConfirm := false
	importFile := false
	// check if the auto confirm flag is set
	for i, arg := range args {
		if arg == "-y" {
			autoConfirm = true
			args = append(args[:i], args[i+1:]...)
		}
		if arg == "-i" {
			importFile = true
			args = append(args[:i], args[i+1:]...)
		}
	}

	// check if the import flag is se
	if importFile {
		filename := "./import.txt"
		if len(args) > 0 {
			filename = args[0]
		}
		importFromFile(filename, autoConfirm)
		return
	}
	importFromString(strings.Join(args, ","), autoConfirm)

}

func importFromFile(importFile string, autoConfirm bool) {
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

			// strip URL into base parts
			urlPrefix := ""
			if strings.HasPrefix(line, "https://") {
				urlPrefix = "https://"
				line = strings.ReplaceAll(line, "https://", "")
			} else if strings.HasPrefix(line, "http://") {
				urlPrefix = "http://"
				line = strings.ReplaceAll(line, "http://", "")
			}
			if strings.HasPrefix(line, "www.") {
				line = strings.ReplaceAll(line, "www.", "")
			}
			urlParts := strings.Split(line, "/")
			line = urlPrefix + strings.Join(urlParts, "/")
			switch {
			case urlParts[0] == "curseforge.com" && urlParts[1] == "minecraft" && urlParts[2] == "mc-mods":
				installArgs = append(installArgs, "cf", "add", line)
			case urlParts[0] == "modrinth.com" && !strings.HasSuffix(urlParts[1], "s") && len(urlParts) > 2:
				installArgs = append(installArgs, "mr", "add", line)
			default:
				installArgs = append(installArgs, "url", "add")
				filepath := strings.Split(previousLine, "/") // split the url by /
				filename := filepath[len(filepath)-1]        // get the last element of the url
				filename = strings.Split(filename, ".")[0]   // remove the file extensions
				installArgs = append(installArgs, filename)
				installArgs = append(installArgs, line)
			}
			packwiz(*flagPackDir, installArgs)
			previousLine = ""
		}
	}
}

func importFromString(importString string, autoConfirm bool) {
	data := strings.Split(importString, " ")
	for _, line := range data {
		line = strings.ReplaceAll(line, " ", "")
		line = strings.ReplaceAll(line, "	", "")
		if !strings.HasPrefix(line, "https://") {
			continue
		} else {
			installArgs := []string{}
			if autoConfirm {
				installArgs = append(installArgs, "-y")
			}
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
				installArgs = append(installArgs, filename)
				installArgs = append(installArgs, line)
				// TODO: replace this with importFromStringURL
				importFromStringURL(line)
			}
			packwiz(*flagPackDir, installArgs)
		}
	}
}

func importFromStringURL(url string) {
	fmt.Println("[PackWrap] [NOTICE] ImportFromStringURL functionality will be changing soon")
}
