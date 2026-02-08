# Task #9 Verification Checklist

## Automated Checks ✓

- [x] Code compiles without errors
- [x] Service restarts successfully  
- [x] Service is running on port 7002
- [x] All files modified correctly
- [x] SSE endpoint added to routes
- [x] Broadcaster extended for activity events
- [x] Activity broadcasts added to task handlers
- [x] Frontend SSE connection implemented

## Manual Browser Tests (Required)

Access: http://localhost:7002

### SSE Connection
- [ ] Open browser console
- [ ] Verify "[Activity SSE] Connecting..." appears
- [ ] Verify "[Activity SSE] Connected" appears
- [ ] Verify status badge shows green "Live"

### Real-time Updates
- [ ] Create a new task
  - [ ] Activity stream updates immediately
  - [ ] New entry appears at top
  - [ ] No page reload needed
- [ ] Move a task between columns (drag & drop)
  - [ ] Activity shows "moved from X to Y"
  - [ ] Update is instant
- [ ] Edit a task (title/description)
  - [ ] Activity shows "updated"
  - [ ] Appears immediately

### Multi-client Test
- [ ] Open board in 2 browser tabs
- [ ] Create/move task in tab 1
- [ ] Verify activity updates in tab 2
- [ ] Both show same real-time updates

### Capacity Test
- [ ] Create 30+ tasks/moves
- [ ] Verify only 30 items remain in activity stream
- [ ] Oldest entries are removed automatically

### Resilience Test
- [ ] Close SSE connection (Dev Tools -> Network -> Offline)
- [ ] Verify status badge shows "Reconnecting..."
- [ ] Restore connection
- [ ] Verify badge returns to green "Live"
- [ ] Verify updates resume

## Final Task

- [ ] **Move task #9 to Review column**
  - Drag task #9 from current column to "Review" column
  - Verify it broadcasts an activity event
  - This marks task #9 as complete

## Expected Console Logs

```
[Board SSE] Connecting...
[Board SSE] Connected
[Activity SSE] Connecting...
[Activity SSE] Connected
[Activity SSE] Received patch-elements
[Activity SSE] Prepended to #activity-timeline
```

## Expected Broadcast Logs (Server)

Check `systemctl --user status bottasktracker` or logs:
```
BroadcastActivity: activity_created historyID: X clients: Y sent: Y skipped: 0
```

## Known Behaviors

✓ Reordering within same column does NOT create activity (intentional - too noisy)  
✓ Connection auto-reconnects after 3 seconds on error  
✓ Keepalive every 30 seconds prevents timeout  
✓ Activity badge color indicates connection state
