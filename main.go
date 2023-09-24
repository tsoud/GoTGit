package main

import (
	"bufio"
	"compress/zlib"
	"crypto/sha1"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"strings"
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

func setupHashObjectCmd() (*flag.FlagSet, *bool, *string) {
	hashObjCmd := flag.NewFlagSet("hash-object", flag.ExitOnError)

	var write bool
	hashObjCmd.BoolVar(&write, "w", false, "Actually write the object into the object database.")

	var objType string
	hashObjCmd.StringVar(&objType, "t", "blob",
		"Specify the type of object to be created (default: \"blob\"). "+
			"Possible values are `commit`, `tree`, `blob`, and `tag`.")

	return hashObjCmd, &write, &objType
}

// TODO: Add functionality for handling non-blob types
func hashObjectCmdHandler(write *bool, objType *string, file string) {
	f, err := os.Open(file)
	if err != nil {
		log.Fatalf("Could not open file: %s\n", err)
	}
	defer f.Close()

	// Read the filesize and create the header
	info, err := f.Stat()
	if err != nil {
		log.Fatalf("Could not create header. Error getting file information: %s\n", err)
	}
	header := fmt.Sprintf("%s %d\u0000", *objType, info.Size())

	// Prepend the header to the file contents and get the SHA1 sum
	store := io.MultiReader(strings.NewReader(header), f)
	buf, err := io.ReadAll(store)
	if err != nil {
		log.Fatalf("Error reading contents of %s: %s\n", file, err)
	}

	objHash := fmt.Sprintf("%x", sha1.Sum(buf))
	fmt.Println(objHash)

	if *write {
		dstDirPath := path.Join(".git/objects", objHash[:2])
		if err := os.MkdirAll(dstDirPath, 0755); err != nil {
			log.Fatalf("Error creating object subdirectory in .git: %s\n", err)
		}
		dstFilePath := path.Join(dstDirPath, objHash[2:])
		dst, err := os.Create(dstFilePath)
		if err != nil {
			log.Fatalf("Could not create object file: %s\n", err)
		}
		defer dst.Close()

		compressed := zlib.NewWriter(dst)
		defer compressed.Close()

		if _, err := compressed.Write(buf); err != nil {
			log.Fatalf("Could not compress object: %s\n", err)
		}
	}
}

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

	case "hash-object":
		hashObjCmdArgs, write, objType := setupHashObjectCmd()
		hashObjCmdArgs.Parse(os.Args[2:])
		file := hashObjCmdArgs.Args()
		if len(file) < 1 {
			log.Fatal("No file given! Usage: mygit hash-object [-w] [-t <type>] <file>")
		}
		if len(file) > 1 {
			log.Printf("Found more than one file. Only hashing %s\n", file[0])
		}
		hashObjectCmdHandler(write, objType, file[0])

	default:
		fmt.Fprintf(os.Stderr, "Unknown command %s\n", command)
		os.Exit(1)
	}
}
