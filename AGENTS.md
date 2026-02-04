# AGENTS.md - botTaskTracker

## Project Overview
Mission Control-style kanban board for tracking bot tasks and work in progress.

## Tech Stack
- **Backend**: Go
- **Frontend**: 
  - daisyUI + TailwindCSS (via CDN currently)
  - Datastar (hypermedia-first, backend-driven UI)
  - templ (Go templating)
- **Database**: SQLite via entgo

## Key Patterns

### Datastar
Backend-driven UI using SSE. Key concepts:
- Server sends HTML fragments via Server-Sent Events
- `data-*` attributes for reactivity
- Signals (`$foo`) for frontend state
- Actions (`@get()`, `@post()`) for backend communication

See `docs/DATASTAR.md` for full reference.

### daisyUI
CSS component library for Tailwind CSS. Provides semantic class names like:
- `btn`, `card`, `badge`, `modal`
- Colors: `primary`, `secondary`, `accent`, `neutral`, `base-*`
- Sizes: `btn-sm`, `btn-lg`, etc.

See `docs/DAISYUI.md` for full reference.

## Project Structure
```
botTaskTracker/
├── cmd/server/main.go    # Entry point
├── ent/                  # Database schemas
│   └── schema/
│       ├── task.go
│       ├── taskhistory.go
│       └── tasktag.go
├── handlers/             # HTTP handlers
├── templates/            # templ templates
│   ├── main.templ       # Base layout
│   └── pages/
│       └── board.templ  # Kanban board
├── static/              # Static assets
│   ├── scripts/
│   │   └── datastar.js
│   └── styles.css
└── data/                # SQLite database
```

## Build & Run
```bash
# Generate templ files
templ generate

# Build
/usr/local/go/bin/go build -o botTaskTracker

# Run (port 7002)
./botTaskTracker
```

## Kanban Columns
1. Backlog
2. In Progress  
3. Review
4. Done

## Key Features
- Task cards with K:V tags
- Activity feed
- Full task history
- Assignee filtering (peter, john, jane)

## Styling Notes
- Using daisyUI "corporate" theme (minimal/professional)
- Horizontal 4-lane kanban layout
- Datastar served locally from `/static/scripts/datastar.js`
