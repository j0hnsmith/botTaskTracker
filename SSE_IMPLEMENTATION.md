# Real-Time Board Updates via SSE

## Overview

This document describes the implementation of real-time board updates using Server-Sent Events (SSE) in the botTaskTracker application.

## Architecture

### Backend Components

#### 1. Broadcaster (`handlers/broadcaster.go`)

The broadcaster manages SSE connections and broadcasts board events to all connected clients.

**Key Components:**

- `BoardEvent`: Represents a board change event with type, task ID, and column
- `Broadcaster`: Manages client connections and event distribution
- `HandleBoardEvents`: HTTP handler for SSE endpoint
- `handleBoardEvent`: Processes and renders events for specific clients

**Event Types:**
- `task_created`: New task added to board
- `task_updated`: Task properties modified
- `task_moved`: Task moved between columns (drag-and-drop)
- `task_deleted`: Task removed from board

#### 2. Event Broadcasting

Modified task handlers to broadcast events:

- `TaskCreateHandler`: Broadcasts `task_created` after successful creation
- `TaskUpdateHandler`: Broadcasts `task_updated` after task modification
- `TaskColumnUpdateHandler`: Broadcasts `task_moved` when tasks are dragged
- `TaskDeleteHandler`: Broadcasts `task_deleted` after task removal

### Frontend Components

#### Client-Side SSE Connection (`templates/pages/board.templ`)

**Features:**

1. **Auto-connect**: Establishes SSE connection on page load
2. **Auto-reconnect**: Reconnects after 3 seconds if connection drops
3. **Visibility handling**: Reconnects when page becomes visible
4. **Cleanup**: Closes connection on page unload

**Connection Flow:**

```javascript
EventSource → /datastar/board/events → Datastar processes events → DOM updates
```

## Data Flow

1. User performs action (create/update/move/delete task)
2. Handler processes request and updates database
3. Handler broadcasts event to all connected clients via Broadcaster
4. Each client's SSE handler receives event
5. Server renders updated task card HTML
6. Datastar processes SSE event and updates DOM
7. Users see live updates without page refresh

## Key Benefits

- **No polling**: Efficient server-push model
- **Automatic reconnection**: Resilient to network issues
- **Minimal client code**: Leverages Datastar's SSE handling
- **Scalable**: Broadcast pattern supports multiple concurrent users
- **Debounced updates**: Built-in through SSE event stream

## Testing

### Manual Testing

1. Open board in two browser windows
2. Create/update/move task in one window
3. Verify changes appear immediately in second window
4. Test reconnection by pausing/resuming browser tab

### Testing SSE Endpoint

```bash
# Connect to SSE stream
curl -N http://localhost:7002/datastar/board/events

# Should see:
event: datastar-patch-signals
data: signals {"boardConnected": true}
```

## Future Enhancements

Potential improvements:

1. **Debouncing**: Batch rapid updates to reduce flicker
2. **Optimistic updates**: Show changes immediately while syncing
3. **User presence**: Show who's currently viewing the board
4. **Conflict resolution**: Handle simultaneous edits gracefully
5. **Activity notifications**: Toast messages for remote updates

## Technical Details

### SSE Endpoint

- **URL**: `/datastar/board/events`
- **Method**: GET
- **Content-Type**: `text/event-stream`
- **Events**: Datastar SSE protocol (patch-elements, patch-signals)

### Datastar Integration

Uses Datastar's Go SDK for SSE:

- `datastar.NewSSE(w, r)`: Creates SSE generator
- `sse.PatchElements()`: Updates DOM elements
- `sse.PatchSignals()`: Updates frontend state
- `sse.RemoveElement()`: Removes elements from DOM

### Reconnection Strategy

- **Initial retry**: 3 seconds after disconnect
- **Visibility-based**: Reconnects when page becomes visible
- **Connection guard**: Prevents multiple simultaneous connections

## Dependencies

- `github.com/starfederation/datastar-go` v1.0.3
- Native browser EventSource API
- Templ for HTML templating
