package cmd

import (
	"fmt"
	"os"

	"github.com/NiHaLOO7/codestash/core"
	"github.com/NiHaLOO7/codestash/storage"
)

func HashObject(filename string) {
	content, err := os.ReadFile(filename)
	if err != nil {
		fmt.Printf("%s not found", filename)
		return
	}
	hash, data := core.HashContent(content, "blob")
	storage.WriteObject(hash, data)
	fmt.Println(hash)
}
