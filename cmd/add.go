package cmd

import (
	"os"
	"github.com/NiHaLOO7/codestash/core"
	"github.com/NiHaLOO7/codestash/storage"
)

func Add(filenames []string) {
	for _, filename := range filenames {
		content, _ := os.ReadFile(filename)
		hash, data := core.HashContent(content, "blob")
		index := core.ReadIndex()
		if index[filename] == hash {
    		continue   // file unchanged, skip
		}
		storage.WriteObject(hash, data)
		core.AddToIndex(filename, hash)
	}
}
