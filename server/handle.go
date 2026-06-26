package server

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

func StartServer() {
	http.HandleFunc("/", handleRequests)
	fmt.Println("Server started on :8080")
	http.ListenAndServe(":8080", nil)
}

func handleRequests(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	parts := strings.Split(path, "/")
	if len(parts) < 2 {
		http.NotFound(w, r)
		return
	}
	repo := parts[1]

	if strings.HasSuffix(path, "info/refs") {
		repoPath := "repos/" + repo
		service := r.URL.Query().Get("service")
		if service == "" {
			service = "git-upload-pack"
		}
		w.Header().Set("Content-Type", "application/x-"+service+"-advertisement")
		fmt.Fprint(w, InfoRefs(repoPath, service))

	} else if strings.HasSuffix(path, "git-upload-pack") {
		bodyByte, _ := io.ReadAll(r.Body)
		body := strings.Split(string(bodyByte), "\n")
		var allObjects []string
		for _, line := range body {
			line = strings.TrimSpace(line)
			if strings.Contains(line, "want ") {
				wantIndex := strings.Index(line, "want ")
				if wantIndex+45 > len(line) {
					continue
				}
				wantedHash := line[wantIndex+5 : wantIndex+45]
				repoPath := "repos/" + repo
				allObjects = append(allObjects, CollectObjects(repoPath, wantedHash)...)
			}
		}
		repoPath := "repos/" + repo
		packfile := CreatePackFile(repoPath, allObjects)
		w.Header().Set("Content-Type", "application/x-git-upload-pack-result")
		fmt.Fprint(w, "0008NAK\n")
		w.Write(packfile)

	} else if strings.HasSuffix(path, "git-receive-pack") {
		bodyByte, _ := io.ReadAll(r.Body)
		flushPos := bytes.Index(bodyByte, []byte("0000"))
		commandLine := string(bodyByte[:flushPos])
		newHash := commandLine[45:85]
		refPart := commandLine[86:]
		nullPos := strings.Index(refPart, "\x00")
		if nullPos != -1 {
			refPart = refPart[:nullPos]
		}
		refName := strings.TrimSpace(refPart)
		packData := bodyByte[flushPos+4:]
		repoPath := "repos/" + repo
		UnpackPackfile(repoPath, packData)
		os.WriteFile(repoPath+"/"+refName, []byte(newHash), 0644)
		w.Header().Set("Content-Type", "application/x-git-receive-pack-result")
		fmt.Fprint(w, PktLine("unpack ok\n"))
		fmt.Fprint(w, PktLine("ok "+refName+"\n"))
		fmt.Fprint(w, PktFlush())

	} else {
		http.NotFound(w, r)
		return
	}
}
