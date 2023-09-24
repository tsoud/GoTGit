package main

import (
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
