package cmd

import (
	"flag"
	"fmt"
	"os"
	"path"

	"github.com/tsoud/GoTGit.git/gitobj"
)

func SetupWriteTreeCmd() (*flag.FlagSet, *bool, *string) {
	writeTreeCmd := flag.NewFlagSet("write-tree", flag.ExitOnError)

	writeTreeCmd.Usage = func() {
		fmt.Fprintf(os.Stderr, "gotgit write-tree [--ignore][--prefix=<prefix>/]\n\nusage:\n")

		writeTreeCmd.PrintDefaults()
	}

	var ignore bool
	writeTreeCmd.BoolVar(&ignore, "ignore", false,
		"Ignore files or folders with patterns specified in `.gotgitignore`. This file should "+
			"be located in the root directory where `write-tree` is being called.",
	)
	var prefix string
	writeTreeCmd.StringVar(&prefix, "prefix", "",
		"Write a tree object for a subdirectory <prefix> in the project.",
	)

	return writeTreeCmd, &ignore, &prefix
}

func WriteTreeCmdHandler(ignore *bool, prefix *string, flags *flag.FlagSet) error {
	rootDir, err := os.Getwd()

	if err != nil {
		return err
	}

	ignoreFile := ""
	if *ignore {
		ignoreFile = path.Join(rootDir, ".gotgitignore")
		if _, err := os.Stat(ignoreFile); err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("no valid `.gotgitignore` file found in %s", rootDir)
			}
			return fmt.Errorf("error reading `.gotgitignore`: %s", err)
		}
	}

	if *prefix != "" {
		rootDir = path.Join(rootDir, *prefix)
		if _, err := os.Stat(rootDir); err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("subdirectory: %s does not exist", rootDir)
			}
			return fmt.Errorf("error processing %s: %s", rootDir, err)
		}
	}

	treeObj, err := gitobj.WriteTree(rootDir, ignoreFile, true)
	if err != nil {
		return fmt.Errorf("error writing tree: %s", err)
	}

	fmt.Println(treeObj.Hash)

	return nil
}
