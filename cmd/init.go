package cmd
import (
	"os"
	"fmt"
	"path/filepath"
)

const PARENT = ".codestash"


func Init() {
	_, err := os.Stat(PARENT)
	if err == nil {
        fmt.Println("Already initialized")
        return
    }
	os.MkdirAll(PARENT+"/objects", 0755)
    os.MkdirAll(PARENT+"/refs/heads", 0755)
    os.WriteFile(PARENT+"/HEAD", []byte("ref: refs/heads/master\n"), 0644)
	abs, _ := filepath.Abs(PARENT)
    fmt.Printf("Initialized in %s/\n", abs)
}