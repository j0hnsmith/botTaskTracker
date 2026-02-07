package handlers

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/j0hnsmith/botTaskTracker/ent/task"
	"github.com/j0hnsmith/botTaskTracker/templates/fragments"
	"github.com/starfederation/datastar-go/datastar"
)

// BoardEvent represents a change to the board state
type BoardEvent struct {
	Type   string // "task_created", "task_updated", "task_moved", "task_deleted"
	TaskID int
	Column string
}

// ActivityEvent represents a change to the activity stream
type ActivityEvent struct {
	Type      string // "activity_created"
	HistoryID int
}

// Broadcaster manages SSE connections and broadcasts board events
type Broadcaster struct {
	mu              sync.RWMutex
	boardClients    map[chan BoardEvent]bool
	activityClients map[chan ActivityEvent]bool
}

// NewBroadcaster creates a new broadcaster
func NewBroadcaster() *Broadcaster {
	return &Broadcaster{
		boardClients:    make(map[chan BoardEvent]bool),
		activityClients: make(map[chan ActivityEvent]bool),
	}
}

// Register adds a new board client to the broadcaster
func (b *Broadcaster) Register(client chan BoardEvent) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.boardClients[client] = true
}

// Unregister removes a board client from the broadcaster
func (b *Broadcaster) Unregister(client chan BoardEvent) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if _, ok := b.boardClients[client]; ok {
		delete(b.boardClients, client)
		close(client)
	}
}

// RegisterActivity adds a new activity client to the broadcaster
func (b *Broadcaster) RegisterActivity(client chan ActivityEvent) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.activityClients[client] = true
}

// UnregisterActivity removes an activity client from the broadcaster
func (b *Broadcaster) UnregisterActivity(client chan ActivityEvent) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if _, ok := b.activityClients[client]; ok {
		delete(b.activityClients, client)
		close(client)
	}
}

// Broadcast sends a board event to all connected board clients
func (b *Broadcaster) Broadcast(event BoardEvent) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	
	clientCount := len(b.boardClients)
	sent := 0
	skipped := 0
	
	for client := range b.boardClients {
		select {
		case client <- event:
			sent++
		default:
			// Client channel is full, skip
			skipped++
		}
	}
	
	// Log broadcast stats (will appear in systemd journal)
	if clientCount > 0 {
		println("Broadcast:", event.Type, "taskID:", event.TaskID, "clients:", clientCount, "sent:", sent, "skipped:", skipped)
	}
}

// BroadcastActivity sends an activity event to all connected activity clients
func (b *Broadcaster) BroadcastActivity(event ActivityEvent) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	
	clientCount := len(b.activityClients)
	sent := 0
	skipped := 0
	
	for client := range b.activityClients {
		select {
		case client <- event:
			sent++
		default:
			// Client channel is full, skip
			skipped++
		}
	}
	
	// Log broadcast stats (will appear in systemd journal)
	if clientCount > 0 {
		println("BroadcastActivity:", event.Type, "historyID:", event.HistoryID, "clients:", clientCount, "sent:", sent, "skipped:", skipped)
	}
}

// HandleBoardEvents handles SSE connections for board events
func (s *Server) HandleBoardEvents(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// Create SSE writer
	sse := datastar.NewSSE(w, r)
	
	// Create event channel for this client
	eventChan := make(chan BoardEvent, 10)
	
	// Register client
	s.Broadcaster.Register(eventChan)
	defer s.Broadcaster.Unregister(eventChan)
	
	// Send initial connection message
	_ = sse.PatchSignals([]byte(`{"boardConnected": true}`))
	
	// Keepalive ticker to prevent connection timeouts
	keepalive := time.NewTicker(30 * time.Second)
	defer keepalive.Stop()
	
	// Listen for events or context cancellation
	for {
		select {
		case <-ctx.Done():
			return
		case <-keepalive.C:
			// Send keepalive comment to prevent timeout
			w.Write([]byte(": keepalive\n\n"))
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
		case event := <-eventChan:
			// Handle the event based on type
			// Don't return on error - log it and continue streaming
			if err := s.handleBoardEvent(ctx, sse, event); err != nil {
				println("handleBoardEvent error:", err.Error())
				// Continue the loop - don't close the SSE connection
			}
		}
	}
}

// handleBoardEvent processes a single board event and sends updates via SSE
func (s *Server) handleBoardEvent(ctx context.Context, sse *datastar.ServerSentEventGenerator, event BoardEvent) error {
	println("handleBoardEvent:", event.Type, "taskID:", event.TaskID)
	switch event.Type {
	case "task_created":
		// Load task with edges
		t, err := s.Client.Task.Query().
			Where(task.IDEQ(event.TaskID)).
			WithTags().
			WithHistory().
			Only(ctx)
		if err != nil {
			return err
		}
		
		// Render the task card
		var htmlBuilder strings.Builder
		err = fragments.TaskCard(t, t.Column).Render(ctx, &htmlBuilder)
		if err != nil {
			return err
		}
		
		// Append to column
		_ = sse.PatchElements(htmlBuilder.String(),
			datastar.WithModeAppend(),
			datastar.WithSelector("#column-"+t.Column))
		
	case "task_updated":
		// Load task with edges
		t, err := s.Client.Task.Query().
			Where(task.IDEQ(event.TaskID)).
			WithTags().
			WithHistory().
			Only(ctx)
		if err != nil {
			return err
		}
		
		// Render the task card
		var htmlBuilder strings.Builder
		err = fragments.TaskCard(t, t.Column).Render(ctx, &htmlBuilder)
		if err != nil {
			return err
		}
		
		// Replace existing card in place
		_ = sse.PatchElements(htmlBuilder.String())
		
	case "task_moved":
		// Load task with edges
		t, err := s.Client.Task.Query().
			Where(task.IDEQ(event.TaskID)).
			WithTags().
			WithHistory().
			Only(ctx)
		if err != nil {
			return err
		}
		
		// Render the task card
		var htmlBuilder strings.Builder
		err = fragments.TaskCard(t, t.Column).Render(ctx, &htmlBuilder)
		if err != nil {
			return err
		}
		
		// Remove from old location and append to new column
		_ = sse.RemoveElement("#task-card-" + strconv.Itoa(event.TaskID))
		_ = sse.PatchElements(htmlBuilder.String(),
			datastar.WithModeAppend(),
			datastar.WithSelector("#column-"+t.Column))
		
	case "task_deleted":
		// Remove the element
		_ = sse.RemoveElement("#task-card-" + strconv.Itoa(event.TaskID))
	}
	
	return nil
}

// HandleActivityEvents handles SSE connections for activity stream events
func (s *Server) HandleActivityEvents(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// Create SSE writer
	sse := datastar.NewSSE(w, r)
	
	// Create event channel for this client
	eventChan := make(chan ActivityEvent, 10)
	
	// Register client
	s.Broadcaster.RegisterActivity(eventChan)
	defer s.Broadcaster.UnregisterActivity(eventChan)
	
	// Send initial connection message
	_ = sse.PatchSignals([]byte(`{"activityConnected": true}`))
	
	// Keepalive ticker to prevent connection timeouts
	keepalive := time.NewTicker(30 * time.Second)
	defer keepalive.Stop()
	
	// Listen for events or context cancellation
	for {
		select {
		case <-ctx.Done():
			return
		case <-keepalive.C:
			// Send keepalive comment to prevent timeout
			w.Write([]byte(": keepalive\n\n"))
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
		case event := <-eventChan:
			// Handle the event
			if err := s.handleActivityEvent(ctx, sse, event); err != nil {
				println("handleActivityEvent error:", err.Error())
				// Continue the loop - don't close the SSE connection
			}
		}
	}
}

// handleActivityEvent processes a single activity event and sends updates via SSE
func (s *Server) handleActivityEvent(ctx context.Context, sse *datastar.ServerSentEventGenerator, event ActivityEvent) error {
	println("handleActivityEvent:", event.Type, "historyID:", event.HistoryID)
	
	switch event.Type {
	case "activity_created":
		// Load the history entry with task edge
		history, err := s.Client.TaskHistory.Query().
			Where(func(selector *sql.Selector) {
				selector.Where(sql.EQ("id", event.HistoryID))
			}).
			WithTask().
			Only(ctx)
		if err != nil {
			return err
		}
		
		// Render single activity item
		var htmlBuilder strings.Builder
		err = fragments.ActivityItem(history, 0).Render(ctx, &htmlBuilder)
		if err != nil {
			return err
		}
		
		// Prepend to activity timeline
		_ = sse.PatchElements(htmlBuilder.String(),
			datastar.WithModePrepend(),
			datastar.WithSelector("#activity-timeline"))
		
		// Execute script to maintain max 30 entries
		_ = sse.ExecuteScript(`
			const timeline = document.getElementById('activity-timeline');
			if (timeline) {
				const items = timeline.querySelectorAll('li');
				if (items.length > 30) {
					// Remove items beyond 30
					for (let i = 30; i < items.length; i++) {
						items[i].remove();
					}
				}
			}
		`)
	}
	
	return nil
}
