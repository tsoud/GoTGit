package cmd

import (
	"bufio"
	"compress/zlib"
	"flag"
	"io"
	"log"
	"os"
	"path"
)

func SetupCatFileCmd() (*flag.FlagSet, *bool) {
	catFileCmd := flag.NewFlagSet("cat-file", flag.ExitOnError)

	var pprint bool
	catFileCmd.BoolVar(&pprint, "p", false, "Pretty-print the contents of <object> based on its type.")

	return catFileCmd, &pprint
}

func CatFileCmdHandler(pprint *bool, file string) {
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
