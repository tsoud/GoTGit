package cmd

import (
	"flag"
	"fmt"
	"log"

	"github.com/tsoud/GoTGit.git/gitobj"
)

func SetupHashObjectCmd() (*flag.FlagSet, *bool, *string) {
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
func HashObjectCmdHandler(write *bool, objType *string, file string) {
	if *objType != "blob" {
		log.Fatal("`hash-object` command only handles blob objects at this time")
	}

	gitObj, err := gitobj.HashBlob(file)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(gitObj.Hash)

	if *write {
		if err := gitObj.Write(); err != nil {
			log.Fatal(err)
		}
	}
}
