package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func detectPackURL(localpath bool) {
	// detect pack url based off of pack.toml location and git remote
	// find pack.toml in current directory or child directories
	packLocation := ""
	err := filepath.Walk(*flagPackDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {

			fmt.Println(err)
			return nil
		}
		if strings.HasSuffix(path, "pack.toml") {
			if *flagPackDir != "." {
				path = strings.Replace(path, *flagPackDir, "", 1)
			}
			packLocation = path
			return nil
		}
		return nil
	})

	if err != nil {
		log.Fatal(err)
	}

	// if pack.toml is not found, exit
	if packLocation == "" {
		fmt.Println("pack.toml not found")
		return
	} else {
		// convert \ to /
		packLocation = strings.ReplaceAll(packLocation, "\\", "/")
	}
	if localpath {
		packLocation, _ = filepath.Abs(*flagPackDir + packLocation)
		packLocation = filepath.ToSlash(packLocation)
		packLocation = strings.Replace(packLocation, "/..minecraft", "/.minecraft", 1)
		fmt.Println(packLocation)
		return
	}
	// get git remote
	remote, err := exec.Command("git", "remote", "get-url", "origin").Output()
	if err != nil {
		log.Fatal(err)
	}
	remoteString := string(remote)
	remoteString = strings.TrimSuffix(remoteString, "\n")
	remoteString = strings.TrimSuffix(remoteString, ".git")
	// get branch name
	branch, err := exec.Command("git", "branch", "--show-current").Output()
	if err != nil {
		log.Fatal(err)
	}
	branchString := string(branch)
	branchString = strings.TrimSuffix(branchString, "\n")

	// split remote into parts
	remoteParts := strings.Split(remoteString, "/")

	part := remoteParts[2]
	urlString := ""
	switch part {
	case "github.com":
		// replace github.com with raw.githubusercontent.com
		remoteString = strings.Replace(remoteString, "github.com", "raw.githubusercontent.com", 1)
		urlString = remoteString + "/tree/" + branchString + "/" + packLocation
	case "gitlab.com":
		urlString = remoteString + "/-/raw/" + branchString + "/" + packLocation
	default:
		// fmt.Println("Unknown Git Remote Host:" + part + "\nAssuming Gitea.")
		urlString = remoteString + "/raw/branch/" + branchString + "/" + packLocation
	}
	fmt.Println(urlString)
}
