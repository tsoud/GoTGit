package cmd

import (
	"flag"
	"log"

	"github.com/tsoud/GoTGit.git/gitobj"
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

	objInfo, err := gitobj.GitObjInfoFromHash(file)
	if err != nil {
		log.Fatalf("error reading object %s: %s", file, err)
	}

	if err := objInfo.PrintContent(); err != nil {
		log.Fatalf("%s", err)
	}
}
