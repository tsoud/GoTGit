package main

import (
	"bufio"
	"compress/zlib"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path"
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

func setupCatFileCmd() (*flag.FlagSet, *bool) {
	catFileCmd := flag.NewFlagSet("cat-file", flag.ExitOnError)

	var pprint bool
	catFileCmd.BoolVar(&pprint, "p", false, "Pretty-print the contents of <object> based on its type.")

	return catFileCmd, &pprint
}

func catFileCmdHandler(pprint *bool, file string) {
	if !*pprint {
		log.Fatalf("Missing flag: `-p`\nThis flag is needed to print contents of <file>.")
	}

	baseFilePath := ".git/objects/"
	fullPath := path.Join(baseFilePath, file[:2], file[2:])

	source, err := os.Open(fullPath)
	if err != nil {
		log.Fatalf("Could not open file: %s", err)
	}
	defer source.Close()

	r, err := zlib.NewReader(source)
	if err != nil {
		log.Fatalf("Error when trying to decompress %s: %s", file, err)
	}
	defer r.Close()

	contents := bufio.NewReader(r)
	// discard header and null byte, print out contents
	if _, err := contents.ReadBytes(0); err != nil {
		log.Fatalf("Error reading header of %s: %s", file, err)
	}
	if _, err := io.Copy(os.Stdout, contents); err != nil {
		log.Fatalf("Error reading contents of %s: %s", file, err)
	}
}

// Usage: your_git.sh <command> <arg1> <arg2> ...
func main() {
	if len(os.Args) < 2 {
		log.Fatalf("\nusage: mygit <command> [<args>...]\n\n")
	}

	switch command := os.Args[1]; command {
	case "init":
		initCmdArgs, quiet := setupInitCommand()
		initCmdArgs.Parse(os.Args[2:])
		initCmdHandler(quiet)

	case "cat-file":
		catFileCmdArgs, pprint := setupCatFileCmd()
		catFileCmdArgs.Parse(os.Args[2:])
		files := catFileCmdArgs.Args()
		if len(files) < 1 {
			log.Fatal("No file given! Usage: mygit cat-file -p <file>")
		}
		if len(files) > 1 {
			log.Printf("More than one file given.\nOnly showing 1st file %s:", files[0])
		}
		catFileCmdHandler(pprint, files[0])

	default:
		fmt.Fprintf(os.Stderr, "Unknown command %s\n", command)
		os.Exit(1)
	}
}
