package web

import (
	"math/rand"
	"net/http"
	"os"
	"strings"
)

func handleExplore(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"Title":  "Explore",
		"Active": "explore",
	}
	renderPage(w, r, "explore.html", data)
}

func handleNotifications(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"Title":  "Notifications",
		"Active": "notifications",
	}
	renderPage(w, r, "notifications.html", data)
}

func handleNewIssue(w http.ResponseWriter, r *http.Request, repo string) {
	data := map[string]interface{}{
		"Title":  "New Issue - " + repo,
		"Repo":   repo,
		"Branch": getDefaultBranch(reposDir + "/" + repo),
		"Active": "issues",
	}
	renderPage(w, r, "issue-new.html", data)
}

func handleIssueDetail(w http.ResponseWriter, r *http.Request, repo, id string) {
	data := map[string]interface{}{
		"Title":   "Issue #" + id + " - " + repo,
		"Repo":    repo,
		"Branch":  getDefaultBranch(reposDir + "/" + repo),
		"IssueID": id,
		"Active":  "issues",
	}
	renderPage(w, r, "issue-detail.html", data)
}

func handleNewPR(w http.ResponseWriter, r *http.Request, repo string) {
	data := map[string]interface{}{
		"Title":    "New Pull Request - " + repo,
		"Repo":     repo,
		"Branch":   getDefaultBranch(reposDir + "/" + repo),
		"Branches": listBranches(reposDir + "/" + repo),
		"Active":   "pulls",
	}
	renderPage(w, r, "pr-new.html", data)
}

func handlePRDetail(w http.ResponseWriter, r *http.Request, repo, id string) {
	data := map[string]interface{}{
		"Title":  "Pull Request #" + id + " - " + repo,
		"Repo":   repo,
		"Branch": getDefaultBranch(reposDir + "/" + repo),
		"PRID":   id,
		"Active": "pulls",
	}
	renderPage(w, r, "pr-detail.html", data)
}

func handleReleases(w http.ResponseWriter, r *http.Request, repo string) {
	data := map[string]interface{}{
		"Title":  "Releases - " + repo,
		"Repo":   repo,
		"Branch": getDefaultBranch(reposDir + "/" + repo),
		"Active": "releases",
	}
	renderPage(w, r, "releases.html", data)
}

func handleNewRelease(w http.ResponseWriter, r *http.Request, repo string) {
	data := map[string]interface{}{
		"Title":    "New Release - " + repo,
		"Repo":     repo,
		"Branch":   getDefaultBranch(reposDir + "/" + repo),
		"Branches": listBranches(reposDir + "/" + repo),
		"Active":   "releases",
	}
	renderPage(w, r, "release-new.html", data)
}

func handleTags(w http.ResponseWriter, r *http.Request, repo string) {
	data := map[string]interface{}{
		"Title":  "Tags - " + repo,
		"Repo":   repo,
		"Branch": getDefaultBranch(reposDir + "/" + repo),
		"Active": "tags",
	}
	renderPage(w, r, "tags.html", data)
}

func handleLabels(w http.ResponseWriter, r *http.Request, repo string) {
	data := map[string]interface{}{
		"Title":  "Labels - " + repo,
		"Repo":   repo,
		"Branch": getDefaultBranch(reposDir + "/" + repo),
		"Active": "issues",
	}
	renderPage(w, r, "labels.html", data)
}

func handleMilestones(w http.ResponseWriter, r *http.Request, repo string) {
	data := map[string]interface{}{
		"Title":  "Milestones - " + repo,
		"Repo":   repo,
		"Branch": getDefaultBranch(reposDir + "/" + repo),
		"Active": "issues",
	}
	renderPage(w, r, "milestones.html", data)
}

func handleSearch(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	data := map[string]interface{}{
		"Title":  "Search",
		"Query":  query,
		"Active": "search",
	}
	renderPage(w, r, "search.html", data)
}

func handleProfile(w http.ResponseWriter, r *http.Request) {
	entries, _ := os.ReadDir(reposDir)
	repoCount := 0
	for _, e := range entries {
		if e.IsDir() && strings.HasSuffix(e.Name(), ".git") {
			repoCount++
		}
	}

	weeks := make([][]int, 52)
	for i := range weeks {
		week := make([]int, 7)
		for j := range week {
			week[j] = rand.Intn(5)
		}
		weeks[i] = week
	}

	data := map[string]interface{}{
		"Title":     "Profile",
		"Active":    "profile",
		"RepoCount": repoCount,
		"Weeks":     weeks,
	}
	renderPage(w, r, "profile.html", data)
}

func handleGists(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"Title":  "Gists",
		"Active": "gists",
	}
	renderPage(w, r, "gists.html", data)
}

func handleNewGist(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"Title":  "New Gist",
		"Active": "gists",
	}
	renderPage(w, r, "gist-new.html", data)
}

func handleProjects(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"Title":  "Projects",
		"Active": "projects",
	}
	renderPage(w, r, "projects.html", data)
}

func handleBlame(w http.ResponseWriter, r *http.Request, repo, branch, filePath string) {
	repoPath := reposDir + "/" + repo
	if branch == "" {
		branch = getDefaultBranch(repoPath)
	}

	commitHash := resolveRef(repoPath, branch)
	commit := parseCommitObject(repoPath, commitHash)

	blobHash := resolveBlobPath(repoPath, commit.Tree, filePath)
	if blobHash == "" {
		http.NotFound(w, r)
		return
	}

	content := readObjectContent(repoPath, blobHash)
	lines := strings.Split(string(content), "\n")

	parts := strings.Split(filePath, "/")
	fileName := parts[len(parts)-1]

	var pathParts []PathPart
	if len(parts) > 1 {
		dirParts := parts[:len(parts)-1]
		for i, p := range dirParts {
			pathParts = append(pathParts, PathPart{
				Name: p,
				Path: strings.Join(dirParts[:i+1], "/"),
			})
		}
	}

	type BlameLine struct {
		Num        int
		Content    string
		CommitHash string
		Author     string
		Date       string
		NewBlock   bool
	}

	var blameLines []BlameLine
	for i, line := range lines {
		blameLines = append(blameLines, BlameLine{
			Num:        i + 1,
			Content:    line,
			CommitHash: commitHash,
			Author:     commit.Author,
			Date:       "recently",
			NewBlock:   i == 0,
		})
	}

	data := map[string]interface{}{
		"Title":     "Blame - " + fileName + " - " + repo,
		"Repo":      repo,
		"Branch":    branch,
		"FileName":  fileName,
		"FilePath":  filePath,
		"Lines":     blameLines,
		"LineCount": len(lines),
		"PathParts": pathParts,
	}
	renderPage(w, r, "blame.html", data)
}

func handleFileHistory(w http.ResponseWriter, r *http.Request, repo, branch, filePath string) {
	repoPath := reposDir + "/" + repo
	if branch == "" {
		branch = getDefaultBranch(repoPath)
	}

	commitHash := resolveRef(repoPath, branch)
	commits := getCommitLog(repoPath, commitHash, 50)

	parts := strings.Split(filePath, "/")
	fileName := parts[len(parts)-1]

	var pathParts []PathPart
	if len(parts) > 1 {
		dirParts := parts[:len(parts)-1]
		for i, p := range dirParts {
			pathParts = append(pathParts, PathPart{
				Name: p,
				Path: strings.Join(dirParts[:i+1], "/"),
			})
		}
	}

	data := map[string]interface{}{
		"Title":     "History - " + fileName + " - " + repo,
		"Repo":      repo,
		"Branch":    branch,
		"FileName":  fileName,
		"FilePath":  filePath,
		"Commits":   commits,
		"PathParts": pathParts,
	}
	renderPage(w, r, "file-history.html", data)
}
