# CodeStash

A self-hosted Git-like version control system built from scratch in Go. Part of a 12-project DSA + System Design roadmap.

CodeStash implements Git internals from the ground up — content-addressable storage, DAG-based commit history, branching, and LCS-based diffing — to deeply understand how version control works under the hood.

## Features (Phase 1 - Custom Git CLI)

| Command | Description |
|---------|-------------|
| `cs init` | Initialize a new repository (`.codestash/` structure) |
| `cs hash <file>` | Hash a file and store as blob object (SHA-1 + zlib) |
| `cs cat <hash>` | Read and display object content by hash |
| `cs add <files>` | Stage files (multi-file, skips unchanged) |
| `cs commit -m "msg"` | Create commit (tree + commit object + HEAD update) |
| `cs log` | Show commit history (DAG traversal via parent pointers) |
| `cs status` | Show modified and untracked files |
| `cs branch [name]` | List branches or create new (nested paths supported) |
| `cs checkout [-b] <branch>` | Switch branch and restore working directory |
| `cs diff` | Show line-by-line diff (LCS algorithm — DP approach) |

## Architecture

```
.codestash/
├── HEAD                 # Points to current branch (ref: refs/heads/master)
├── index                # Staging area (hash filepath per line)
├── objects/             # Content-addressable store (SHA-1 → zlib compressed)
│   ├── 7f/2a9b...      # Blob, Tree, or Commit objects
│   └── ...
└── refs/
    └── heads/           # Branch files (each stores latest commit hash)
        ├── master
        └── feature/abc  # Nested branches supported
```

## Internal Data Structures & Algorithms

- **SHA-1 Hashing** — Content-addressable storage (same content = same hash)
- **Zlib Compression** — All objects stored compressed
- **DAG (Directed Acyclic Graph)** — Commit history via parent pointers
- **LCS (Longest Common Subsequence)** — Diff algorithm, implemented 3 ways:
  - Pure recursion (exponential)
  - Memoization (top-down DP)
  - Bottom-up DP with backtracking (optimal)
- **Tree Objects** — Directory snapshots (filename → blob hash mapping)
- **Index File** — Simple staging area tracking

## Object Format

```
<type> <size>\0<content>
```

- **Blob**: Raw file content
- **Tree**: `blob <hash> <filename>\n` per entry
- **Commit**: `tree <hash>\nparent <hash>\nauthor <name> <timestamp>\n\n<message>\n`

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

## Branch Validation

CodeStash implements Git-like branch conflict detection:
- Cannot create a branch that already exists
- Cannot create `feature` if `feature/abc` exists (directory conflict)
- Cannot create `feature/abc` if `feature` exists (file conflict)
- Nested branches (`feature/abc/deep`) fully supported

## Project Roadmap

This is **Project 4** of a 12-project systems programming roadmap:

| # | Project | Status |
|---|---------|--------|
| 1 | Rate Limiter | Done |
| 2 | URL Shortener | Done |
| 3 | Mini Redis | Done |
| 4 | **CodeStash** | **Phase 1 Done** |
| 5 | Search Engine | Upcoming |
| 6 | Load Balancer | Upcoming |
| 7 | Distributed Task Queue | Upcoming |
| 8 | Mini SQLite | Upcoming |
| 9 | Log Storage Engine | Upcoming |
| 10 | Mini DynamoDB | Upcoming |
| 11 | Mini Docker | Upcoming |
| 12 | Mini Kubernetes | Upcoming |

## Phases

- **Phase 1** (Done): Custom Git CLI from scratch
- **Phase 2** (Next): Real Git compatibility (same object format as `.git/`)
- **Phase 3**: Server + Git Smart HTTP Protocol (push/pull/clone)
- **Phase 4**: Web UI + Deploy (browse repos, commits, diffs in browser)

## Testing

All API endpoints tested with [ApiPad](https://github.com/NiHaLOO7/ApiPad) — a lightweight Postman alternative built from scratch.

## Tech Stack

- **Language**: Go
- **Storage**: File system (content-addressable)
- **Hashing**: crypto/sha1
- **Compression**: compress/zlib
