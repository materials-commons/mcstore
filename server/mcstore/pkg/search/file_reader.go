package search

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/materials-commons/mcstore/pkg/app"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

const twoMeg = 2 * 1024 * 1024

var tikableMediaTypes map[string]bool = map[string]bool{
	"application/msword":                                                        true,
	"application/pdf":                                                           true,
	"application/rtf":                                                           true,
	"application/vnd.ms-excel":                                                  true,
	"application/vnd.ms-office":                                                 true,
	"application/vnd.ms-powerpoint":                                             true,
	"application/vnd.ms-powerpoint.presentation.macroEnabled.12":                true,
	"application/vnd.openxmlformats-officedocument.presentationml.presentation": true,
	"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":         true,
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document":   true,
	"application/vnd.sealedmedia.softseal.pdf":                                  true,
	"text/plain; charset=utf-8":                                                 true,
}

func ReadFileContents(fileID, mimeType, name string, size int64) string {
	switch mimeType {
	case "text/csv":
		//fmt.Println("Reading csv file: ", fileID, name, size)
		if contents, err := readCSVLines(fileID); err == nil {
			return contents
		}
	case "text/plain":
		if size > twoMeg {
			return ""
		}
		//fmt.Println("Reading text file: ", fileID, name, size)
		if contents, err := ioutil.ReadFile(app.MCDir.FilePath(fileID)); err == nil {
			return string(contents)
		}
	default:
		if _, ok := tikableMediaTypes[mimeType]; ok {
			if contents := extractUsingTika(fileID, mimeType, name, size); contents != "" {
				return contents
			}
		}
	}
	return ""
}

func readCSVLines(fileID string) (string, error) {
	if file, err := os.Open(app.MCDir.FilePath(fileID)); err == nil {
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			text := scanner.Text()
			if text != "" && !strings.HasPrefix(text, "#") {
				return text, nil
			}
		}
		//fmt.Println("readCSVLines no data")
		return "", errors.New("No data")
	} else {
		//fmt.Println("readCSVLines failed to open", err)
		return "", err
	}
}

func extractUsingTika(fileID, mimeType, name string, size int64) string {
	if size > twoMeg {
		return ""
	}

	out, err := exec.Command("tika.sh", "--text", app.MCDir.FilePath(fileID)).Output()
	if err != nil {
		fmt.Println("Tika failed for:", fileID, name, mimeType)
		fmt.Println("exec failed:", err)
		return ""
	}

	return string(out)
}
