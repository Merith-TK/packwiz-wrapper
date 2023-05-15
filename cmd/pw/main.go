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
	Version = "0.2.0"

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

	switch args[0] {
	case "version":
		fmt.Println("PackWrap version", Version)
	case "help":
		flag.Usage()
	case "import":
		importFromFile(args[1])
	case "modlist":
		modlist()
	default:
		packwiz(*flagPackDir, args)
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
