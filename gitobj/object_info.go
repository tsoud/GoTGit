package gitobj

import (
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

type GitObjInfo struct {
	Hash      string
	Type      string
	Size      int
	HeaderLen int
}

func GitObjInfoFromHash(hash string) (*GitObjInfo, error) {
	g := &GitObjInfo{Hash: hash}
	if err := g.ProcessObjHeader(); err != nil {
		return nil, err
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
		return err
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

func (g *GitObjInfo) GetContent() (io.ReadCloser, error) {
	content, src, err := g.ReadFromFile()
	if err != nil {
		return nil, err
	}
	defer src.Close()

	buf := make([]byte, g.HeaderLen+1)
	if _, err := content.Read(buf); err != nil && err != io.EOF {
		return nil, fmt.Errorf("unable to read object contents: %s", err)
	}

	return content, nil
}

func (g *GitObjInfo) PrintContent() error {
	content, err := g.GetContent()
	if err != nil {
		return err
	}
	defer content.Close()

	io.Copy(os.Stdout, content)
	return nil
}

func (g *GitObjInfo) PrintTreeContent(outputType string) error {
	content, err := g.GetContent()
	if err != nil {
		return err
	}
	defer content.Close()

	treeBuf := new(bytes.Buffer)
	if _, err := io.Copy(treeBuf, content); err != nil {
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

		filename, err := treeBuf.ReadBytes(0)
		if err != nil {
			return fmt.Errorf("unable to read filename: %s", err)
		}
		filename = filename[:len(filename)-1]
		if outputType == "name-only" {
			output.WriteString(fmt.Sprintf("%s\n", string(filename)))
		}

		sha1Hash := make([]byte, 20)
		if _, err := treeBuf.Read(sha1Hash); err != nil {
			return fmt.Errorf("unable to read SHA1 hash: %s", err)
		}

		objHash := hex.EncodeToString(sha1Hash)
		gitObj, err := GitObjInfoFromHash(objHash)
		if err != nil {
			return fmt.Errorf("%s", err)
		}
		if outputType == "default" {
			output.WriteString(fmt.Sprintf("%s %s %s\t%s\n", mode, gitObj.Type, gitObj.Hash, filename))
		}

		size := strconv.Itoa(gitObj.Size)
		if gitObj.Type == "tree" {
			size = "-"
		}
		// add extra padding for larger sizes if needed
		width := 7
		if width < len(size) {
			width = len(size)
		}
		if outputType == "long" || outputType == "l" {
			output.WriteString(fmt.Sprintf("%s %s %s %*s\t%s\n",
				mode, gitObj.Type, gitObj.Hash, width, size, filename))
		}
	}

	fmt.Print(output.String())
	return nil
}
