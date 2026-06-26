# CodeStash

A lightweight, self-hosted version control system built from scratch in Go. Fully compatible with real Git — objects, index, commits, and branches are cross-readable between `cs` and `git` CLI.

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
| `cs serve` | Start Git HTTP server on :8080 |

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

# Start HTTP server (serves repos from repos/ directory)
./cs serve

# Clone from server (from another terminal)
git clone http://localhost:8080/myrepo.git
```

## Git Compatibility

CodeStash uses `.git` as its directory and writes all objects in real Git binary format:

- **Index**: Binary DIRC format (v2) — `git status` reads it natively
- **Trees**: Binary `<mode> <name>\x00<20-byte-hash>` format
- **Commits**: Standard `author/committer <name> <email> <timestamp> <tz>` format
- **Objects**: SHA-1 + zlib compressed, stored in `objects/xx/xxx...` structure

**Cross-compatibility verified:**
- `cs commit` → `git log` reads it ✓
- `git commit` → `cs log` reads it ✓
- `git status` clean after `cs add` + `cs commit` ✓
- `git cat-file` reads all cs objects ✓

## Architecture

```
.git/
├── HEAD                 # Points to current branch
├── config               # User config ([user] section)
├── index                # Binary staging area (DIRC v2)
├── objects/             # Content-addressable object store (SHA-1 + zlib)
│   ├── 7f/2a9b...      # Blob, Tree, or Commit objects
│   └── ...
└── refs/
    └── heads/           # Branch references
        ├── master
        └── feature/abc
```

## Git HTTP Server

CodeStash includes a built-in Git Smart HTTP server. Any standard `git` client can clone from it:

```bash
# Start server
./cs serve

# Clone from another machine
git clone http://localhost:8080/myrepo.git
```

Supports:
- `GET /info/refs` — reference discovery
- `POST /git-upload-pack` — clone/pull (packfile transfer)
- Multi-repo serving from `repos/` directory

## Branch Validation

- Cannot create a branch that already exists
- Cannot create `feature` if `feature/abc` exists (directory conflict)
- Cannot create `feature/abc` if `feature` exists (file conflict)
- Nested branches fully supported

## Testing

Tested with [ApiPad](https://github.com/NiHaLOO7/ApiPad) — a lightweight Postman alternative.

## Tech Stack

- **Language**: Go
- **Storage**: File system (content-addressable)
- **Hashing**: crypto/sha1
- **Compression**: compress/zlib
- **Binary Encoding**: encoding/binary (big-endian), encoding/hex
- **Index Format**: Git DIRC v2 (binary)
- **HTTP Server**: net/http (Git Smart HTTP Protocol)
- **Packfile**: Custom packfile creation with variable-length encoding
