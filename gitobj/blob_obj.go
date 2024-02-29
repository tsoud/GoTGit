package gitobj

import (
	"crypto/sha1"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

func HashBlob(file string) (*GitObject, error) {
	// Create a blob object from a file.
	blobObj := &GitObject{}
	f, err := os.Open(file)
	if err != nil {
		log.Fatalf("Could not open file: %s\n", err)
	}
	defer f.Close()

	// Read the file size and create the header.
	info, err := f.Stat()
	if err != nil {
		return nil, fmt.Errorf("could not create header - error getting file information: %s", err)
	}

	// Process blob information:
	blobObj.Name = info.Name()
	blobObj.Type = "blob"
	blobObj.Mode = GetObjectMode(info.Mode().String())
	blobObj.Size = int(info.Size())
	header := fmt.Sprintf("%s %d\u0000", blobObj.Type, blobObj.Size)

	// prepend the header to the file contents to get SHA1 sum
	store := io.MultiReader(strings.NewReader(header), f)
	blobObj.Content, err = io.ReadAll(store)
	if err != nil {
		return nil, fmt.Errorf("error reading contents of %s: %s", file, err)
	}

	blobObj.Hash = fmt.Sprintf("%x", sha1.Sum(blobObj.Content))

	return blobObj, nil
}
