# Pull Request: Add drag-to-reorder functionality within columns

## Branch
- **Source:** feature/task-reordering-within-columns
- **Target:** main
- **Repository:** https://github.com/j0hnsmith/botTaskTracker

## Overview
This PR adds the ability to reorder tasks within the same column by dragging them up or down, setting their priority order.

## Current State
- Tasks can be dragged between columns ✅
- No way to reorder tasks within the same column ❌

## Changes Made

### 1. Frontend (drag-drop.js)
- Detect position changes within the same column
- Send position updates to new endpoint when tasks are reordered
- Keep existing drag-between-columns functionality working

### 2. Backend (handlers/tasks.go)
- Added new endpoint `PATCH /datastar/tasks/{id}/position` for position updates
- Added `reorderTasksInColumn()` helper function for within-column reordering
- Updated `reorderTasksOnColumnChange()` to handle cross-column moves with positions
- Added `renderColumnUpdate()` to efficiently re-render entire columns
- Added `recompactColumnPositions()` to ensure sequential positions

### 3. Routes (handlers/server.go)
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
- [ ] Manual testing required for UI/UX

## How to Create the PR
Since `gh` CLI is not available, please create the PR manually:

1. Visit: https://github.com/j0hnsmith/botTaskTracker/pull/new/feature/task-reordering-within-columns
2. Set base branch to `main`
3. Copy this description to the PR body
4. Submit the PR

Alternatively, if you have GitHub CLI installed elsewhere:
```bash
gh pr create --repo j0hnsmith/botTaskTracker \
  --head feature/task-reordering-within-columns \
  --base main \
  --title "Add drag-to-reorder functionality within columns" \
  --body-file PR_DETAILS.md
```

## Related
Implements the requested feature for drag-to-reorder functionality within botTaskTracker columns.
