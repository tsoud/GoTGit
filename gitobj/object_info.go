package gitobj

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"io"
	"os"
	"path"
	"strconv"
	"strings"
)

const GitBaseDir = ".git/objects"

type GitObjInfo struct {
	Hash      string
	Type      string
	Size      int
	HeaderLen int
}

func GitObjInfoFromHash(hash string) (*GitObjInfo, error) {
	g := &GitObjInfo{Hash: hash}
	if err := g.ProcessObjHeader(); err != nil {
		return nil, fmt.Errorf("%s", err)
	}

	return g, nil
}

func (g *GitObjInfo) ReadFromFile() (io.ReadCloser, *os.File, error) {
	fullPath := path.Join(GitBaseDir, g.Hash[:2], g.Hash[2:])

	src, err := os.Open(fullPath)
	if err != nil {
		return nil, nil, fmt.Errorf("could not open file: %s", err)
	}

	src_uncmpr, err := zlib.NewReader(src)
	if err != nil {
		src.Close()
		return nil, nil, fmt.Errorf("error decompressing file %s: %s", g.Hash, err)
	}

	return src_uncmpr, src, nil
}

func (g *GitObjInfo) ProcessObjHeader() error {
	headerBytes := make([]byte, 128)
	data, src, err := g.ReadFromFile()
	if err != nil {
		return fmt.Errorf("%s", err)
	}

	n, err := data.Read(headerBytes)
	if err != nil && err != io.EOF {
		return fmt.Errorf("error reading header of %s: %s", g.Hash, err)
	}
	defer src.Close()
	defer data.Close()

	idx := bytes.IndexByte(headerBytes[:n], 0)
	if idx == -1 {
		return fmt.Errorf("error reading header: no null byte found")
	}

	headerParts := strings.Split(string(headerBytes[:idx]), " ")
	if len(headerParts) != 2 {
		// TODO: Handle invalid object types
		return fmt.Errorf("error: could not extract type and size from header")
	}

	size, err := strconv.Atoi(headerParts[1])
	if err != nil {
		return fmt.Errorf("error getting object size: %s", err)
	}

	g.Size = size
	g.Type = headerParts[0]
	g.HeaderLen = idx

	return nil
}

func (g *GitObjInfo) PrintContent() error {
	content, src, err := g.ReadFromFile()
	if err != nil {
		return fmt.Errorf("%s", err)
	}
	defer src.Close()
	defer content.Close()

	buf := make([]byte, g.HeaderLen+1) //[]byte{}
	if _, err := content.Read(buf); err != nil && err != io.EOF {
		return fmt.Errorf("unable to read object contents: %s", err)
	}

	io.Copy(os.Stdout, content)
	return nil
}
