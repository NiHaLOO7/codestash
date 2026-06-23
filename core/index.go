package core

import (
	"fmt"
	"os"
	"strings"
)
const PARENT = ".codestash"
const FILE_PATH = PARENT + "/index"

func ReadIndex() map[string]string {
	_, err := os.Stat(PARENT)
	if err != nil {
		fmt.Println("Git not initialized")
        return nil
	}
	file, err := os.ReadFile(FILE_PATH)
	hashmap := make(map[string]string)
	if err != nil {
        return hashmap
	}
	lines := strings.Split(string(file), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
        	continue
    	}
		parts := strings.Split(line, " ")
		hashmap[parts[1]] = parts[0]
	}
	return hashmap
}

func WriteIndex(index map[string]string) {
	_, err := os.Stat(PARENT)
	if err != nil {
		fmt.Println("Git not initialized")
        return
	}
	var content string
	for key, value := range index {
		content += value + " " + key + "\n"
	}

	os.WriteFile(FILE_PATH, []byte(content), 0644)
}

func AddToIndex(filename string, hash string) {
	index := ReadIndex()
	index[filename] = hash
	WriteIndex(index)
}


