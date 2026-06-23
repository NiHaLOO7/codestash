package cmd

import (
	"github.com/NiHaLOO7/codestash/core"
	"fmt"
	"os"
	"path/filepath"
)

var ignore = []string{".codestash", "exercises", "cmd", "core", "storage", ".gitignore", "go.mod", "go.sum", "main.go", "test.txt"}

func Status() {
	branch := core.CurrentBranch()
	var untracked []string
	var modified []string
	index := core.ReadIndex()
	filepath.Walk(".", func(path string, info os.FileInfo, err error) error {

		if info.IsDir() {
			for _, name := range ignore {
        		if path == name {
            		return filepath.SkipDir
        		}
    		}
			return nil
		} 
		for _, name := range ignore {
			if path == name {
				return nil
			}
		}
		content, _ := os.ReadFile(path)
		currentHash, _ := core.HashContent(content, "blob")
		hash, exists := index[path]
		if !exists {
			untracked = append(untracked, path)
		} else if hash != currentHash {
			modified = append(modified, path)
		}
		return nil
	})
	fmt.Printf("On branch %s\n", branch)
	if len(modified) == 0 && len(untracked) == 0 {
		fmt.Println("nothing to commit, working tree clean")
		return
	}
	if len(modified) > 0 {
		fmt.Println("\nChanges not staged for commit:")
		for _, f := range modified {
			fmt.Printf("  modified:   %s\n", f)
		}
	}
	if len(untracked) > 0 {
		fmt.Println("\nUntracked files:")
		for _, f := range untracked {
			fmt.Printf("  %s\n", f)
		}
	}

}
