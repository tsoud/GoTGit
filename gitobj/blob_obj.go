package gitobj

import (
	"compress/zlib"
	"crypto/sha1"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"strings"
)

type BlobObj struct {
	Hash   string
	Header string
	Body   []byte
}

func HashBlob(file string) (*BlobObj, error) {
	blobObj := &BlobObj{}
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
	blobObj.Header = fmt.Sprintf("blob %d\u0000", info.Size())

	// Prepend the header to the file contents and get the SHA1 sum
	store := io.MultiReader(strings.NewReader(blobObj.Header), f)
	blobObj.Body, err = io.ReadAll(store)
	if err != nil {
		log.Fatalf("Error reading contents of %s: %s\n", file, err)
	}

	blobObj.Hash = fmt.Sprintf("%x", sha1.Sum(blobObj.Body))

	return blobObj, nil
}

func (blob *BlobObj) WriteBlob() {
	dstDirPath := path.Join(GitBaseDir, blob.Hash[:2])
	if err := os.MkdirAll(dstDirPath, 0755); err != nil {
		log.Fatalf("Error creating object subdirectory in .git: %s\n", err)
	}
	dstFilePath := path.Join(dstDirPath, blob.Hash[2:])
	dst, err := os.Create(dstFilePath)
	if err != nil {
		log.Fatalf("Could not create object file: %s\n", err)
	}
	defer dst.Close()

	compressed := zlib.NewWriter(dst)
	defer compressed.Close()

	if _, err := compressed.Write(blob.Body); err != nil {
		log.Fatalf("Could not compress object: %s\n", err)
	}
}
