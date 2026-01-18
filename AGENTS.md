# PROJECT KNOWLEDGE BASE

**Generated:** 2026-01-18
**Commit:** 8aa628d
**Branch:** main

## OVERVIEW
Blank Go module "vip-switch-go". No source code yet. Apache 2.0 licensed.

## STRUCTURE
```
vip-switch-go/
├── go.mod              # Go module definition (vip-switch-go)
├── LICENSE             # Apache 2.0 (staged, not committed)
├── .gitignore          # Excludes .idea/, *.iml
└── vip-switch-go.iml   # IntelliJ/GoLand config
```

## WHERE TO LOOK
| Task | Location | Notes |
|------|----------|-------|
| Module config | go.mod | Currently empty, no deps |
| License | LICENSE | Apache 2.0, Copyright 2026 lowezheng |

## CODE MAP
**N/A** - No Go source files exist yet.

## CONVENTIONS
**None established** - Blank project. Standard Go patterns expected:
- cmd/ for entry points
- internal/ for private packages
- pkg/ for public libraries
- *_test.go for test files

## ANTI-PATTERNS (THIS PROJECT)
**None** - No code to audit.

## UNIQUE STYLES
**None** - Blank slate project.

## COMMANDS
```bash
# Initialize Go module (already done)
go mod init vip-switch-go

# Build (when main.go exists)
go build

# Run tests (when *_test.go files exist)
go test ./...

# Add dependencies (as needed)
go get <package>
```

## NOTES
- This is a greenfield Go project with zero implementation
- LICENSE is staged but not committed
- Remote repo: https://github.com/lowezheng/vip-switch-go.git
- No CI/CD, Makefile, or Dockerfile configured yet
- Project structure should follow standard Go project layout
- Module name: vip-switch-go (consider changing to github.com/lowezheng/vip-switch-go)
