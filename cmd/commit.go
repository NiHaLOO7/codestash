package cmd

import (
	"fmt"
	"github.com/NiHaLOO7/codestash/core"
)

func Commit(args []string) {
	message := ""
	if args[0] == "-m" {
		message = args[1]
	}
	if message == "" {
		fmt.Println("Message required for commit")
		return
	}
	index := core.ReadIndex()
	if len(index) == 0 {
		fmt.Println("Nothing to commit")
		return
	}
	treeHash := core.CreateTree(index)
	parent := core.GetHead()
	commitHash := core.CreateCommit(treeHash, parent, message)
	core.UpdateHead(commitHash)
	fmt.Println(commitHash[:7])
}
