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

var (
	Version = "0.3.0"

	// flags
	flagHelp = flag.Bool("h", false, "show help")

	flagRefresh = flag.Bool("r", false, "refresh modpack after operations")

	flagConfirm = flag.Bool("y", false, "auto confirm (when using the import flag)")
	flagSide    = flag.Bool("c", false, "client side mod (when using the import flag)")

	flagPackDir = flag.String("d", ".", "pack directory")

	args []string
)

func main() {
	flag.Parse()
	args = flag.Args()

	if _, err := exec.LookPath("packwiz"); err != nil {
		fmt.Println("[PackWrap] \n[ERROR] packwiz is not installed,\nplease install it with 'go install github.com/packwiz/packwiz@latest'")
		return
	}

	if *flagHelp {
		flag.Usage()
		return
	}

	// usage
	// pw modlist -raw (optional)
	// pw import -i import.txt -y (optional)
	// pw reinstall -y (optional)
	// pw batch <command>

	// if there are no arguments, show help
	if len(args) == 0 {
		flag.Usage()
		println()
		packwiz(*flagPackDir, []string{})
		return
	}

	switch args[0] {
	case "version":
		fmt.Println("PackWrap version", Version)
	case "help":
		flag.Usage()
	case "import":
		importFromFile(args[1])
	case "modlist":
		modlist()
	case "reinstall":
		reinstall()
	case "batch":
		batchMode(*flagPackDir, args[1:])
	default:
		packwiz(*flagPackDir, args)
	}

}

func batchMode(dir string, args []string) {
	// get all folders in pack dir
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		fmt.Println("[ERROR]\n", err)
		os.Exit(1)
	}
	for _, file := range files {
		if file.IsDir() {
			// get filepath
			filePath := filepath.Join(dir, file.Name())
			filePath = strings.ReplaceAll(filePath, "\\", "/") + "/"
			packwiz(filePath, args)
			if *flagRefresh {
				packwiz(filePath, []string{"refresh"})
			}
		}
	}
}
