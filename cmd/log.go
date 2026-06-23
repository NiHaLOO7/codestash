package cmd

import (
	"fmt"
	"github.com/NiHaLOO7/codestash/storage"
	"github.com/NiHaLOO7/codestash/core"
)

func Log() {
	head := core.GetHead()
	for head != "" {
		data, _ := storage.ReadObject(head)
		commit := core.ParseCommit(data)
		fmt.Printf("commit %s\n", head)
        fmt.Printf("Author: %s\n", commit.Author)
        fmt.Printf("\n    %s\n\n", commit.Message)
		head = commit.Parent
	}
}
