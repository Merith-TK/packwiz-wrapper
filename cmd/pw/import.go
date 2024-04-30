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
	// import from string
	fmt.Println("[PackWrap] [NOTICE] importing from string is experimental and may not work as expected.")
	importFromString(args, autoConfirm)

}

// returns url, path, name
func parseLine(line string, previousLine string) (string, string, string) {
	if !strings.Contains(line, " ") {
		// get filename from url
		filename := strings.Split(line, "/")[len(strings.Split(line, "/"))-1]
		return line, previousLine, filename
	}
	parts := strings.Split(line, " ")
	// https://cdn.merith.xyz/icon.png /images/
	// return https://cdn.merith.xyz/images.png, /images/, icon.png
	if strings.HasSuffix(previousLine, "/") {
		url := parts[0]
		path := parts[1]
		name := strings.Split(url, "/")[len(strings.Split(url, "/"))-1]
		return url, path, name
	}
	// https://cdn.merith.xyz/icon.png /icon.png
	// return https://cdn.merith.xyz/images.png, /, icon.png
	url := parts[0]
	path := parts[1]
	pathParts := strings.Split(path, "/")
	if len(pathParts) > 1 {
		path = strings.Join(pathParts[:len(pathParts)-1], "/") + "/"
	}
	if path == "" {
		path = "/"
	}
	// get name from end of path
	name := pathParts[len(pathParts)-1]
	fmt.Println("[PackWrap] [NOTICE] `packwiz` does not support importing file urls to specific filenames yet.\n\tusing the filename from the url.")
	return url, path, name
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
		// clean formatting to make it easier to parse
		line = strings.TrimLeft(line, " ")
		line = strings.TrimLeft(line, "\t")
		// skip lines that are not urls or directories
		if !strings.HasPrefix(line, "https://") {
			if strings.HasPrefix(line, "/") {
				previousLine = line
			}
			continue
		} else {
			if importFromSource(line, autoConfirm) {
				continue
			}
			if importFromURL(line, previousLine, autoConfirm) {
				continue
			}

		}
	}
}

func importFromString(mods []string, autoConfirm bool) {
	source := mods[0]
	// verify source
	if source != "cf" && source != "mr" && source != "url" {
		fmt.Println("[PackWrap] [ERROR] invalid source")
		return
	}
	if source != "url" {
		for _, mod := range mods[1:] {
			if autoConfirm {
				packwiz(*flagPackDir, []string{source, "add", mod, "-y"})
			} else {
				packwiz(*flagPackDir, []string{source, "add", mod})
			}
		}
	} else {
		for _, mod := range mods[1:] {
			// get data from parseLine
			modSplit := strings.Split(mod, ",")
			url := modSplit[0]
			metapath := "/"
			if len(modSplit) > 1 {
				metapath = modSplit[1]
			}
			importFromURL(url, metapath, autoConfirm)
		}
	}

}

func importFromSource(source string, autoConfirm bool) bool {
	installArgs := []string{}
	switch {
	case strings.HasPrefix(source, "https://www.curseforge.com/"):
		installArgs = append(installArgs, "cf", "add", source)
	case strings.HasPrefix(source, "https://modrinth.com/"):
		installArgs = append(installArgs, "mr", "add", source)
	default:
		return false
	}
	if autoConfirm {
		installArgs = append(installArgs, "-y")
	}
	packwiz(*flagPackDir, installArgs)
	return true
}

func importFromURL(url string, targetPath string, autoConfirm bool) bool {
	installArgs := []string{}
	if autoConfirm {
		installArgs = append(installArgs, "-y")
	}

	installArgs = append(installArgs, "url", "add")
	// get data from parseLine
	url, metafolder, filename := parseLine(url, targetPath)

	installArgs = append(installArgs, "--meta-folder", metafolder)
	if filename == "" {
		filename = strings.Split(url, "/")[len(strings.Split(url, "/"))-1]
	} else {
		installArgs = append(installArgs, "--meta-name", filename)
	}
	installArgs = append(installArgs, filename, url)
	// TODO: manually override the filename in the pw.toml

	packwiz(*flagPackDir, installArgs)
	return true
}
