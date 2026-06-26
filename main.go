package main

import (
	"os"
	"github.com/NiHaLOO7/codestash/cmd"
	"github.com/NiHaLOO7/codestash/server"
	"github.com/NiHaLOO7/codestash/web"
)

func main() {
	command := os.Args[1]

	switch command {
	case "init":
		cmd.Init()
	case "hash":
		cmd.HashObject(os.Args[2])
	case "cat":
		cmd.CatFile(os.Args[2])
	case "add":
		cmd.Add(os.Args[2:])
	case "commit":
		cmd.Commit(os.Args[2:])
	case "log":
		cmd.Log()
	case "status":
		cmd.Status()
	case "branch":
		cmd.Branch(os.Args[2:])
	case "checkout":
		cmd.Checkout(os.Args[2:])
	case "diff":
		cmd.Diff()
	case "serve":
		server.StartServer()
	case "web":
		web.StartWeb("repos")
	}
}
