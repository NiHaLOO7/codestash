package storage

import (
	"bytes"
	"compress/zlib"
	"io"
	"fmt"
	"os"
	"github.com/NiHaLOO7/codestash/internal"
)

var PARENT = internal.PARENT
func WriteObject(hash string, data []byte) {
	folder := hash[:2]
	file := hash[2:]
	_, err := os.Stat(PARENT)
	if err != nil {
		fmt.Println("Git not initialized")
        return
	}
	os.MkdirAll(PARENT + "/objects/" + folder, 0755)
	var buf bytes.Buffer
	writer := zlib.NewWriter(&buf)
	writer.Write(data)
	writer.Close()
	compressed := buf.Bytes()
	os.WriteFile(PARENT + "/objects/" + folder + "/" + file, compressed, 0644)
}

func ReadObject(hash string) ([]byte, error) {
	filepath := PARENT + "/objects/" + hash[:2] + "/" + hash[2:]
	compressed, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}
	reader, _ := zlib.NewReader(bytes.NewReader(compressed))
	decompressed, _ := io.ReadAll(reader)
	reader.Close()
	return decompressed, nil
}
