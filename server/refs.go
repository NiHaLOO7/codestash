package server

import (
	"os"
	"strings"
	"path/filepath"
	"io/fs"
)



func InfoRefs(repoPath string, service string) string {
	refBytes, _ := os.ReadFile(repoPath + "/HEAD")
    ref := strings.TrimSpace(string(refBytes))
    branchPath := strings.TrimPrefix(ref, "ref: ")
    hashBytes, _ := os.ReadFile(repoPath + "/" + branchPath)
    headHash := strings.TrimSpace(string(hashBytes))
	headsPath := repoPath + "/refs/heads/"
	var response string

	response += PktLine("# service=" + service + "\n")
	response += PktFlush()

	response += PktLine(headHash + " HEAD\x00 report-status\n")



	filepath.WalkDir(headsPath, 
		func(path string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() {
            	return nil
        	}
            hash, _ := os.ReadFile(path)
        	name := strings.TrimPrefix(path, headsPath)
			response += PktLine(strings.TrimSpace(string(hash)) + " refs/heads/" + name + "\n")
        	return nil
		})

	response += PktFlush()
	return response
	

}