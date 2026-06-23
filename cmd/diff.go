package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/NiHaLOO7/codestash/core"
	"github.com/NiHaLOO7/codestash/storage"
)

func Diff() {
	head := core.GetHead()
	commitData, _ := storage.ReadObject(head)
	commit := core.ParseCommit(commitData)
	treeData, _ := storage.ReadObject(commit.Tree)
	tree := core.ParseTree(treeData)
	for filename, blobHash := range tree {
		blob, _ := storage.ReadObject(blobHash)
		oldContent := core.ParseObject(blob)
		newContent, _ := os.ReadFile(filename)
		if string(oldContent) == string(newContent) {
			continue
		}
		result := diffLineDP(strings.Split(string(oldContent), "\n"),
			strings.Split(string(newContent), "\n"))
		fmt.Println("--- a/" + filename)
		fmt.Println("+++ b/" + filename)

		for _, line := range result {
			fmt.Println(line)
		}
	}
}


func diffLineDP(old, new []string) []string {
	m := len(old)
	n := len(new)
	dp := make([][]int, m+1)
	for i := 0; i <= m; i++ {
    	dp[i] = make([]int, n+1)
	}
	for i := m-1; i >= 0; i-- {
		for j := n-1; j >= 0; j-- {
			if old[i] == new[j] {
				dp[i][j] = dp[i+1][j+1] + 1
			} else {
				dp[i][j] = max(dp[i+1][j], dp[i][j+1])
			}
		}
	}
	i := 0
	j := 0
	var result []string
	for i < m && j < n {
		if old[i] == new[j] {
			result = append(result, " " + old[i])
			i++
			j++
		} else if dp[i+1][j] >= dp[i][j+1] {
			result = append(result, "- " + old[i])
			i++
		} else {
			result = append(result, "+ " + new[j])
			j++
		}
	}
	for i < m {
		result = append(result, "- " + old[i])
		i++
	}
	for j < n {
		result = append(result, "+ " + new[j])
		j++
	}
	return result
}

// func prefixLines(lines []string, prefix string) []string {
// 	var result []string
// 	for _, line := range lines {
// 		result = append(result, prefix+line)
// 	}
// 	return result
// }

// func countChanges(lines []string) int {
// 	count := 0
// 	for _, line := range lines {
// 		if strings.HasPrefix(line, "+") || strings.HasPrefix(line, "-") {
// 			count++
// 		}
// 	}
// 	return count
// }

// func diffLines(old, new []string) []string {
// 	memo := make(map[string][]string)
// 	var lcs func(i, j int) []string
// 	lcs = func(i int, j int) []string {

// 		if i == len(old) {
// 			return prefixLines(new[j:], "+")
// 		}
// 		if j == len(new) {
// 			return prefixLines(old[i:], "-")
// 		}

// 		key := fmt.Sprintf("%d_%d", i, j)
// 		data, exists := memo[key]
// 		if exists {
// 			return data
// 		}

// 		if old[i] == new[j] {
// 			rest := lcs(i+1, j+1)
// 			memo[key] = append([]string{" " + old[i]}, rest...)
// 			return memo[key]
// 		}
// 		choice1 := lcs(i+1, j)
// 		choice2 := lcs(i, j+1)
// 		choice1 = append([]string{"- " + old[i]}, choice1...)
// 		choice2 = append([]string{"+ " + new[j]}, choice2...)
// 		if countChanges(choice1) <= countChanges(choice2) {
// 			memo[key] = choice1
// 			return memo[key]
// 		}
// 		memo[key] = choice2
// 		return memo[key]
// 	}

// 	return lcs(0, 0)
// }

