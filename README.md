# Bot Task Tracker

Mission Control-style kanban board for tracking bot tasks and work in progress.

## Features

- **Kanban board:** Backlog → In Progress → Review → Done
- **K:V tags:** Flexible key-value tagging (project, priority, readyToStart, type)
- **Activity feed:** Real-time stream of changes
- **Task history:** Full audit trail per card
- **Assignees:** Track who's working on what
- **Filtering:** By assignee, tag, status

## Tech Stack

- **Backend:** Go 1.25 + SQLite + entgo ORM
- **Frontend:** Datastar + Tailwind CSS + daisyUI
- **Templates:** templ for type-safe HTML
- **Real-time:** Server-Sent Events (SSE)

## Development

**Quick start:**
```bash
make run    # Build and run
```

**Other targets:**
```bash
make clean  # Remove old binaries
make build  # Build assets and generate code only
```

The `make run` target automatically:
1. Builds CSS assets with `npm run build:linux`
2. Generates templ templates with `go generate ./...`
3. Runs the server with `go run .` (no binary created)

**Manual steps:**
```bash
# Build CSS
npm run build:linux

# Generate templates
go generate ./...

# Run (never use go build - always go run)
go run .
```

Server runs on port 7002.

**Important:** Always use `go run .` instead of building binaries to avoid stale asset issues.

## Deployment

```bash
docker-compose up -d
```
