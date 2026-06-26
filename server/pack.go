package server

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"encoding/binary"
	"encoding/hex"
	"io"
	"os"
	"strings"
)

func getTypeCode(typeName string) int {
	switch typeName {
		case "commit":
			return 1
		case "tree":
			return 2
		case "blob":
			return 3
		default:
			return 0
	}
}

func CreatePackFile(repoPath string, objects []string) []byte {
	var buf bytes.Buffer
	buf.WriteString("PACK")
	binary.Write(&buf, binary.BigEndian, uint32(2))
	binary.Write(&buf, binary.BigEndian, uint32(len(objects)))

	for _, hash := range objects {
		path := repoPath + "/objects/" + hash[:2] + "/" + hash[2:]
		data, _ := os.ReadFile(path)
		reader, _ := zlib.NewReader(bytes.NewReader(data))
		decompressed, _ := io.ReadAll(reader)
		nullPos := bytes.IndexByte(decompressed, 0)
		header := string(decompressed[:nullPos])     // "blob 12"
		content := decompressed[nullPos+1:]          // actual data
		parts := strings.Split(header, " ")
		typeName := parts[0]
		typeNum := getTypeCode(typeName)
		byteData := EncodeTypeSize(typeNum, len(content))
		var compressedBuf bytes.Buffer
		zlibWriter := zlib.NewWriter(&compressedBuf)
		zlibWriter.Write(content)
		zlibWriter.Close()
		buf.Write(byteData)
		buf.Write(compressedBuf.Bytes())
	}
	checksum := sha1.Sum(buf.Bytes())
	buf.Write(checksum[:])

	return buf.Bytes()
}

func EncodeTypeSize(objType int, size int) []byte {
	firstByte := objType << 4 | (size & 0x0F)
	size = size >> 4
	var result []byte
	if size > 0 {
		firstByte |= 0x80
	}
	result = append(result, byte(firstByte))
	for size > 0 {
		nextByte := size & 0x7F
		size = size >> 7 
		if size > 0 {
			nextByte |= 0x80
		}
		result = append(result, byte(nextByte))
	}
	return result
}

func CollectObjects(repoPath string, hash string) []string {
		if len(hash) != 40 {
			return nil
		}
		path := repoPath + "/objects/" + hash[:2] + "/" + hash[2:]
		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		reader, err := zlib.NewReader(bytes.NewReader(data))
		if err != nil {
			return nil
		}
		decompressed, _ := io.ReadAll(reader)
		nullPos := bytes.IndexByte(decompressed, 0)
		header := string(decompressed[:nullPos])     // "blob 12"
		content := decompressed[nullPos+1:]          // actual data
		parts := strings.Split(header, " ")
		typeName := parts[0]
		var result []string
    	result = append(result, hash)   // khud ko add kar
		switch typeName {
		case "commit":
			lines := strings.Split(string(content), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if after, ok :=strings.CutPrefix(line, "tree "); ok  {
					treeHash := after
                	result = append(result, CollectObjects(repoPath, treeHash)...)
				}
				if after, ok :=strings.CutPrefix(line, "parent "); ok  {
					parentHash := after
                	result = append(result, CollectObjects(repoPath, parentHash)...)
				}
			}
		case "tree":
			i := 0
			for i < len(content) {
				spacePos := bytes.IndexByte(content[i:], ' ')
				i += spacePos+1
				nullPos := bytes.IndexByte(content[i:],0)
				i += nullPos + 1
				rawHash := content[i : i+20]
				hexHash := hex.EncodeToString(rawHash)
				i += 20
				result = append(result, CollectObjects(repoPath, hexHash)...)
			}
		}
		return result
}
