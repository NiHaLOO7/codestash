# CodeStash

A lightweight, self-hosted version control system built from scratch in Go. Designed to be a simple alternative to Git with familiar commands and a clean architecture.

## Commands

| Command | Description |
|---------|-------------|
| `cs init` | Initialize a new repository |
| `cs hash <file>` | Hash and store a file as blob object |
| `cs cat <hash>` | Read and display object content by hash |
| `cs add <files>` | Stage files for commit |
| `cs commit -m "msg"` | Create a new commit |
| `cs log` | Show commit history |
| `cs status` | Show modified and untracked files |
| `cs branch [name]` | List or create branches |
| `cs checkout [-b] <branch>` | Switch branches |
| `cs diff` | Show line-by-line changes |

## Quick Start

```bash
# Build
go build -o cs .

# Initialize repo
./cs init

# Create and track files
echo "hello" > file.txt
./cs add file.txt
./cs commit -m "initial commit"

# View history
./cs log

# Branching
./cs branch feature/new-thing
./cs checkout feature/new-thing

# Make changes and diff
echo "changed" > file.txt
./cs diff
```

## Architecture

```
.codestash/
├── HEAD                 # Points to current branch
├── index                # Staging area
├── objects/             # Content-addressable object store (SHA-1 + zlib)
│   ├── 7f/2a9b...      # Blob, Tree, or Commit objects
│   └── ...
└── refs/
    └── heads/           # Branch references
        ├── master
        └── feature/abc
```

## Branch Validation

- Cannot create a branch that already exists
- Cannot create `feature` if `feature/abc` exists (directory conflict)
- Cannot create `feature/abc` if `feature` exists (file conflict)
- Nested branches fully supported

## Roadmap

- **Phase 1** (Done): Core CLI commands
- **Phase 2** (Next): Real Git compatibility
- **Phase 3**: Server + Git Smart HTTP Protocol (push/pull/clone)
- **Phase 4**: Web UI + Deploy

## Testing

Tested with [ApiPad](https://github.com/NiHaLOO7/ApiPad) — a lightweight Postman alternative.

## Tech Stack

- **Language**: Go
- **Storage**: File system (content-addressable)
- **Hashing**: crypto/sha1
- **Compression**: compress/zlib
