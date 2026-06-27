package web

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/NiHaLOO7/codestash/auth"
)

type RepoInfo struct {
	Name   string
	Branch string
}

type FileEntry struct {
	Name  string
	Path  string
	IsDir bool
}

type CommitInfo struct {
	Hash    string
	Message string
	Author  string
	Tree    string
	Parent  string
}

type PathPart struct {
	Name string
	Path string
}

type FileLine struct {
	Num     int
	Content string
}

type DiffLine struct {
	OldNum  string
	NewNum  string
	Content template.HTML
	Type    string
}

type DiffFile struct {
	FileName string
	Lines    []DiffLine
}

var templates *template.Template
var reposDir string

func StartWeb(repos string) {
	reposDir = repos

	auth.CleanExpiredSessions()

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))
	http.HandleFunc("/", handleWeb)

	fmt.Println("Web UI started on :9090")
	fmt.Println("Security: X-Frame-Options, X-Content-Type-Options, HttpOnly cookies enabled")
	http.ListenAndServe(":9090", nil)
}

func handleWeb(w http.ResponseWriter, r *http.Request) {
	path := strings.Trim(r.URL.Path, "/")

	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-XSS-Protection", "1; mode=block")
	w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

	// Auth protection: redirect to login if not authenticated
	user := auth.GetCurrentUser(r)
	if user == "" && path != "login" && path != "signup" && !strings.HasPrefix(path, "static") {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if path == "" {
		handleRepos(w, r)
		return
	}

	if path == "new" {
		handleNewRepo(w, r)
		return
	}

	if path == "signup" {
		handleSignup(w, r)
		return
	}

	if path == "login" {
		handleLogin(w, r)
		return
	}

	if path == "logout" {
		handleLogout(w, r)
		return
	}

	if path == "settings/profile" {
		handleUserSettings(w, r)
		return
	}

	if strings.HasPrefix(path, "user/") || path == "profile" {
		handleProfile(w, r)
		return
	}

	if path == "explore" {
		handleExplore(w, r)
		return
	}

	if path == "notifications" {
		handleNotifications(w, r)
		return
	}

	if path == "search" {
		handleSearch(w, r)
		return
	}

	if path == "gists" {
		handleGists(w, r)
		return
	}

	if path == "gists/new" {
		handleNewGist(w, r)
		return
	}

	if path == "projects" {
		handleProjects(w, r)
		return
	}

	parts := strings.SplitN(path, "/", 4)
	repo := parts[0]

	if len(parts) == 1 {
		handleTree(w, r, repo, "", "")
		return
	}

	action := parts[1]
	switch action {
	case "tree":
		branch := ""
		subpath := ""
		if len(parts) >= 3 {
			branch = parts[2]
		}
		if len(parts) >= 4 {
			subpath = parts[3]
		}
		handleTree(w, r, repo, branch, subpath)
	case "blob":
		branch := ""
		filePath := ""
		if len(parts) >= 3 {
			branch = parts[2]
		}
		if len(parts) >= 4 {
			filePath = parts[3]
		}
		handleBlob(w, r, repo, branch, filePath)
	case "commits":
		branch := "main"
		if len(parts) >= 3 {
			branch = parts[2]
		}
		handleCommits(w, r, repo, branch)
	case "commit":
		hash := ""
		if len(parts) >= 3 {
			hash = parts[2]
		}
		handleDiff(w, r, repo, hash)
	case "new-branch":
		sourceBranch := "main"
		if len(parts) >= 3 {
			sourceBranch = parts[2]
		}
		handleNewBranch(w, r, repo, sourceBranch)
	case "edit":
		branch := ""
		filePath := ""
		if len(parts) >= 3 {
			branch = parts[2]
		}
		if len(parts) >= 4 {
			filePath = parts[3]
		}
		handleEdit(w, r, repo, branch, filePath)
	case "issues":
		if len(parts) >= 3 {
			if parts[2] == "new" {
				handleNewIssue(w, r, repo)
			} else {
				handleIssueDetail(w, r, repo, parts[2])
			}
		} else {
			handleIssues(w, r, repo)
		}
	case "pulls":
		if len(parts) >= 3 {
			if parts[2] == "new" {
				handleNewPR(w, r, repo)
			} else {
				handlePRDetail(w, r, repo, parts[2])
			}
		} else {
			handlePulls(w, r, repo)
		}
	case "pipelines":
		handlePipelines(w, r, repo)
	case "wiki":
		handleWiki(w, r, repo)
	case "settings":
		handleRepoSettings(w, r, repo)
	case "releases":
		if len(parts) >= 3 && parts[2] == "new" {
			handleNewRelease(w, r, repo)
		} else {
			handleReleases(w, r, repo)
		}
	case "tags":
		handleTags(w, r, repo)
	case "labels":
		handleLabels(w, r, repo)
	case "milestones":
		handleMilestones(w, r, repo)
	case "blame":
		branch := ""
		filePath := ""
		if len(parts) >= 3 {
			branch = parts[2]
		}
		if len(parts) >= 4 {
			filePath = parts[3]
		}
		handleBlame(w, r, repo, branch, filePath)
	case "history":
		branch := ""
		filePath := ""
		if len(parts) >= 3 {
			branch = parts[2]
		}
		if len(parts) >= 4 {
			filePath = parts[3]
		}
		handleFileHistory(w, r, repo, branch, filePath)
	default:
		http.NotFound(w, r)
	}
}

func handleIssues(w http.ResponseWriter, r *http.Request, repo string) {
	data := map[string]interface{}{
		"Title":  "Issues - " + repo,
		"Repo":   repo,
		"Branch": getDefaultBranch(reposDir + "/" + repo),
		"Issues": []map[string]string{},
		"Active": "issues",
	}
	renderPage(w, r, "issues.html", data)
}

func handlePulls(w http.ResponseWriter, r *http.Request, repo string) {
	data := map[string]interface{}{
		"Title":  "Pull Requests - " + repo,
		"Repo":   repo,
		"Branch": getDefaultBranch(reposDir + "/" + repo),
		"Pulls":  []map[string]string{},
		"Active": "pulls",
	}
	renderPage(w, r, "pulls.html", data)
}

func handlePipelines(w http.ResponseWriter, r *http.Request, repo string) {
	data := map[string]interface{}{
		"Title":     "CI/CD - " + repo,
		"Repo":      repo,
		"Branch":    getDefaultBranch(reposDir + "/" + repo),
		"Pipelines": []map[string]string{},
		"Active":    "pipelines",
	}
	renderPage(w, r, "pipelines.html", data)
}

func handleWiki(w http.ResponseWriter, r *http.Request, repo string) {
	data := map[string]interface{}{
		"Title":  "Wiki - " + repo,
		"Repo":   repo,
		"Branch": getDefaultBranch(reposDir + "/" + repo),
		"Pages":  []map[string]string{},
		"Active": "wiki",
	}
	renderPage(w, r, "wiki.html", data)
}

func handleRepoSettings(w http.ResponseWriter, r *http.Request, repo string) {
	data := map[string]interface{}{
		"Title":  "Settings - " + repo,
		"Repo":   repo,
		"Branch": getDefaultBranch(reposDir + "/" + repo),
		"Active": "settings",
	}
	renderPage(w, r, "settings.html", data)
}

func handleUserSettings(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"Title":  "Profile Settings",
		"Active": "settings",
	}
	renderPage(w, r, "user-settings.html", data)
}

func handleRepos(w http.ResponseWriter, r *http.Request) {
	entries, _ := os.ReadDir(reposDir)
	var repos []RepoInfo
	for _, e := range entries {
		if e.IsDir() && strings.HasSuffix(e.Name(), ".git") {
			branch := getDefaultBranch(reposDir + "/" + e.Name())
			repos = append(repos, RepoInfo{Name: e.Name(), Branch: branch})
		}
	}

	data := map[string]interface{}{
		"Title":     "Repositories",
		"Repos":     repos,
		"RepoCount": len(repos),
		"Active":    "repos",
	}
	renderPage(w, r, "repos.html", data)
}

func handleNewRepo(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		name := r.FormValue("name")
		if name == "" {
			http.Error(w, "Repository name is required", 400)
			return
		}
		if !strings.HasSuffix(name, ".git") {
			name = name + ".git"
		}

		repoPath := reposDir + "/" + name
		os.MkdirAll(repoPath+"/objects", 0755)
		os.MkdirAll(repoPath+"/refs/heads", 0755)
		os.WriteFile(repoPath+"/HEAD", []byte("ref: refs/heads/main\n"), 0644)

		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	data := map[string]interface{}{
		"Title":  "New Repository",
		"Active": "new",
	}
	renderPage(w, r, "new-repo.html", data)
}

func handleNewBranch(w http.ResponseWriter, r *http.Request, repo, sourceBranch string) {
	repoPath := reposDir + "/" + repo

	if r.Method == "POST" {
		branchName := r.FormValue("branch")
		if branchName == "" {
			http.Error(w, "Branch name is required", 400)
			return
		}

		sourceHash := resolveRef(repoPath, sourceBranch)
		if sourceHash == "" {
			http.Error(w, "Source branch not found", 400)
			return
		}

		branchDir := filepath.Dir(repoPath + "/refs/heads/" + branchName)
		os.MkdirAll(branchDir, 0755)
		os.WriteFile(repoPath+"/refs/heads/"+branchName, []byte(sourceHash), 0644)

		http.Redirect(w, r, "/"+repo+"/tree/"+branchName, http.StatusSeeOther)
		return
	}

	data := map[string]interface{}{
		"Title":        "Create Branch - " + repo,
		"Repo":         repo,
		"SourceBranch": sourceBranch,
	}
	renderPage(w, r, "new-branch.html", data)
}

func handleEdit(w http.ResponseWriter, r *http.Request, repo, branch, filePath string) {
	repoPath := reposDir + "/" + repo
	if branch == "" {
		branch = getDefaultBranch(repoPath)
	}

	commitHash := resolveRef(repoPath, branch)
	commit := parseCommitObject(repoPath, commitHash)

	if r.Method == "POST" {
		content := r.FormValue("content")
		message := r.FormValue("message")
		author := r.FormValue("author")
		email := r.FormValue("email")

		if message == "" {
			message = "Update " + filePath
		}
		if author == "" {
			author = "CodeStash User"
		}
		if email == "" {
			email = "user@codestash.local"
		}

		blobHash := writeObject(repoPath, "blob", []byte(content))
		newTreeHash := updateTree(repoPath, commit.Tree, filePath, blobHash)
		timestamp := time.Now().Unix()
		commitContent := fmt.Sprintf("tree %s\nparent %s\nauthor %s <%s> %d +0000\ncommitter %s <%s> %d +0000\n\n%s\n",
			newTreeHash, commitHash, author, email, timestamp, author, email, timestamp, message)
		newCommitHash := writeObject(repoPath, "commit", []byte(commitContent))
		os.WriteFile(repoPath+"/refs/heads/"+branch, []byte(newCommitHash), 0644)

		http.Redirect(w, r, "/"+repo+"/blob/"+branch+"/"+filePath, http.StatusSeeOther)
		return
	}

	blobHash := resolveBlobPath(repoPath, commit.Tree, filePath)
	content := ""
	if blobHash != "" {
		content = string(readObjectContent(repoPath, blobHash))
	}

	data := map[string]interface{}{
		"Title":    "Edit " + filePath + " - " + repo,
		"Repo":     repo,
		"Branch":   branch,
		"FilePath": filePath,
		"Content":  content,
	}
	renderPage(w, r, "edit.html", data)
}

func handleTree(w http.ResponseWriter, r *http.Request, repo, branch, subpath string) {
	repoPath := reposDir + "/" + repo
	if branch == "" {
		branch = getDefaultBranch(repoPath)
	}

	commitHash := resolveRef(repoPath, branch)
	if commitHash == "" {
		data := map[string]interface{}{
			"Title":        repo,
			"Repo":         repo,
			"Branch":       branch,
			"Files":        []FileEntry{},
			"Commits":      []CommitInfo{},
			"Branches":     listBranches(repoPath),
			"PathParts":    []PathPart{},
			"LatestCommit": nil,
		}
		renderPage(w, r, "tree.html", data)
		return
	}

	commit := parseCommitObject(repoPath, commitHash)
	treeHash := commit.Tree

	if subpath != "" {
		treeHash = resolveTreePath(repoPath, treeHash, subpath)
		if treeHash == "" {
			http.NotFound(w, r)
			return
		}
	}

	files := listTree(repoPath, treeHash, subpath)
	sort.Slice(files, func(i, j int) bool {
		if files[i].IsDir != files[j].IsDir {
			return files[i].IsDir
		}
		return files[i].Name < files[j].Name
	})

	commits := getCommitLog(repoPath, commitHash, 5)
	branches := listBranches(repoPath)

	var pathParts []PathPart
	if subpath != "" {
		parts := strings.Split(subpath, "/")
		for i, p := range parts {
			pathParts = append(pathParts, PathPart{
				Name: p,
				Path: strings.Join(parts[:i+1], "/"),
			})
		}
	}

	var latestCommit *CommitInfo
	if len(commits) > 0 {
		latestCommit = &commits[0]
	}

	data := map[string]interface{}{
		"Title":        repo,
		"Repo":         repo,
		"Branch":       branch,
		"Files":        files,
		"Commits":      commits,
		"Branches":     branches,
		"PathParts":    pathParts,
		"LatestCommit": latestCommit,
	}
	renderPage(w, r, "tree.html", data)
}

func handleBlob(w http.ResponseWriter, r *http.Request, repo, branch, filePath string) {
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

	var fileLines []FileLine
	for i, line := range lines {
		fileLines = append(fileLines, FileLine{Num: i + 1, Content: line})
	}

	parts := strings.Split(filePath, "/")
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
		"Title":     filePath + " - " + repo,
		"Repo":      repo,
		"Branch":    branch,
		"FileName":  parts[len(parts)-1],
		"FilePath":  filePath,
		"Lines":     fileLines,
		"LineCount": len(lines),
		"PathParts": pathParts,
	}
	renderPage(w, r, "blob.html", data)
}

func handleCommits(w http.ResponseWriter, r *http.Request, repo, branch string) {
	repoPath := reposDir + "/" + repo
	commitHash := resolveRef(repoPath, branch)
	commits := getCommitLog(repoPath, commitHash, 50)

	data := map[string]interface{}{
		"Title":   "Commits - " + repo,
		"Repo":    repo,
		"Branch":  branch,
		"Commits": commits,
	}
	renderPage(w, r, "commits.html", data)
}

func handleDiff(w http.ResponseWriter, r *http.Request, repo, hash string) {
	repoPath := reposDir + "/" + repo
	commit := parseCommitObject(repoPath, hash)
	branch := getDefaultBranch(repoPath)

	var diffs []DiffFile
	newTree := parseTreeObject(repoPath, commit.Tree)

	var oldTree map[string]string
	if commit.Parent != "" {
		parentCommit := parseCommitObject(repoPath, commit.Parent)
		oldTree = parseTreeObject(repoPath, parentCommit.Tree)
	} else {
		oldTree = make(map[string]string)
	}

	for name, newHash := range newTree {
		oldHash, exists := oldTree[name]
		if !exists {
			content := readObjectContent(repoPath, newHash)
			lines := strings.Split(string(content), "\n")
			var diffLines []DiffLine
			for i, line := range lines {
				diffLines = append(diffLines, DiffLine{
					OldNum:  "",
					NewNum:  fmt.Sprintf("%d", i+1),
					Content: template.HTML("+ " + template.HTMLEscapeString(line)),
					Type:    "addition",
				})
			}
			diffs = append(diffs, DiffFile{FileName: name, Lines: diffLines})
		} else if oldHash != newHash {
			oldContent := string(readObjectContent(repoPath, oldHash))
			newContent := string(readObjectContent(repoPath, newHash))
			diffLines := computeWordDiff(oldContent, newContent)
			diffs = append(diffs, DiffFile{FileName: name, Lines: diffLines})
		}
	}

	for name := range oldTree {
		if _, exists := newTree[name]; !exists {
			content := readObjectContent(repoPath, oldTree[name])
			lines := strings.Split(string(content), "\n")
			var diffLines []DiffLine
			for i, line := range lines {
				diffLines = append(diffLines, DiffLine{
					OldNum:  fmt.Sprintf("%d", i+1),
					NewNum:  "",
					Content: template.HTML("- " + template.HTMLEscapeString(line)),
					Type:    "deletion",
				})
			}
			diffs = append(diffs, DiffFile{FileName: name + " (deleted)", Lines: diffLines})
		}
	}

	data := map[string]interface{}{
		"Title":   "Commit " + hash[:7] + " - " + repo,
		"Repo":    repo,
		"Branch":  branch,
		"Hash":    hash,
		"Message": commit.Message,
		"Author":  commit.Author,
		"Diffs":   diffs,
	}
	renderPage(w, r, "diff.html", data)
}

func computeWordDiff(oldContent, newContent string) []DiffLine {
	oldLines := strings.Split(oldContent, "\n")
	newLines := strings.Split(newContent, "\n")
	var result []DiffLine

	lcs := lcsLines(oldLines, newLines)

	oldIdx, newIdx, lcsIdx := 0, 0, 0

	for lcsIdx < len(lcs) {
		for oldIdx < len(oldLines) && oldLines[oldIdx] != lcs[lcsIdx] {
			result = append(result, DiffLine{
				OldNum:  fmt.Sprintf("%d", oldIdx+1),
				NewNum:  "",
				Content: renderWordDelete(oldLines[oldIdx]),
				Type:    "deletion",
			})
			oldIdx++
		}
		for newIdx < len(newLines) && newLines[newIdx] != lcs[lcsIdx] {
			result = append(result, DiffLine{
				OldNum:  "",
				NewNum:  fmt.Sprintf("%d", newIdx+1),
				Content: renderWordAdd(newLines[newIdx]),
				Type:    "addition",
			})
			newIdx++
		}
		result = append(result, DiffLine{
			OldNum:  fmt.Sprintf("%d", oldIdx+1),
			NewNum:  fmt.Sprintf("%d", newIdx+1),
			Content: template.HTML("  " + template.HTMLEscapeString(lcs[lcsIdx])),
			Type:    "context",
		})
		oldIdx++
		newIdx++
		lcsIdx++
	}

	for oldIdx < len(oldLines) {
		result = append(result, DiffLine{
			OldNum:  fmt.Sprintf("%d", oldIdx+1),
			NewNum:  "",
			Content: renderWordDelete(oldLines[oldIdx]),
			Type:    "deletion",
		})
		oldIdx++
	}
	for newIdx < len(newLines) {
		result = append(result, DiffLine{
			OldNum:  "",
			NewNum:  fmt.Sprintf("%d", newIdx+1),
			Content: renderWordAdd(newLines[newIdx]),
			Type:    "addition",
		})
		newIdx++
	}

	return result
}

func renderWordDelete(line string) template.HTML {
	words := strings.Fields(line)
	var parts []string
	for _, w := range words {
		parts = append(parts, `<span class="diff-word-del">`+template.HTMLEscapeString(w)+`</span>`)
	}
	return template.HTML("- " + strings.Join(parts, " "))
}

func renderWordAdd(line string) template.HTML {
	words := strings.Fields(line)
	var parts []string
	for _, w := range words {
		parts = append(parts, `<span class="diff-word-add">`+template.HTMLEscapeString(w)+`</span>`)
	}
	return template.HTML("+ " + strings.Join(parts, " "))
}

func lcsLines(a, b []string) []string {
	m, n := len(a), len(b)
	dp := make([][]int, m+1)
	for i := range dp {
		dp[i] = make([]int, n+1)
	}
	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			if a[i-1] == b[j-1] {
				dp[i][j] = dp[i-1][j-1] + 1
			} else {
				if dp[i-1][j] > dp[i][j-1] {
					dp[i][j] = dp[i-1][j]
				} else {
					dp[i][j] = dp[i][j-1]
				}
			}
		}
	}

	var result []string
	i, j := m, n
	for i > 0 && j > 0 {
		if a[i-1] == b[j-1] {
			result = append([]string{a[i-1]}, result...)
			i--
			j--
		} else if dp[i-1][j] > dp[i][j-1] {
			i--
		} else {
			j--
		}
	}
	return result
}

// Git helpers

func getDefaultBranch(repoPath string) string {
	data, err := os.ReadFile(repoPath + "/HEAD")
	if err != nil {
		return "main"
	}
	ref := strings.TrimSpace(string(data))
	ref = strings.TrimPrefix(ref, "ref: refs/heads/")
	return ref
}

func resolveRef(repoPath, branch string) string {
	data, err := os.ReadFile(repoPath + "/refs/heads/" + branch)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

func listBranches(repoPath string) []string {
	var branches []string
	headsPath := repoPath + "/refs/heads/"
	filepath.WalkDir(headsPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		name := strings.TrimPrefix(path, headsPath)
		branches = append(branches, name)
		return nil
	})
	return branches
}

func readObject(repoPath, hash string) []byte {
	path := repoPath + "/objects/" + hash[:2] + "/" + hash[2:]
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	reader, err := zlib.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil
	}
	defer reader.Close()
	content, _ := io.ReadAll(reader)
	return content
}

func readObjectContent(repoPath, hash string) []byte {
	data := readObject(repoPath, hash)
	if data == nil {
		return nil
	}
	nullPos := bytes.IndexByte(data, 0)
	if nullPos < 0 {
		return data
	}
	return data[nullPos+1:]
}

func writeObject(repoPath, objType string, content []byte) string {
	header := fmt.Sprintf("%s %d\x00", objType, len(content))
	full := append([]byte(header), content...)

	h := sha1.Sum(full)
	hash := hex.EncodeToString(h[:])

	dir := repoPath + "/objects/" + hash[:2]
	os.MkdirAll(dir, 0755)

	var buf bytes.Buffer
	w, _ := zlib.NewWriterLevel(&buf, zlib.BestCompression)
	w.Write(full)
	w.Close()

	os.WriteFile(dir+"/"+hash[2:], buf.Bytes(), 0644)
	return hash
}

func updateTree(repoPath, treeHash, filePath, newBlobHash string) string {
	parts := strings.Split(filePath, "/")
	return updateTreeRecursive(repoPath, treeHash, parts, newBlobHash)
}

func updateTreeRecursive(repoPath, treeHash string, pathParts []string, newBlobHash string) string {
	content := readObjectContent(repoPath, treeHash)
	type treeEntry struct {
		mode string
		name string
		hash []byte
	}
	var entries []treeEntry

	i := 0
	for i < len(content) {
		spacePos := bytes.IndexByte(content[i:], ' ')
		if spacePos < 0 {
			break
		}
		mode := string(content[i : i+spacePos])
		i += spacePos + 1
		nullPos := bytes.IndexByte(content[i:], 0)
		if nullPos < 0 {
			break
		}
		name := string(content[i : i+nullPos])
		i += nullPos + 1
		if i+20 > len(content) {
			break
		}
		hash := make([]byte, 20)
		copy(hash, content[i:i+20])
		i += 20
		entries = append(entries, treeEntry{mode: mode, name: name, hash: hash})
	}

	targetName := pathParts[0]
	found := false

	for idx, entry := range entries {
		if entry.name == targetName {
			found = true
			if len(pathParts) == 1 {
				hashBytes, _ := hex.DecodeString(newBlobHash)
				entries[idx].hash = hashBytes
				entries[idx].mode = "100644"
			} else {
				oldSubTreeHash := hex.EncodeToString(entry.hash)
				newSubTreeHash := updateTreeRecursive(repoPath, oldSubTreeHash, pathParts[1:], newBlobHash)
				hashBytes, _ := hex.DecodeString(newSubTreeHash)
				entries[idx].hash = hashBytes
			}
			break
		}
	}

	if !found {
		if len(pathParts) == 1 {
			hashBytes, _ := hex.DecodeString(newBlobHash)
			entries = append(entries, treeEntry{mode: "100644", name: targetName, hash: hashBytes})
		}
	}

	var treeBuf bytes.Buffer
	for _, entry := range entries {
		treeBuf.WriteString(entry.mode + " " + entry.name)
		treeBuf.WriteByte(0)
		treeBuf.Write(entry.hash)
	}

	return writeObject(repoPath, "tree", treeBuf.Bytes())
}

func parseCommitObject(repoPath, hash string) CommitInfo {
	content := readObjectContent(repoPath, hash)
	lines := strings.Split(string(content), "\n")
	commit := CommitInfo{Hash: hash}
	messageStart := false
	for _, line := range lines {
		if messageStart {
			if commit.Message == "" {
				commit.Message = strings.TrimSpace(line)
			}
			continue
		}
		if strings.HasPrefix(line, "tree ") {
			commit.Tree = strings.TrimPrefix(line, "tree ")
		} else if strings.HasPrefix(line, "parent ") {
			commit.Parent = strings.TrimPrefix(line, "parent ")
		} else if strings.HasPrefix(line, "author ") {
			authorLine := strings.TrimPrefix(line, "author ")
			parts := strings.Split(authorLine, " <")
			if len(parts) > 0 {
				commit.Author = parts[0]
			}
		} else if line == "" {
			messageStart = true
		}
	}
	return commit
}

func parseTreeObject(repoPath, hash string) map[string]string {
	content := readObjectContent(repoPath, hash)
	tree := make(map[string]string)
	i := 0
	for i < len(content) {
		spacePos := bytes.IndexByte(content[i:], ' ')
		if spacePos < 0 {
			break
		}
		i += spacePos + 1
		nullPos := bytes.IndexByte(content[i:], 0)
		if nullPos < 0 {
			break
		}
		name := string(content[i : i+nullPos])
		i += nullPos + 1
		if i+20 > len(content) {
			break
		}
		rawHash := content[i : i+20]
		hexHash := hex.EncodeToString(rawHash)
		tree[name] = hexHash
		i += 20
	}
	return tree
}

func listTree(repoPath, treeHash, prefix string) []FileEntry {
	content := readObjectContent(repoPath, treeHash)
	var entries []FileEntry
	i := 0
	for i < len(content) {
		spacePos := bytes.IndexByte(content[i:], ' ')
		if spacePos < 0 {
			break
		}
		mode := string(content[i : i+spacePos])
		i += spacePos + 1
		nullPos := bytes.IndexByte(content[i:], 0)
		if nullPos < 0 {
			break
		}
		name := string(content[i : i+nullPos])
		i += nullPos + 1
		if i+20 > len(content) {
			break
		}
		i += 20

		path := name
		if prefix != "" {
			path = prefix + "/" + name
		}
		isDir := mode == "40000"
		entries = append(entries, FileEntry{Name: name, Path: path, IsDir: isDir})
	}
	return entries
}

func resolveTreePath(repoPath, treeHash, subpath string) string {
	parts := strings.Split(subpath, "/")
	currentTree := treeHash
	for _, part := range parts {
		tree := parseTreeObject(repoPath, currentTree)
		hash, ok := tree[part]
		if !ok {
			return ""
		}
		obj := readObject(repoPath, hash)
		if obj == nil {
			return ""
		}
		header := string(obj[:bytes.IndexByte(obj, 0)])
		if strings.HasPrefix(header, "tree") {
			currentTree = hash
		} else {
			return ""
		}
	}
	return currentTree
}

func resolveBlobPath(repoPath, treeHash, filePath string) string {
	parts := strings.Split(filePath, "/")
	currentTree := treeHash
	for i, part := range parts {
		tree := parseTreeObject(repoPath, currentTree)
		hash, ok := tree[part]
		if !ok {
			return ""
		}
		if i == len(parts)-1 {
			return hash
		}
		currentTree = hash
	}
	return ""
}

func getCommitLog(repoPath, hash string, limit int) []CommitInfo {
	var commits []CommitInfo
	current := hash
	for current != "" && len(commits) < limit {
		commit := parseCommitObject(repoPath, current)
		commits = append(commits, commit)
		current = commit.Parent
	}
	return commits
}

func hashObject(data []byte) string {
	h := sha1.Sum(data)
	return hex.EncodeToString(h[:])
}

func renderPage(w http.ResponseWriter, r *http.Request, name string, data map[string]interface{}) {
	if _, ok := data["CurrentUser"]; !ok {
		data["CurrentUser"] = auth.GetCurrentUser(r)
	}
	if _, ok := data["Active"]; !ok {
		data["Active"] = ""
	}
	if _, ok := data["HideChrome"]; !ok {
		data["HideChrome"] = false
	}
	renderTemplate(w, name, data)
}

func renderTemplate(w http.ResponseWriter, name string, data interface{}) {
	tmpl, err := template.ParseFiles("web/templates/layout.html", "web/templates/"+name)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	var buf bytes.Buffer
	err = tmpl.ExecuteTemplate(&buf, "layout", data)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(buf.Bytes())
}

