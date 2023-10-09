package cmd

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/tsoud/GoTGit.git/gitobj"
)

const CatFileUsageMsg = "usage: cat-file (-p | -t | -s) <object>\n"

func validCatFileFlags() []string {
	return []string{"s", "t", "p"}
}

func SetupCatFileCmd() *flag.FlagSet {
	catFileCmd := flag.NewFlagSet("cat-file", flag.ExitOnError)

	var pprint bool
	catFileCmd.BoolVar(&pprint, "p", false, "Pretty-print the contents of <object>.")
	var getType bool
	catFileCmd.BoolVar(&getType, "t", false, "Show object type identified by <object>.")
	var getSize bool
	catFileCmd.BoolVar(&getSize, "s", false, "Show the object size identified by <object>.")

	return catFileCmd
}

func validateCatFileFlags(fs *flag.FlagSet) {
	if fs.NFlag() == 0 {
		log.Fatalf("Missing required flag: `-p`, `-s`, or `-t`.\n%s", CatFileUsageMsg)
	}
	if fs.NFlag() > 1 {
		log.Fatalf(
			"`cat-file` takes only one flag: `-p`, `-s`, or `-t`.\n%s", CatFileUsageMsg)
	}
}

func catFileOption(fs *flag.FlagSet) (string, error) {
	validateCatFileFlags(fs)

	for _, flag := range validCatFileFlags() {
		if fs.Lookup(flag).Value.String() == "true" {
			return flag, nil
		}
	}

	return "", errors.New("invalid option")
}

func CatFileCmdHandler(file string, fs *flag.FlagSet) {
	fs.Parse(os.Args[2:])

	infoType, err := catFileOption(fs)
	if err != nil {
		log.Fatalf("%s", err)
	}

	catFile(file, infoType)
}

func catFile(file, infoType string) {
	objInfo, err := gitobj.GitObjInfoFromHash(file)
	if err != nil {
		log.Fatalf("error reading object %s: %s", file, err)
	}

	switch infoType {
	case "t":
		fmt.Printf("%s\n", objInfo.Type)
	case "s":
		fmt.Printf("%d\n", objInfo.Size)
	case "p":
		if err := objInfo.PrintContent(); err != nil {
			log.Fatalf("%s", err)
		}
	}
}
