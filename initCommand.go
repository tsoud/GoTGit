package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

func setupInitCommand() (*flag.FlagSet, *bool) {
	initCmd := flag.NewFlagSet("init", flag.ExitOnError)

	var quiet bool
	usage := "Only print error and warning messages; all other output will be suppressed."
	defaultVal := false
	initCmd.BoolVar(&quiet, "quiet", defaultVal, usage)
	initCmd.BoolVar(&quiet, "q", defaultVal, "(shorthand ver.) "+usage)

	return initCmd, &quiet
}

func initCmdHandler(quiet *bool) {
	for _, dir := range []string{".git", ".git/objects", ".git/refs"} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Printf("Error creating directory: %s\n", err)
		}
	}

	headFileContents := []byte("ref: refs/heads/master\n")
	if err := os.WriteFile(".git/HEAD", headFileContents, 0644); err != nil {
		log.Printf("Error writing file: %s\n", err)
	}

	if !*quiet {
		fmt.Println("Initialized git directory")
	}
}
