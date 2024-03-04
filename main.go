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
		catFileCmdArgs := cmd.SetupCatFileCmd()
		catFileCmdArgs.Parse(os.Args[2:])
		files := catFileCmdArgs.Args()
		if len(files) < 1 {
			log.Fatalf("No file given! %s", cmd.CatFileUsageMsg)
		}
		if len(files) > 1 {
			log.Printf("More than one file given.\nOnly showing 1st file %s:", files[0])
		}
		cmd.CatFileCmdHandler(files[0], catFileCmdArgs)

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

	case "ls-tree":
		lsTreeCmdArgs := cmd.SetupLSTreeCmd()
		lsTreeCmdArgs.Parse(os.Args[2:])
		file := lsTreeCmdArgs.Args()
		if len(file) < 1 {
			log.Fatalf("No file given! %s", cmd.LSTreeUsageMsg)
		}
		if len(file) > 1 {
			log.Printf("Found more than one file. Only displaying %s\n", file[0])
		}

		cmd.LSTreeCmdHandler(file[0], lsTreeCmdArgs)

	case "write-tree":
		writeTreeCmdArgs, ignore, prefix := cmd.SetupWriteTreeCmd()
		writeTreeCmdArgs.Parse(os.Args[2:])

		if err := cmd.WriteTreeCmdHandler(ignore, prefix, writeTreeCmdArgs); err != nil {
			log.Fatal(err)
		}

	default:
		fmt.Fprintf(os.Stderr, "Unknown command %s\n", command)
		os.Exit(1)
	}
}
