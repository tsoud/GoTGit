package gitobj

import (
	"bufio"
	"bytes"
	"compress/zlib"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path"
	"strconv"
	"strings"
)

const GitBaseDir = ".gotgit_test/objects"

type GitObject struct {
	Hash    string
	Name    string
	Type    string
	Mode    string
	Size    int
	Content []byte
}

func GetObjectMode(mode string) string {
	// Translate FileInfo to Git file mode convention
	switch {
	case strings.HasPrefix(mode, "d"):
		return "040000"
	case strings.HasPrefix(mode, "L"):
		return "120000"
	case strings.ContainsRune(mode, 'x'):
		return "100755"
	default:
		return "100644"
	}
}

func typeIsValid(objType string) bool {
	for _, validType := range []string{"blob", "tree", "commit", "tag"} {
		if validType == objType {
			return true
		}
	}
	return false
}

func ReadGitObj(objectHash string) (*GitObject, error) {
	// Read the object type and contents from a given hash.
	fullPath := path.Join(GitBaseDir, objectHash[:2], objectHash[2:])

	src, err := os.Open(fullPath)
	if err != nil {
		return nil, fmt.Errorf("could not open object %s: %s", objectHash, err)
	}
	defer src.Close()

	contents, err := zlib.NewReader(src)
	if err != nil {
		src.Close()
		return nil, fmt.Errorf("error decompressing %s: %s", objectHash, err)
	}
	defer contents.Close()

	contentBuffer := bufio.NewReader(contents)

	// Read the object header. Note that properly formatted objects must contain
	// a null byte between the header and body.
	headerBytes, err := contentBuffer.ReadBytes(0)
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("error reading header of %s: %s", objectHash, err)
	}
	// Drop the null byte.
	header := string(headerBytes[:len(headerBytes)-1])
	// Object headers should have the format "<object type> <size>".
	headerParts := strings.Split(header, " ")
	if len(headerParts) < 2 {
		return nil, fmt.Errorf("malformed header: \"%s\"", header)
	}
	if !typeIsValid(headerParts[0]) {
		return nil, fmt.Errorf("invalid type \"%s\" found for object %s", headerParts[0], objectHash)
	}
	objType := headerParts[0]
	objSize, err := strconv.Atoi(headerParts[1])
	if err != nil {
		return nil, fmt.Errorf("could not convert object size %v: %s", headerParts[1], err)
	}

	// Read the object contents after validating header.
	objContent, err := io.ReadAll(contentBuffer)
	if err != nil {
		return nil, fmt.Errorf("could not read contents of %s: %s", objectHash, err)
	}

	return &GitObject{
		Hash:    objectHash,
		Type:    objType,
		Size:    objSize,
		Content: objContent,
	}, nil
}

func PrintBlob(gitobj *GitObject) error {
	contentReader := bytes.NewReader(gitobj.Content)

	io.Copy(os.Stdout, contentReader)
	return nil
}

func PrintTree(gitobj *GitObject, outputType string) error {
	if gitobj.Type != "tree" {
		return fmt.Errorf("%s is not a tree object", gitobj.Hash)
	}

	contentReader := bytes.NewReader(gitobj.Content)
	treeBuf := new(bytes.Buffer)
	if _, err := io.Copy(treeBuf, contentReader); err != nil {
		return fmt.Errorf("could not read tree contents: %s", err)
	}

	var output strings.Builder

	for treeBuf.Len() > 0 {
		mode, err := treeBuf.ReadString(' ')
		if err != nil && err != io.EOF {
			return fmt.Errorf("unable to extract mode: %s", err)
		}
		// pad mode string if leading zero is omitted by Git
		mode = fmt.Sprintf("%06s", mode[:len(mode)-1])

		objName, err := treeBuf.ReadBytes(0)
		if err != nil {
			return fmt.Errorf("unable to read object name: %s", err)
		}
		objName = objName[:len(objName)-1]
		if outputType == "name-only" {
			output.WriteString(fmt.Sprintf("%s\n", string(objName)))
		}

		sha1Hash := make([]byte, 20)
		if _, err := treeBuf.Read(sha1Hash); err != nil {
			return fmt.Errorf("unable to read SHA1 hash: %s", err)
		}

		objHash := hex.EncodeToString(sha1Hash)
		objInfo, err := ReadGitObj(objHash)
		if err != nil {
			return fmt.Errorf("%s", err)
		}
		if outputType == "default" {
			output.WriteString(fmt.Sprintf("%s %s %s\t%s\n", mode, objInfo.Type, objInfo.Hash, objName))
		}

		size := strconv.Itoa(objInfo.Size)
		if objInfo.Type == "tree" {
			size = "-"
		}
		// add extra padding for larger sizes if needed
		width := 7
		if width < len(size) {
			width = len(size)
		}
		if outputType == "long" || outputType == "l" {
			output.WriteString(fmt.Sprintf("%s %s %s %*s\t%s\n",
				mode, objInfo.Type, objInfo.Hash, width, size, objName))
		}
	}

	fmt.Print(output.String())
	return nil
}

func (blob *GitObject) Write() error {
	dstDirPath := path.Join(GitBaseDir, blob.Hash[:2])
	if err := os.MkdirAll(dstDirPath, 0755); err != nil {
		return fmt.Errorf("error creating object subdirectory in .git: %s", err)
	}
	dstFilePath := path.Join(dstDirPath, blob.Hash[2:])
	dst, err := os.Create(dstFilePath)
	if err != nil {
		return fmt.Errorf("could not create object file: %s", err)
	}
	defer dst.Close()

	compressed := zlib.NewWriter(dst)
	defer compressed.Close()

	if _, err := compressed.Write(blob.Content); err != nil {
		return fmt.Errorf("could not compress object: %s", err)
	}

	return nil
}
