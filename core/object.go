package core

import (
	"bytes"
	"fmt"
	"time"
	"strings"
	"github.com/NiHaLOO7/codestash/storage"
)


type CommitData struct {
    Tree    string
    Parent  string
    Author  string
    Message string
}

func ParseObject(blob []byte) []byte {
	pos := bytes.IndexByte(blob, 0)
	return blob[pos + 1: ]
}

func ParseCommit(data []byte) CommitData {
	content := string(ParseObject(data))
	lines := strings.Split(content, "\n")

	commit := CommitData{}
	messageStart := false

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if messageStart {
            commit.Message += line
            continue
        }
		if strings.HasPrefix(line, "tree "){
			commit.Tree = strings.TrimPrefix(line, "tree ")
		} else if strings.HasPrefix(line, "parent ") {
			commit.Parent = strings.TrimPrefix(line, "parent ")
		} else if strings.HasPrefix(line, "author ") {
			commit.Author = strings.TrimPrefix(line, "author ")
		} else if line == "" {
			messageStart = true
		}
	}
	return commit
}

func ParseTree(data []byte) map[string]string {
	content := string(ParseObject(data))
	lines := strings.Split(content, "\n")
	index := make(map[string]string)
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == ""{
			continue
		}
		parts := strings.Split(line, " ")
		index[parts[2]] = parts[1]
	}
	return index
}

// Tree Hash
func CreateTree(index map[string]string) string {
	var content string
	for filename, hash := range index {
		content += "blob " + hash + " " + filename + "\n"
	}
	hash, data := HashContent([]byte(content), "tree")
	storage.WriteObject(hash, data)
	return hash
}

func CreateCommit(treeHash, parent, message string) string {
	content := "tree " + treeHash + "\n"
	if parent != "" {
		content += "parent " + parent + "\n"
	}
	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	content += "author CodeStash User " + timestamp + "\n"
	content += "\n" + message + "\n"
	hash, data := HashContent([]byte(content), "commit")
	storage.WriteObject(hash, data)
	return hash
}