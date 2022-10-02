package main

import (
	"flag"
	"fmt"
	"io/fs"
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
	Version = "0.1.1"

	// flags
	flagHelp    = flag.Bool("h", false, "show help")
	flagVersion = flag.Bool("v", false, "show version")
	flagPackDir = flag.String("d", ".", "pack directory (when used with -b this is where the modpacks are located")
	flagBatch   = flag.Bool("b", false, "batch build")

	flagRefresh = flag.Bool("r", false, "refresh modpack after operations")

	// import file
	flagImport  = flag.String("i", "", "import links from file")
	flagMetaDir = flag.String("m", "", "meta directory (when using the import flag)")
	flagConfirm = flag.Bool("y", false, "auto confirm (when using the import flag)")
	flagSide    = flag.Bool("c", false, "client side mod (when using the import flag)")

	args []string
)

func main() {
	flag.Parse()
	args = flag.Args()

	if _, err := exec.LookPath("packwiz"); err != nil {
		fmt.Println("[PackWrap] \n[ERROR] packwiz is not installed,\nplease install it with 'go install github.com/packwiz/packwiz@latest'")
		return
	}

	if !strings.HasSuffix(*flagPackDir, "/") {
		*flagPackDir += "/"
	}
	fmt.Println("[PackWrap] PackDir:", *flagPackDir)

	if *flagVersion {
		fmt.Println("[PackWrap] version:", Version)
		return
	}
	if *flagHelp {
		//fmt.Println("[PackWrap]")
		flag.Usage()
		fmt.Println("")
		packwiz(*flagPackDir, []string{"help"})
		return
	}
	if *flagImport != "" {
		if *flagBatch {
			fmt.Println("[PackWrap] [ERROR] -b and -i conflict")
			return
		}
		importFromFile()
		return
	}

	if *flagBatch {
		fmt.Println("[PackWrap] Batch mode")
		batchMode(*flagPackDir, args)
	} else {
		packwiz(*flagPackDir, args)
		if *flagRefresh {
			packwiz(*flagPackDir, []string{"refresh"})
		}
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
			installArgs := []string{}
			metaDir := "mods"

			if strings.Contains(line, "modrinth.com/mod/") {
				installArgs = append(installArgs, "add", "mr")
			} else if strings.Contains(line, "curseforge.com/minecraft") {
				installArgs = append(installArgs, "add", "cf")
			} else {
				fmt.Println("[ERROR] unknown host", line)
				continue
			}
			if strings.Contains(line, "curseforge.com/minecraft/texture-packs/") {
				installArgs = append(installArgs, "--category", "texture-packs")
				metaDir = "resourcepacks"
			}
			if *flagMetaDir != "" {
				installArgs = append(installArgs, "--meta-dir", *flagMetaDir)
				metaDir = *flagMetaDir
			}
			if *flagConfirm {
				installArgs = append(installArgs, "-y")
			}

			installArgs = append(installArgs, line)
			packwiz(*flagPackDir, installArgs)

			if *flagSide {
				m_file := filepath.Join(*flagPackDir, metaDir, fmt.Sprintf("%s.pw.toml", line))
				f, err := ioutil.ReadFile(m_file)
				if err != nil {
					fmt.Println("Cannot read file: ", m_file)
					continue
				} else {
					ioutil.WriteFile(m_file, []byte(strings.Replace(string(f), "\"side\": \"both\"", "\"side\": \"client\"", -1)), fs.FileMode(os.O_WRONLY))
				}
			}

		}
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
