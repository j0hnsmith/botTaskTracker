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

// UnifiedEvent represents any event that can be broadcast (board or activity)
type UnifiedEvent struct {
	EventType string // "board" or "activity"
	Type      string // "task_created", "task_updated", "task_moved", "task_deleted", "activity_created"
	TaskID    int
	HistoryID int
	Column    string
	Nonce     string // Client nonce to prevent echo-back
}

// Broadcaster manages SSE connections and broadcasts unified events
type Broadcaster struct {
	mu      sync.RWMutex
	clients map[chan UnifiedEvent]bool
}

// NewBroadcaster creates a new broadcaster
func NewBroadcaster() *Broadcaster {
	return &Broadcaster{
		clients: make(map[chan UnifiedEvent]bool),
	}
}

// Register adds a new client to the broadcaster
func (b *Broadcaster) Register(client chan UnifiedEvent) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.clients[client] = true
}

// Unregister removes a client from the broadcaster
func (b *Broadcaster) Unregister(client chan UnifiedEvent) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if _, ok := b.clients[client]; ok {
		delete(b.clients, client)
		close(client)
	}
}

// BroadcastBoard sends a board event to all connected clients
func (b *Broadcaster) BroadcastBoard(taskID int, eventType, column, nonce string) {
	event := UnifiedEvent{
		EventType: "board",
		Type:      eventType,
		TaskID:    taskID,
		Column:    column,
		Nonce:     nonce,
	}
	b.broadcast(event)
}

// BroadcastActivity sends an activity event to all connected clients
func (b *Broadcaster) BroadcastActivity(historyID int) {
	event := UnifiedEvent{
		EventType: "activity",
		Type:      "activity_created",
		HistoryID: historyID,
	}
	b.broadcast(event)
}

// broadcast sends a unified event to all connected clients
func (b *Broadcaster) broadcast(event UnifiedEvent) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	
	clientCount := len(b.clients)
	sent := 0
	skipped := 0
	
	for client := range b.clients {
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
		println("Broadcast:", event.EventType, event.Type, "taskID:", event.TaskID, "historyID:", event.HistoryID, "clients:", clientCount, "sent:", sent, "skipped:", skipped)
	}
}

// HandleEvents handles SSE connections for unified events (board + activity)
func (s *Server) HandleEvents(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// Create SSE writer
	sse := datastar.NewSSE(w, r)
	
	// Create event channel for this client
	eventChan := make(chan UnifiedEvent, 10)
	
	// Register client
	s.Broadcaster.Register(eventChan)
	defer s.Broadcaster.Unregister(eventChan)
	
	// Send initial connection message
	_ = sse.PatchSignals([]byte(`{"sseConnected": true}`))
	
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
			// Route event based on type
			var err error
			if event.EventType == "board" {
				err = s.handleBoardEvent(ctx, sse, event)
			} else if event.EventType == "activity" {
				err = s.handleActivityEvent(ctx, sse, event)
			}
			
			if err != nil {
				println("handleEvent error:", err.Error())
				// Continue the loop - don't close the SSE connection
			}
		}
	}
}

// handleBoardEvent processes a single board event and sends updates via SSE
func (s *Server) handleBoardEvent(ctx context.Context, sse *datastar.ServerSentEventGenerator, event UnifiedEvent) error {
	println("handleBoardEvent:", event.Type, "taskID:", event.TaskID, "nonce:", event.Nonce)
	
	// Send nonce signal before each patch so frontend can check
	if event.Nonce != "" {
		_ = sse.PatchSignals([]byte(`{"lastEventNonce": "` + event.Nonce + `"}`))
	}
	
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

// handleActivityEvent processes a single activity event and sends updates via SSE
func (s *Server) handleActivityEvent(ctx context.Context, sse *datastar.ServerSentEventGenerator, event UnifiedEvent) error {
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
