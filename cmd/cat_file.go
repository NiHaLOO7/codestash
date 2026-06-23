package cmd
import (
	"fmt"
	"github.com/NiHaLOO7/codestash/storage"
	"github.com/NiHaLOO7/codestash/core"
)

func CatFile(hash string) {
	decompressed, err := storage.ReadObject(hash)
	if err != nil {
		fmt.Print("Error while parsing the data")
		return
	}
	data := core.ParseObject(decompressed)
	fmt.Print(string(data))
}
