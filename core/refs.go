package core

import (
	"os"
	"fmt"
	"io/fs"
	"strings"
	"path/filepath"
)

const HEAD_PATH = ".codestash/HEAD"
const BRANCHES_PATH = ".codestash/refs/heads/"

// Parent
func GetHead() string {
	_, err := os.Stat(PARENT)
	if err != nil {
		fmt.Println("Git not initialized")
        return ""
	}
	head, _ := os.ReadFile(HEAD_PATH)
	ref := strings.TrimSpace(string(head))
	branchPath := strings.TrimPrefix(ref, "ref: ")
	commitHash, err := os.ReadFile(PARENT+"/"+branchPath)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(commitHash))
}

// Update Parent
func UpdateHead(hash string) {
	_, err := os.Stat(PARENT)
	if err != nil {
		fmt.Println("Git not initialized")
        return
	}
	head, _ := os.ReadFile(HEAD_PATH)
	ref := strings.TrimSpace(string(head))
	branchPath := strings.TrimPrefix(ref, "ref: ")
	os.WriteFile(PARENT+"/"+branchPath, []byte(hash), 0644)
}

func CurrentBranch() string {
	_, err := os.Stat(PARENT)
	if err != nil {
		fmt.Println("Git not initialized")
        return ""
	}
	head, _ := os.ReadFile(HEAD_PATH)
	ref := strings.TrimSpace(string(head))
	path := strings.TrimPrefix(ref, "ref: refs/heads/")
	return path
}

func ListBranches() []string {
	_, err := os.Stat(PARENT)
	if err != nil {
		fmt.Println("Git not initialized")
        return nil
	}
	var branches []string
	filepath.WalkDir(BRANCHES_PATH, 
		func(path string, d fs.DirEntry, err error) error {
			if err != nil {
            	return err
        	}
        	if d.IsDir() {
            	return nil
        	} else {
            	name := strings.TrimPrefix(path, BRANCHES_PATH)
				branches = append(branches, name)
        	}
        	return nil
		})
	return branches
}

func CreateBranch(name string) {
	head := GetHead()
	info, err := os.Stat(BRANCHES_PATH + name)
	if err == nil {
		if info.IsDir() {
        	fmt.Println("conflicts with existing branches under that path")
    	} else {
        	fmt.Println("branch already exists")
    	}
    	return
	}
	parts := strings.Split(name, "/")
	parts = parts[:len(parts)-1]
	var current_dict strings.Builder; 
	current_dict.WriteString(BRANCHES_PATH)
	for _, part := range parts {
		current_dict.WriteString(part);
		current_dict.WriteString("/")
		info, err := os.Stat(strings.TrimSuffix(current_dict.String(),"/"))
		if err == nil && !info.IsDir() {
			fmt.Println("branch already exists")
        	return
		} else if err == nil && info.IsDir() {
			continue
		}
	}
	os.MkdirAll(filepath.Dir(BRANCHES_PATH + name), 0755)
	os.WriteFile(BRANCHES_PATH + name, []byte(head), 0644)
}


func SwitchBranch(name string) {
	_, err := os.Stat(PARENT)
	if err != nil {
		fmt.Println("Git not initialized")
        return
	}
	_, err = os.Stat(BRANCHES_PATH + name)
	if err == nil {
		os.WriteFile(PARENT+"/HEAD", []byte("ref: refs/heads/"+name+"\n"), 0644)
		return
	}
	fmt.Printf("Branch %s does not exist", name)
}