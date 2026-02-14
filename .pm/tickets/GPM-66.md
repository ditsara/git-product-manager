---
assignee: ""
blocks: []
created_at: "2026-02-14T08:49:08Z"
depends_on: []
id: GPM-66
labels:
    - infrastructure
    - devops
parent: ""
points: 3
priority: high
related: []
status: done
title: Setup DevContainer for isolated Copilot CLI
type: task
updated_at: "2026-02-14T09:18:42Z"
---


# Description

**[Claude Sonnet 4.5]**

## Problem Statement

We want to use GitHub Copilot CLI with "yolo permissions" (let it run commands, modify files, etc.) but safely isolated from the host system. The solution is a container that:
- Runs in the background with the codebase mounted
- Hosts Copilot CLI inside the container
- Allows editing with system Neovim (on host, not in container)
- Provides wrapper scripts to invoke Copilot CLI inside the container

## Architecture Overview

```
┌─────────────────────────────────────────────────────────┐
│ HOST SYSTEM                                             │
│                                                         │
│  ┌──────────────┐         ┌─────────────────────────┐  │
│  │ Your Neovim  │────────▶│  Project Directory      │  │
│  │ (editing)    │  edits  │  /home/user/gpm/        │  │
│  └──────────────┘         └──────────┬──────────────┘  │
│                                      │                  │
│                                      │ mounted as       │
│                                      │ /workspace       │
│  ┌──────────────┐                   │                  │
│  │ Wrapper      │                   │                  │
│  │ Script       │───────────────────┼─────────────┐    │
│  │ `dev-copilot`│   docker exec     │             │    │
│  └──────────────┘                   │             │    │
│                                      │             │    │
│        │                             ▼             ▼    │
│        │              ┌──────────────────────────────┐  │
│        │              │ CONTAINER (gpm-dev)          │  │
│        │              │                              │  │
│        │              │  /workspace (mounted)        │  │
│        └─────────────▶│  ├── cmd/                    │  │
│         Copilot CLI   │  ├── internal/               │  │
│         commands      │  └── ...                     │  │
│                       │                              │  │
│                       │  GitHub Copilot CLI          │  │
│                       │  (gh copilot suggest ...)    │  │
│                       │  ✅ Can modify files         │  │
│                       │  ✅ Can run commands         │  │
│                       │  ❌ Can't touch ~/Documents  │  │
│                       │  ❌ Can't touch ~/.ssh       │  │
│                       └──────────────────────────────┘  │
└─────────────────────────────────────────────────────────┘
```

## Solution Approach

**Phase 1: Container Setup**
- Create `.devcontainer/Dockerfile` with Go + SQLite + GitHub CLI
- Create `docker-compose.yml` for persistent container
- Container runs in background with codebase mounted
- Add `pm` to $PATH inside container (points to `/workspace/bin/pm`)

**Phase 2: Copilot CLI Inside Container**
- Install GitHub Copilot CLI via npm (`@github/copilot`) in container
- Authenticate once (persists in container volume)
- All Copilot commands run inside container context

**Phase 3: Editor Workflow**
- Edit files with system Neovim (on host)
- Changes instantly visible in container (volume mount)
- Run tests/builds inside container
- Copilot modifications happen in mounted volume (persist to host)

## How Copilot CLI Works

### Where Does Copilot CLI Run?

**INSIDE THE CONTAINER** - This is the key to safe isolation.

**Why inside container:**
✅ **Safe to give yolo permissions** - Can only affect `/workspace`  
✅ **Can't access host files** - `~/Documents`, `~/.ssh`, etc. are invisible  
✅ **Disposable** - If it goes rogue, just `docker-compose down`  
✅ **Authenticated once** - Credentials persist in container volume  

**Why NOT on host:**
❌ **No isolation** - Copilot has full access to your entire system  
❌ **Defeats the purpose** - We want container isolation  

### How Do You Use It?

**Wrapper Scripts (Recommended)**

Create `scripts/dev-copilot.sh`

### Authentication Flow

**One-time setup (persists across container restarts):**

```bash
# Start container
docker-compose -f .devcontainer/docker-compose.yml up -d

# Exec into it and authenticate
docker-compose -f .devcontainer/docker-compose.yml exec dev copilot
# Follow interactive authentication flow...

# Exit container
exit

# Now you can use the wrapper from host!
./scripts/dev-copilot.sh
```

**Authentication persists because:**
- Docker Compose uses a named volume for `/home/vscode/.copilot`
- GitHub Copilot CLI stores auth in `~/.copilot/`
- Volume survives `docker-compose down` (but not `docker-compose down -v`)

## Edge Cases

**What if I want to use Copilot on host too?**
- Install `@github/copilot` via npm separately on host
- Use `./scripts/dev-copilot.sh` for container-isolated work
- Use `copilot` for general host work

**What if authentication expires?**
- Just re-run `copilot` inside container to re-authenticate
- Or delete container volume and start fresh

**What if I'm editing a file and Copilot modifies it?**
- Neovim will detect external change
- Reload with `:e` or `:checktime`
- Use `:set autoread` in Neovim for automatic reload

**What if the container stops?**
- Restart: `docker-compose up -d`
- Wrapper script should check if container is running

**Can Copilot see my entire git history?**
- Yes - the entire project directory is mounted
- This is intentional (Copilot needs context)
- But it can't see parent directories or other projects

## Implementation Steps

### Step 1: Create DevContainer Files

- [x] Create `.devcontainer/devcontainer.json`
  ```json
  {
    "name": "GPM Dev Container",
    "dockerComposeFile": "docker-compose.yml",
    "service": "dev",
    "workspaceFolder": "/workspace",
    "customizations": {
      "vscode": {
        "extensions": ["golang.go", "GitHub.copilot"]
      }
    }
  }
  ```

- [x] Create `.devcontainer/Dockerfile`
  ```dockerfile
  FROM golang:1.24.4-bookworm

  # Install dependencies
  RUN apt-get update && apt-get install -y \
      git sqlite3 make curl sudo \
      ca-certificates gnupg \
      && rm -rf /var/lib/apt/lists/*

  # Install NodeJS 22.x using the official NodeSource setup
  RUN curl -fsSL https://deb.nodesource.com/setup_22.x | bash - \
      && apt-get install -y nodejs \
      && rm -rf /var/lib/apt/lists/*

  # Install GitHub Copilot CLI globally
  RUN npm install -g @github/copilot

  # Create non-root user
  ARG USERNAME=vscode
  ARG USER_UID=1000
  ARG USER_GID=$USER_UID
  RUN groupadd --gid $USER_GID $USERNAME \
      && useradd --uid $USER_UID --gid $USER_GID -m $USERNAME \
      && echo "$USERNAME ALL=(ALL) NOPASSWD:ALL" >> /etc/sudoers

  # Pre-create the .copilot directory with correct ownership
  RUN mkdir -p /home/vscode/.copilot && \
      chown -R vscode:vscode /home/vscode/.copilot

  USER $USERNAME
  WORKDIR /workspace

  # Add pm binary to PATH (points to /workspace/bin/pm)
  ENV PATH="/workspace/bin:${PATH}"
  ```

- [x] Create `.devcontainer/docker-compose.yml`
  ```yaml
  version: '3.8'
  
  services:
    dev:
      build:
        context: ..
        dockerfile: .devcontainer/Dockerfile
        args:
          USER_UID: ${UID:-1000}
          USER_GID: ${GID:-1000}
      volumes:
        - ..:/workspace:cached
        - copilot-config:/home/vscode/.copilot
        - go-cache:/home/vscode/go
      working_dir: /workspace
      stdin_open: true
      tty: true
      command: sleep infinity
  
  volumes:
    copilot-config:  # Persists GitHub Copilot CLI auth
    go-cache:        # Persists Go modules cache
  ```

### Step 2: Create Wrapper Script

- [x] Create `scripts/dev-copilot.sh`
  ```bash
  #!/bin/bash
  docker-compose -f .devcontainer/docker-compose.yml exec dev copilot
  ```

- [x] Make executable: `chmod +x scripts/dev-copilot.sh`

- [x] Add to `/workspace/bin` to PATH (optional)

### Step 3: Initial Setup

- [x] Build and start container:
  ```bash
  docker-compose -f .devcontainer/docker-compose.yml build
  docker-compose -f .devcontainer/docker-compose.yml up -d
  ```

- [x] Authenticate Copilot (one-time):
  ```bash
  docker-compose -f .devcontainer/docker-compose.yml exec dev copilot
  # Follow interactive authentication flow...
  ```

- [x] Test wrapper:
  ```bash
  ./scripts/dev-copilot.sh
  ```

### Step 4: Add Convenience Scripts

- [x] Create `scripts/dev-start.sh`
  ```bash
  #!/bin/bash
  docker-compose -f .devcontainer/docker-compose.yml up -d
  echo "Dev container started. Use './scripts/dev-copilot.sh' to interact with Copilot CLI."
  ```

- [x] Create `scripts/dev-stop.sh`
  ```bash
  #!/bin/bash
  docker-compose -f .devcontainer/docker-compose.yml down
  echo "Dev container stopped. Auth is preserved in volumes."
  ```

- [x] Create `scripts/dev-shell.sh` (for debugging)
  ```bash
  #!/bin/bash
  docker-compose -f .devcontainer/docker-compose.yml exec dev bash
  ```

### Step 5: Update Documentation

- [x] Add to `.gitignore`:
  ```
  .devcontainer/.env
  ```

- [x] Update README.md with:
  ```markdown
  ## Development with Copilot CLI
  
  This project uses a containerized Copilot CLI for safe AI assistance.
  
  ### Setup (one-time)
  ```bash
  # Start container
  docker-compose -f .devcontainer/docker-compose.yml up -d
  
  # Authenticate Copilot
  docker-compose -f .devcontainer/docker-compose.yml exec dev copilot
  # Follow interactive authentication...
  ```
  
  ### Daily Usage
  ```bash
  # Start container (if not running)
  ./scripts/dev-start.sh
  
  # Edit code with your system Neovim
  nvim cmd/pm/list.go
  
  # Ask Copilot for help
  ./scripts/dev-copilot.sh
  
  # Stop container (optional - can leave running)
  ./scripts/dev-stop.sh
  ```
  ```

## Acceptance Criteria

- [x] DevContainer files created (`.devcontainer/`)
- [x] Container builds successfully
- [x] Container runs in background (docker-compose)
- [x] GitHub Copilot CLI installed in container
- [x] `pm` command available in container $PATH (points to `/workspace/bin/pm`)
- [x] Authentication persists across container restarts
- [x] Wrapper script `dev-copilot.sh` works from host
- [x] Can edit files with system Neovim (changes visible in container)
- [x] Copilot can modify files (changes persist to host)
- [x] Container cannot access host files outside project directory
- [x] Documentation updated

## Workflow Summary

```bash
# ONE-TIME SETUP
docker-compose -f .devcontainer/docker-compose.yml build
docker-compose -f .devcontainer/docker-compose.yml up -d
docker-compose -f .devcontainer/docker-compose.yml exec dev copilot
# Follow authentication flow...

# DAILY WORKFLOW
# 1. Ensure container is running
./scripts/dev-start.sh

# 2. Edit code on host
nvim cmd/pm/blocked.go

# 3. Use Copilot from container
./scripts/dev-copilot.sh

# 4. Build and run pm inside container
docker-compose -f .devcontainer/docker-compose.yml exec dev make build
docker-compose -f .devcontainer/docker-compose.yml exec dev pm list

# 5. Run tests in container (optional)
docker-compose -f .devcontainer/docker-compose.yml exec dev go test ./...

# 6. Stop container when done (or leave running)
./scripts/dev-stop.sh
```

**Note:** The `pm` command is available inside the container because `/workspace/bin` is in $PATH. Build the binary once with `make build` (or `go build -o bin/pm ./cmd/pm`), and it will be accessible as `pm` from anywhere in the container.

## Security Benefits

✅ **Filesystem isolation:** Copilot can only see/modify `/workspace`  
✅ **Can't access:** `~/.ssh`, `~/Documents`, other projects  
✅ **Disposable:** Nuke container if anything goes wrong  
✅ **Persistent auth:** No need to re-authenticate daily  
✅ **Editor stays clean:** System Neovim unmodified, no plugins needed  

## Alternative Approaches Considered

**Option 1: ❌ Run Copilot on host**
- No isolation - defeats the purpose

**Option 2: ❌ Edit inside container with Neovim**
- Requires mounting Neovim config
- Less portable (Vim vs Neovim vs other editors)
- More complex

**Option 3: ✅ This approach (Container for Copilot, host for editing)**
- Best isolation
- Editor-agnostic
- Clean separation of concerns

## Notes

- Container runs 24/7 (or start/stop as needed)
- Low resource usage when idle
- Authentication token is the only sensitive data in container
- If paranoid: Use `--network=none` in docker-compose (but Copilot needs internet)
