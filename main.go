package main

import (
	"fmt"
	"log"
	"os"

	"github.com/tsoud/GoTGit.git/cmd"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("\nusage: mygit <command> [<args>...]\n\n")
	}

	switch command := os.Args[1]; command {
	case "init":
		initCmdArgs, quiet := cmd.SetupInitCommand()
		initCmdArgs.Parse(os.Args[2:])
		cmd.InitCmdHandler(quiet)

	case "cat-file":
		catFileCmdArgs, pprint := cmd.SetupCatFileCmd()
		catFileCmdArgs.Parse(os.Args[2:])
		files := catFileCmdArgs.Args()
		if len(files) < 1 {
			log.Fatal("No file given! Usage: mygit cat-file -p <file>")
		}
		if len(files) > 1 {
			log.Printf("More than one file given.\nOnly showing 1st file %s:", files[0])
		}
		cmd.CatFileCmdHandler(pprint, files[0])

	case "hash-object":
		hashObjCmdArgs, write, objType := cmd.SetupHashObjectCmd()
		hashObjCmdArgs.Parse(os.Args[2:])
		file := hashObjCmdArgs.Args()
		if len(file) < 1 {
			log.Fatal("No file given! Usage: mygit hash-object [-w] [-t <type>] <file>")
		}
		if len(file) > 1 {
			log.Printf("Found more than one file. Only hashing %s\n", file[0])
		}
		cmd.HashObjectCmdHandler(write, objType, file[0])

	default:
		fmt.Fprintf(os.Stderr, "Unknown command %s\n", command)
		os.Exit(1)
	}
}
