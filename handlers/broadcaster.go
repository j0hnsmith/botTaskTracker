package handlers

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"sync"

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

// Broadcaster manages SSE connections and broadcasts board events
type Broadcaster struct {
	mu      sync.RWMutex
	clients map[chan BoardEvent]bool
}

// NewBroadcaster creates a new broadcaster
func NewBroadcaster() *Broadcaster {
	return &Broadcaster{
		clients: make(map[chan BoardEvent]bool),
	}
}

// Register adds a new client to the broadcaster
func (b *Broadcaster) Register(client chan BoardEvent) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.clients[client] = true
}

// Unregister removes a client from the broadcaster
func (b *Broadcaster) Unregister(client chan BoardEvent) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if _, ok := b.clients[client]; ok {
		delete(b.clients, client)
		close(client)
	}
}

// Broadcast sends an event to all connected clients
func (b *Broadcaster) Broadcast(event BoardEvent) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for client := range b.clients {
		select {
		case client <- event:
		default:
			// Client channel is full, skip
		}
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
	
	// Listen for events or context cancellation
	for {
		select {
		case <-ctx.Done():
			return
		case event := <-eventChan:
			// Handle the event based on type
			if err := s.handleBoardEvent(ctx, sse, event); err != nil {
				return
			}
		}
	}
}

// handleBoardEvent processes a single board event and sends updates via SSE
func (s *Server) handleBoardEvent(ctx context.Context, sse *datastar.ServerSentEventGenerator, event BoardEvent) error {
	switch event.Type {
	case "task_created", "task_updated", "task_moved":
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
		
		// Send the update
		if event.Type == "task_created" {
			// Append to column
			_ = sse.PatchElements(htmlBuilder.String(),
				datastar.WithModeAppend(),
				datastar.WithSelector("#column-"+t.Column))
		} else {
			// Replace existing card
			_ = sse.PatchElements(htmlBuilder.String())
		}
		
	case "task_deleted":
		// Remove the element
		_ = sse.RemoveElement("#task-card-" + strconv.Itoa(event.TaskID))
	}
	
	return nil
}
