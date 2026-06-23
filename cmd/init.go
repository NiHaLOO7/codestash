package cmd
import (
	"os"
	"fmt"
	"path/filepath"
)

var PARENT = ".git"


func Init() {
	_, err := os.Stat(PARENT)
	if err == nil {
        fmt.Println("Already initialized")
        return
    }
	os.MkdirAll(PARENT+"/objects", 0755)
    os.MkdirAll(PARENT+"/refs/heads", 0755)
    os.WriteFile(PARENT+"/HEAD", []byte("ref: refs/heads/master\n"), 0644)
	os.WriteFile(PARENT+"/config", []byte("[user]\n\tname = CodeStash User\n\temail = user@codestash.dev\n"), 0644)
	abs, _ := filepath.Abs(PARENT)
    fmt.Printf("Initialized in %s/\n", abs)
}