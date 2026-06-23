package cmd

import (
	"fmt"
	"os"

	"github.com/NiHaLOO7/codestash/core"
	"github.com/NiHaLOO7/codestash/storage"
)


func Checkout(args []string) {
	name := args[0]
	if name == "-b" && len(args) > 1 {
		name = args[1]
		core.CreateBranch(name)
	}
	core.SwitchBranch(name)
	head := core.GetHead()
	commitData, _ := storage.ReadObject(head) 
	commit := core.ParseCommit(commitData)
	treeData, _ := storage.ReadObject(commit.Tree)
	tree := core.ParseTree(treeData) 
	for filename, blobHash := range tree {
		data, _ := storage.ReadObject(blobHash)
		content := core.ParseObject(data)
		os.WriteFile(filename, content, 0644)
	}
	fmt.Printf("Switched to branch '%s'\n", name)
}