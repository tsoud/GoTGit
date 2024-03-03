package gitobj

import (
	"bufio"
	"bytes"
	"cmp"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path"
	"slices"
	"strings"
)

type Tree struct {
	Name    string
	Parent  *Tree
	Objects []*GitObject
}

func ignore(ignoreFile, rootDir string) (map[string]bool, error) {
	// Read a file (e.g. .gitignore) with patterns to skip when creating trees.
	// `ignoreFile` specifies the patterns to skip (an empty string means skip nothing)
	// and `rootDir` specifies the directory to which the patterns should be applied.
	ignore := make(map[string]bool)

	if ignoreFile == "" {
		return ignore, nil
	}

	fsys := os.DirFS(rootDir)
	f, err := os.Open(ignoreFile)

	if err != nil {
		return nil, fmt.Errorf("could not access %s: %s", ignoreFile, err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	var ignoredFiles []string

	for scanner.Scan() {
		files, err := fs.Glob(fsys, scanner.Text())
		if err != nil {
			return nil, fmt.Errorf("error reading patterns in %s: %s", ignoreFile, err)
		}
		ignoredFiles = append(ignoredFiles, files...)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("could not read %s contents: %s", ignoreFile, err)
	}

	for _, f := range ignoredFiles {
		ignore[path.Join(rootDir, f)] = true
	}

	return ignore, nil
}

func HashTree(name string, objects []*GitObject) (*GitObject, error) {
	// Create a tree object from an ordered list of Git objects.
	treeObj := &GitObject{}
	treeObj.Name = name

	var content strings.Builder
	for _, obj := range objects {
		objHash, err := hex.DecodeString(obj.Hash)
		if err != nil {
			return nil, fmt.Errorf(
				"error decoding SHA1 hash of %s (%s):\n%s", obj.Name, obj.Name, err)
		}
		content.WriteString(
			fmt.Sprintf("%s %s\u0000%s", obj.Mode, obj.Name, string(objHash)))
	}

	contentBytes := []byte(content.String())
	treeObj.Size = len(contentBytes)
	treeObj.Type = "tree"
	treeObj.Mode = "040000"
	header := fmt.Sprintf("%s %d\u0000", treeObj.Type, treeObj.Size)
	treeObj.Content = bytes.Join([][]byte{[]byte(header), contentBytes}, []byte(""))
	hash := sha1.Sum(treeObj.Content)
	treeObj.Hash = hex.EncodeToString(hash[:])

	return treeObj, nil
}

func makeTree(
	rootDir string, rootTree *Tree, ignoredFiles map[string]bool, write bool,
) (*GitObject, error) {
	// Build a tree object recursively and optionally write out all the items in the tree
	// and its sub-trees. Returns the root-level tree object.
	files, err := fs.ReadDir(os.DirFS(rootDir), ".")
	if err != nil {
		log.Fatalf("cannot create tree from %s: %s", rootDir, err)
	}

	var fullPath string
	for _, file := range files {
		fullPath = path.Join(rootDir, file.Name())
		_, ok := ignoredFiles[fullPath]
		if ok {
			continue
		}
		if !file.IsDir() {
			blobObj, err := HashBlob(fullPath)
			if write {
				err := blobObj.Write()
				if err != nil {
					return nil, fmt.Errorf("error writing %s:\n%s", fullPath, err)
				}
			}
			if err != nil {
				return nil, fmt.Errorf("error hashing %s:\n%s", fullPath, err)
			}
			rootTree.Objects = append(rootTree.Objects, blobObj)
		} else {
			subTree := &Tree{Name: file.Name(), Parent: rootTree}
			treeObj, err := makeTree(fullPath, subTree, ignoredFiles, write)
			if err != nil {
				return nil, fmt.Errorf("error creating tree from %s:\n%s", fullPath, err)
			}
			subTree.Parent.Objects = append(subTree.Parent.Objects, treeObj)
		}
	}

	slices.SortFunc(rootTree.Objects, func(x, y *GitObject) int {
		return cmp.Compare(x.Name, y.Name)
	})

	treeObj, err := HashTree(path.Base(rootDir), rootTree.Objects)
	if err != nil {
		return nil, fmt.Errorf("error hashing tree %s:\n%s", rootDir, err)
	}
	if write {
		err := treeObj.Write()
		if err != nil {
			return nil, fmt.Errorf("error writing %s:\n%s", rootDir, err)
		}
	}

	return treeObj, nil
}

func WriteTree(rootDir, ignoreFile string, write bool) (*GitObject, error) {
	// Writes out a tree object starting from the given directory as the root.
	// If `write` is `true`, all objects in the tree and its sub-trees will be
	// written to the data store. Returns the root-level tree object.
	ignored, err := ignore(ignoreFile, rootDir)
	if err != nil {
		return nil, fmt.Errorf("cannot read `.ignore` file %s: %s", ignoreFile, err)
	}
	baseTree := &Tree{Name: path.Base(rootDir)}
	return makeTree(rootDir, baseTree, ignored, write)
}
