package core

import (
	"bytes"
	"fmt"
	"time"
	"strings"
	"github.com/NiHaLOO7/codestash/storage"
	"encoding/hex"
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
	content := ParseObject(data)
	index := make(map[string]string)
	i := 0
	for i < len(content) {
		spacePos := bytes.IndexByte(content[i:], ' ')
		 // _ := content[i:i+spacePos]
		i += spacePos+1
		nullPos := bytes.IndexByte(content[i:],0)
		filename := string(content[i : i + nullPos])
		i += nullPos + 1
		rawHash := content[i : i+20]
		hexHash := hex.EncodeToString(rawHash)
		index[filename] = hexHash
		i += 20
	}
	return index
}

// Tree Hash
func CreateTree(index map[string]string) string {
	var buf bytes.Buffer
	for filename, hash := range index {
		buf.WriteString("100644 ");
		buf.WriteString(filename)
		buf.WriteByte(0)
		rawBytes, _ := hex.DecodeString(hash)
		buf.Write(rawBytes)
	}
	hash, data := HashContent(buf.Bytes(), "tree")
	storage.WriteObject(hash, data)
	return hash
}

func CreateCommit(treeHash, parent, message string) string {
	content := "tree " + treeHash + "\n"
	if parent != "" {
		content += "parent " + parent + "\n"
	}
	config := ReadConfig()
	timezone := time.Now().Format("-0700")
	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	author := config["name"] + " <" + config["email"] + "> " + timestamp + " " + timezone
	content += "author "+author+ "\n"
	content += "committer " + author + "\n"
	content += "\n" + message + "\n"
	hash, data := HashContent([]byte(content), "commit")
	storage.WriteObject(hash, data)
	return hash
}