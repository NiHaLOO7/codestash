
package server

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"encoding/binary"
	"encoding/hex"
	"io"
	"os"
	"fmt"
)

func getTypeName(typeNum int) string {
	switch typeNum {
		case 1:
			return "commit"
		case 2:
			return "tree"
		case 3:
			return "blob"
		default:
			return ""
	}
}

func UnpackPackfile(repoPath string, data []byte) {
	count := binary.BigEndian.Uint32(data[8:12])
	offset := 12
	for i := 0; i < int(count); i++ {
		objType, _, bytesRead := decodeTypeSize(data, offset)
		offset += bytesRead
		underlying := bytes.NewReader(data[offset:])
		zlibReader, _ := zlib.NewReader(underlying)
		content, _ := io.ReadAll(zlibReader)
		zlibReader.Close()
		consumed := len(data[offset:]) - underlying.Len()
		offset += consumed

		typeName := getTypeName(objType)
		header := fmt.Sprintf("%s %d\x00", typeName, len(content))
		object := append([]byte(header), content...)
		hashBytes := sha1.Sum(object)
		hash := hex.EncodeToString(hashBytes[:])

		dir := repoPath + "/objects/" + hash[:2]
		os.MkdirAll(dir, 0755)
		filePath := dir + "/" + hash[2:]
		var buf bytes.Buffer
		w := zlib.NewWriter(&buf)
		w.Write(object)
		w.Close()
		os.WriteFile(filePath, buf.Bytes(), 0644)
	}
}

func decodeTypeSize(data []byte, offset int) (int, int, int) {
	firstByte := int(data[offset])
	objType := (firstByte >> 4) & 0x07
	size := firstByte & 0x0F
	shift := 4
	bytesRead := 1
	for firstByte & 0x80 != 0 {
		firstByte = int(data[offset + bytesRead])
		size |= (firstByte & 0x7F) << shift
		shift += 7
		bytesRead++
	}
	return objType, size, bytesRead
}