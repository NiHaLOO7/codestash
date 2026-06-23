package core

import (
	"fmt"
	"os"
	"encoding/binary"
	"encoding/hex"
	"bytes"
	"crypto/sha1"
)
func indexPath() string { return PARENT + "/index" }

func ReadIndex() map[string]string {
	_, err := os.Stat(PARENT)
	if err != nil {
		fmt.Println("Git not initialized")
		return nil
	}
	data, err := os.ReadFile(indexPath())
	hashmap := make(map[string]string)
	if err != nil {
		return hashmap
	}
	if len(data) < 12 || string(data[:4]) != "DIRC" {
		return hashmap
	}
	entryCount := binary.BigEndian.Uint32(data[8:12])
	i := 12
	for e := uint32(0); e < entryCount; e++ {
		i += 40
		hexHash := hex.EncodeToString(data[i : i+20])
		i += 20
		flags := binary.BigEndian.Uint16(data[i : i+2])
		nameLen := int(flags & 0xFFF)
		i += 2
		filename := string(data[i : i+nameLen])
		i += nameLen
		entrySize := 62 + nameLen
		padding := 8 - (entrySize % 8)
		if padding == 0 {
			padding = 8
		}
		i += padding
		hashmap[filename] = hexHash
	}
	return hashmap
}

func WriteIndex(index map[string]string) {
	_, err := os.Stat(PARENT)
	if err != nil {
		fmt.Println("Git not initialized")
        return
	}
	var buf bytes.Buffer
	buf.WriteString("DIRC")
	binary.Write(&buf, binary.BigEndian, uint32(2))
	binary.Write(&buf, binary.BigEndian, uint32(len(index)))
	for filename, hash := range index {
		info, _ := os.Stat(filename)
		mtime := info.ModTime()
		binary.Write(&buf, binary.BigEndian, uint32(mtime.Unix()))
		binary.Write(&buf, binary.BigEndian, uint32(mtime.Nanosecond()))
		binary.Write(&buf, binary.BigEndian, uint32(mtime.Unix()))
		binary.Write(&buf, binary.BigEndian, uint32(mtime.Nanosecond()))
		binary.Write(&buf, binary.BigEndian, uint32(0))
		binary.Write(&buf, binary.BigEndian, uint32(0))
		binary.Write(&buf, binary.BigEndian, uint32(0100644))
		binary.Write(&buf, binary.BigEndian, uint32(0))
		binary.Write(&buf, binary.BigEndian, uint32(0))
		binary.Write(&buf, binary.BigEndian, uint32(info.Size()))
		rawHash, _ := hex.DecodeString(hash)
		buf.Write(rawHash)
		binary.Write(&buf, binary.BigEndian, uint16(len(filename)&0xFFF))
		buf.WriteString(filename)
		entrySize := 62 + len(filename)
		padding := 8 - (entrySize % 8)
		if padding == 0 {
			padding = 8
		}
		buf.Write(make([]byte, padding))
	}
	checksum := sha1.Sum(buf.Bytes())
	buf.Write(checksum[:])
	os.WriteFile(indexPath(), buf.Bytes(), 0644)
}

func AddToIndex(filename string, hash string) {
	index := ReadIndex()
	index[filename] = hash
	WriteIndex(index)
}


