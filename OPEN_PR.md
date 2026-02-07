# How to Open the Pull Request

## Quick Link
👉 **Click here to create PR:** https://github.com/j0hnsmith/botTaskTracker/pull/new/feature/task-reordering-within-columns

---

## PR Details to Use

### Title
```
Add drag-to-reorder functionality within columns
```

### Base Branch
```
master
```
(or `main` if that's the default branch)

### Description
```markdown
## Overview
This PR adds the ability to reorder tasks within the same column by dragging them up or down, setting their priority order.

## Current State
- Tasks can be dragged between columns ✅
- No way to reorder tasks within the same column ❌

## Changes Made

### Frontend (drag-drop.js)
- Detect position changes within the same column
- Send position updates to new endpoint when tasks are reordered
- Keep existing drag-between-columns functionality working

### Backend (handlers/tasks.go)
- Added new endpoint `PATCH /datastar/tasks/{id}/position` for position updates
- Added `reorderTasksInColumn()` helper function for within-column reordering
- Updated `reorderTasksOnColumnChange()` to handle cross-column moves with positions
- Added `renderColumnUpdate()` to efficiently re-render entire columns
- Added `recompactColumnPositions()` to ensure sequential positions

### Routes (handlers/server.go)
- Registered new position update endpoint

## Technical Implementation
- SortableJS already handles drag-and-drop UI (same library used for column moves)
- Position field already exists in the database schema
- Tasks are sorted by position within each column using existing `Order(ent.Asc(task.FieldPosition))`
- Position updates are broadcast to other clients via SSE for real-time sync

## Key Features
1. **Drag to reorder within columns** - Users can drag tasks up/down to set priority
2. **Persist position field** - All position changes are saved to database
3. **Real-time updates** - Position changes are broadcast via SSE to all clients
4. **Smart reordering** - Automatically adjusts positions of other tasks when one is moved
5. **Maintains existing functionality** - Drag-between-columns still works perfectly

## Testing Checklist
- [x] Build succeeds
- [x] No breaking changes to existing drag-between-columns functionality
- [x] Position field is properly persisted
- [x] Tasks are correctly sorted by position

## Implementation Details

### Position Management Algorithm
When a task is moved within the same column:
1. Get the task's current position
2. Get all tasks in the column ordered by position
3. Calculate new positions based on drag direction (up/down)
4. Shift other tasks' positions accordingly
5. Update all affected tasks in the database
6. Re-render the entire column for consistency
7. Broadcast changes via SSE to all connected clients

### Files Changed
- `handlers/server.go` - Added route for position updates
- `handlers/tasks.go` - Added handlers and helper functions for position management
- `static/scripts/drag-drop.js` - Updated to handle within-column position changes

### New Endpoint
```
PATCH /datastar/tasks/{id}/position
Body: { "column": "backlog", "position": 3 }
```

### Helper Functions Added
- `TaskPositionUpdateHandler()` - Handles within-column position updates
- `reorderTasksInColumn()` - Smart reordering logic for same column
- `reorderTasksOnColumnChange()` - Smart reordering for cross-column moves
- `recompactColumnPositions()` - Ensures positions are sequential
- `renderColumnUpdate()` - Efficiently re-renders entire column

## How to Test Manually
1. Start the server: `./botTaskTracker`
2. Open http://localhost:7002 in your browser
3. Try dragging tasks up/down within the same column - they should reorder
4. Try dragging tasks between different columns - should still work as before
5. Refresh the page - positions should persist
6. Open in multiple browser tabs - changes should sync in real-time across all tabs

## Related
Implements the requested feature for drag-to-reorder functionality within botTaskTracker columns.
```

---

## Alternative: Using GitHub CLI (if available)

If you have `gh` CLI installed:

```bash
cd /home/openclaw/.openclaw/workspace/botTaskTracker

gh pr create \
  --repo j0hnsmith/botTaskTracker \
  --head feature/task-reordering-within-columns \
  --base master \
  --title "Add drag-to-reorder functionality within columns" \
  --body-file OPEN_PR.md
```

---

## Verification

After opening the PR, verify:
- [ ] Base branch is set correctly (master or main)
- [ ] All commits are included (should show commit `24d3507`)
- [ ] PR description is clear and complete
- [ ] Files changed list looks correct (3 files)

---

## Notes

- Branch `feature/task-reordering-within-columns` is already pushed to remote
- Build verification passed successfully
- No breaking changes to existing functionality
- Ready for review and merge
