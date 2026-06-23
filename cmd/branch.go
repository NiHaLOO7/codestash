package cmd

import (
	"fmt"

	"github.com/NiHaLOO7/codestash/core"
)
func Branch(args []string) {
	if len(args) == 0 {
		branches := core.ListBranches()
		current := core.CurrentBranch()
		for _, branch := range branches {
			if branch == current {
				branch = "* "+branch
			}
			fmt.Println(branch)
		}
	} else {
		core.CreateBranch(args[0])
	}
}
