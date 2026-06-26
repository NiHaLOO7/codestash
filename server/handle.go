package server

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

func StartServer() {
	http.HandleFunc("/", handleRequests)
	fmt.Println("Server started on :8080") // print karo
	http.ListenAndServe(":8080", nil)

}

func handleRequests(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	parts := strings.Split(path, "/")
	repo := parts[1]
	if strings.HasSuffix(path, "info/refs") {
		repoPath := "repos/" + repo
		w.Header().Set("Content-Type", "application/x-git-upload-pack-advertisement")
		fmt.Fprint(w, InfoRefs(repoPath))
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
		fmt.Fprintf(w, "receive-pack called for %s", repo)
	} else {
		http.NotFound(w, r)
		return
	}

}
