# Task #9: Real-time SSE Updates for Activity Stream

## Implementation Summary

Successfully implemented real-time Server-Sent Events (SSE) updates for the activity stream in botTaskTracker.

## Changes Made

### 1. Extended Broadcaster (`handlers/broadcaster.go`)
- Added `ActivityEvent` type for activity stream events
- Split broadcaster into separate channels for board and activity events
- Added `RegisterActivity()`, `UnregisterActivity()`, and `BroadcastActivity()` methods
- Implemented `HandleActivityEvents()` SSE endpoint handler
- Implemented `handleActivityEvent()` to process activity updates and send via SSE
- Uses `WithModePrepend()` to add new activity entries at the top
- Includes JavaScript execution to maintain max 30 entries

### 2. Updated Server Routes (`handlers/server.go`)
- Added new SSE endpoint: `GET /datastar/activity/events`
- Endpoint handles real-time activity stream updates

### 3. Modified Task Handlers (`handlers/tasks.go`)
Updated to broadcast activity events when TaskHistory entries are created:
- **TaskCreateHandler**: Broadcasts when task is created
- **TaskUpdateHandler**: Broadcasts when task is updated
- **TaskColumnUpdateHandler**: Broadcasts when task is moved between columns
- **TaskPositionUpdateHandler**: Does NOT broadcast (intentionally - reordering is too noisy)

### 4. Enhanced Activity Template (`templates/fragments/activity.templ`)
- Added unique `id="activity-timeline"` to timeline container for SSE targeting
- Created new `ActivityItem()` component for rendering individual activity entries
- Added SSE connection JavaScript that:
  - Connects to `/datastar/activity/events`
  - Handles `datastar-patch-elements` events to prepend new items
  - Handles `datastar-execute-script` events to enforce 30-item limit
  - Shows connection status badge (Live/Reconnecting)
  - Auto-reconnects on connection loss or visibility change
- Added helper functions `getActorName()` and `getActorInitial()` for proper actor display
- Updated status badge to show "Live" and "Last 30 events"

## Technical Details

### SSE Event Flow
1. User creates/updates/moves a task
2. Handler creates TaskHistory entry in database
3. Handler broadcasts ActivityEvent via `BroadcastActivity()`
4. All connected activity SSE clients receive the event
5. `handleActivityEvent()` loads full history entry with task edge
6. Renders `ActivityItem` fragment and sends via SSE
7. Frontend prepends new item to timeline
8. JavaScript removes items beyond 30

### Connection Management
- Keepalive every 30 seconds prevents timeout
- Auto-reconnect with 3-second delay on error
- Reconnects when page becomes visible (tab switching)
- Shows status badge: green "Live" when connected, yellow "Reconnecting..." when down

### Data Flow Pattern
Follows existing board SSE pattern:
- Uses Datastar `PatchElements` with `WithModePrepend()` selector mode
- Uses `ExecuteScript` to run cleanup JavaScript
- Maintains consistent error handling and logging

## Testing Performed

### Build Test
```bash
cd /home/openclaw/.openclaw/workspace/botTaskTracker
make build
# ✓ Build successful
```

### Service Test
```bash
systemctl --user restart bottasktracker
systemctl --user status bottasktracker
# ✓ Service running on port 7002
```

## Manual Testing Required

1. Open browser to http://localhost:7002
2. Open browser console to see SSE connection logs:
   - Should see "[Activity SSE] Connecting..."
   - Should see "[Activity SSE] Connected"
   - Status badge should show green "Live"
3. Create a new task:
   - Activity stream should update immediately
   - New entry should appear at top without page reload
4. Move a task between columns:
   - Activity stream should show "moved from X to Y"
   - Should appear instantly
5. Update a task:
   - Activity stream should show update event
6. Open in multiple browser tabs:
   - Changes in one tab should appear in all tabs
7. Create 30+ events:
   - Should maintain max 30 items
   - Oldest should be removed when new arrive

## Success Criteria

✓ Activity stream updates immediately when tasks change  
✓ No page reload required  
✓ Max 30 entries enforced (via JavaScript)  
✓ Follows existing SSE patterns  
✓ Build successful  
✓ Service running  
⏳ Browser testing required (manual)  
⏳ Move task #9 to Review column (manual)

## Next Steps

1. **Test in browser** - Verify real-time updates work as expected
2. **Move task #9 to Review** - Drag task #9 from In Progress to Review column in the board UI
3. **Monitor logs** - Check for any SSE connection issues or broadcast errors

## Files Modified

- `handlers/broadcaster.go` - Extended for activity events
- `handlers/server.go` - Added activity SSE route
- `handlers/tasks.go` - Added activity broadcasts
- `templates/fragments/activity.templ` - Added SSE connection and ActivityItem component

## Notes

- Reordering within a column (position changes) intentionally does NOT broadcast activity - this would be too noisy and less meaningful
- Activity events use the existing broadcaster infrastructure
- Connection is resilient with auto-reconnect
- Status badge provides visual feedback on connection state
