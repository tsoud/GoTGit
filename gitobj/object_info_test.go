package gitobj

import (
	"bytes"
	"io"
	"log"
	"os"
	"regexp"
	"testing"
)

type headerTestUnit struct {
	testName    string
	objHash     string
	wantObjType string
	wantObjSize int
}

func TestMain(m *testing.M) {
	os.Chdir("/home/tamer/go_projects/GoTGit")
	os.Exit(m.Run())
}

var testBlobFile = "test_blob.txt"

var testTreeFiles = map[string]string{
	"default":   "test_tree.txt",
	"name-only": "test_tree_nameonly.txt",
	"long":      "test_tree_long.txt",
	"l":         "test_tree_long.txt",
}

// var testCommitFile = "test_commit.txt"

func readTestFile(file string) string {
	contents, err := os.ReadFile(file)
	if err != nil {
		log.Fatal(err)
	}
	return string(contents)
}

func TestProcessObjHeader(t *testing.T) {
	tests := []headerTestUnit{
		{
			"test_blob1",
			"8e852b3d9aa0ff658deecf04d7c4c53f35077ad3",
			"blob",
			46,
		},
		{
			"test_blob2",
			"20e89f06b3d2bbbadea4f6e6b9dd47cc1b6afd70",
			"blob",
			889,
		},
		{
			"test_tree",
			"0ade50c56e62ba62260417cffd1a56844a4e5835",
			"tree",
			170,
		},
		{
			"test_commit",
			"83e0e8dbd81beb8b42f15c979a36c6c7d21d6b79",
			"commit",
			891,
		},
	}

	for _, test := range tests {
		gotObj, err := GitObjInfoFromHash(test.objHash)
		if err != nil {
			t.Fatal(err)
		}
		t.Run(test.testName, func(t *testing.T) {
			gotObj.ProcessObjHeader()
			if gotObj.Type != test.wantObjType || gotObj.Size != test.wantObjSize {
				t.Errorf("Wanted type: %q and size: %q, got %q and %q",
					test.wantObjType, test.wantObjSize, gotObj.Type, gotObj.Size)
			}
		})
	}
}

func TestPrintContent(t *testing.T) {
	testBlob := readTestFile(testBlobFile)

	reHash := regexp.MustCompile(`(?m)^[a-z\d]{40}`)
	hashes := reHash.FindAllString(testBlob, -1)

	reBody := regexp.MustCompile(`>>>((?:.|\n)*?)<<<`)
	blobs := reBody.FindAllStringSubmatch(testBlob, -1)

	for i, blob := range blobs {
		bObj, err := GitObjInfoFromHash(hashes[i])
		if err != nil {
			t.Fatal(err)
		}

		// Capture text from stdout
		clear := os.Stdout
		r, w, err := os.Pipe()
		if err != nil {
			t.Fatal(err)
		}
		os.Stdout = w
		err = bObj.PrintContent()
		if err != nil {
			t.Fatal(err)
		}
		w.Close()
		os.Stdout = clear

		var buf bytes.Buffer
		io.Copy(&buf, r)
		if res := buf.String(); res != blob[1] {
			t.Errorf("Wanted:\n%q\n\nGot:\n%q", blob[1], res)
		}
	}
}

func TestPrintTreeContent(t *testing.T) {
	for outType, testfile := range testTreeFiles {
		testTree := readTestFile(testfile)

		reHash := regexp.MustCompile(`(?m)^[a-z\d]{40}`)
		hashes := reHash.FindAllString(testTree, -1)

		reBody := regexp.MustCompile(`>>>((?:.|\n)*?)<<<`)
		trees := reBody.FindAllStringSubmatch(testTree, -1)

		for i, tree := range trees {
			tObj, err := GitObjInfoFromHash(hashes[i])
			if err != nil {
				t.Fatal(err)
			}

			// Capture text from stdout
			clear := os.Stdout
			r, w, err := os.Pipe()
			if err != nil {
				t.Fatal(err)
			}
			os.Stdout = w
			err = tObj.PrintTreeContent(outType)
			if err != nil {
				t.Fatal(err)
			}
			w.Close()
			os.Stdout = clear

			var buf bytes.Buffer
			io.Copy(&buf, r)
			if res := buf.String(); res != tree[1] {
				t.Errorf("\nWanted:\n%q\nGot:\n%q\n---\n", tree[1], res)
			}
		}
	}
}
