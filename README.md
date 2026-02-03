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

```bash
# Generate ent schemas
go generate ./ent

# Generate templ templates
templ generate

# Run server
go run ./cmd/server
```

Server runs on port 7002.

## Deployment

```bash
docker-compose up -d
```
