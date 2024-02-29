package cmd

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/tsoud/GoTGit.git/gitobj"
)

const CatFileUsageMsg = "usage: cat-file (-p | -t | -s) <object>\n"

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

func validateCatFileFlags(fs *flag.FlagSet) error {
	if fs.NFlag() == 0 {
		return fmt.Errorf("missing required flag: `-p`, `-s`, or `-t`.\n%s", CatFileUsageMsg)
	}
	if fs.NFlag() > 1 {
		return fmt.Errorf(
			"`cat-file` takes only one flag: `-p`, `-s`, or `-t`.\n%s", CatFileUsageMsg)
	}

	return nil
}

func catFileOption(fs *flag.FlagSet) string {
	// Return a valid option for formatting output from the `cat-file` command.
	for _, flag := range []string{"s", "t", "p"} {
		if fs.Lookup(flag).Value.String() == "true" {
			return flag
		}
	}

	return ""
}

func catFile(objHash, outType string) {
	objInfo, err := gitobj.ReadGitObj(objHash)
	if err != nil {
		log.Fatalf("error reading object %s: %s", objHash, err)
	}

	switch outType {
	case "t":
		fmt.Printf("%s\n", objInfo.Type)
	case "s":
		fmt.Printf("%d\n", objInfo.Size)
	case "p":
		if err := gitobj.PrintBlob(objInfo); err != nil {
			log.Fatal(err)
		}
	}
}

func CatFileCmdHandler(objHash string, fs *flag.FlagSet) {
	fs.Parse(os.Args[2:])
	if err := validateCatFileFlags(fs); err != nil {
		log.Fatal(err)
	}

	outType := catFileOption(fs)
	catFile(objHash, outType)
}
