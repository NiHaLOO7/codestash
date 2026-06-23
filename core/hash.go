package core
import (
	"fmt"
	"crypto/sha1"
)

func HashContent(content []byte, objType string) (string, []byte) {
	data := []byte(fmt.Sprintf("%s %d\x00%s", objType, len(content), content))
	hash := sha1.Sum(data)
	return fmt.Sprintf("%x", hash), data
}
