package server

import (
	"fmt"
)


func PktLine(data string) string {
	length := len(data) + 4
	//four character hash. max 64kb
	hash := fmt.Sprintf("%04x", length)
	return hash + data
}

func PktFlush() string {
	return "0000"
}

