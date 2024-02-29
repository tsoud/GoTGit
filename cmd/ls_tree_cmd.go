package cmd

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/tsoud/GoTGit.git/gitobj"
)

const LSTreeUsageMsg = "usage: ls-tree [--long | -l] | [--name-only] <tree-ish>\n"

func validLSTreeFlags() []string {
	return []string{"name-only", "long", "l"}
}

func SetupLSTreeCmd() *flag.FlagSet {
	lsTreeCmd := flag.NewFlagSet("cat-file", flag.ExitOnError)

	var long bool
	lsTreeCmd.BoolVar(&long, "long", false, "List object size of blob (file) entries.")
	lsTreeCmd.BoolVar(&long, "l", false, "(shorthand ver.) List object size of blob (file) entries.")
	var nameOnly bool
	lsTreeCmd.BoolVar(&nameOnly, "name-only", false, "List only filenames "+
		"(instead of the \"long\" output), one per line.")

	return lsTreeCmd
}

func validateLSTreeFlags(fs *flag.FlagSet) error {
	if fs.NFlag() > 1 {
		return fmt.Errorf(
			"`ls-tree` takes only one flag: `--name-only`, `--long`, or `-l`.\n%s", LSTreeUsageMsg)
	}

	return nil
}

func lsTreeOption(fs *flag.FlagSet) (string, error) {
	if err := validateLSTreeFlags(fs); err != nil {
		return "", err
	}

	if fs.NFlag() == 0 {
		return "default", nil
	}

	for _, flag := range validLSTreeFlags() {
		if fs.Lookup(flag).Value.String() == "true" {
			return flag, nil
		}
	}

	return "", fmt.Errorf("invalid option - %s", LSTreeUsageMsg)
}

func LSTreeCmdHandler(objHash string, fs *flag.FlagSet) {
	fs.Parse(os.Args[2:])

	outputType, err := lsTreeOption(fs)
	if err != nil {
		log.Fatal(err)
	}

	treeObj, err := gitobj.ReadGitObj(objHash)
	if err != nil {
		log.Fatal(err)
	}

	if err := gitobj.PrintTree(treeObj, outputType); err != nil {
		log.Fatal(err)
	}
}
