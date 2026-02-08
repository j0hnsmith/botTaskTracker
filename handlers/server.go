package handlers

import (
	"context"
	"database/sql"
	"io/fs"
	"log/slog"
	"net/http"
	"os"

	entsql "entgo.io/ent/dialect/sql"
	"github.com/j0hnsmith/botTaskTracker/ent"
	sqlite "modernc.org/sqlite"
)

type Server struct {
	Client      *ent.Client
	Broadcaster *Broadcaster
}

func NewServer(ctx context.Context) (*Server, error) {
	if err := os.MkdirAll("data", 0o755); err != nil {
		return nil, err
	}

	sql.Register("sqlite3", &sqlite.Driver{})
	drv, err := entsql.Open("sqlite3", "file:data/bot_task_tracker.db")
	if err != nil {
		return nil, err
	}

	if _, err := drv.DB().ExecContext(ctx, "PRAGMA foreign_keys = ON"); err != nil {
		return nil, err
	}

	client := ent.NewClient(ent.Driver(drv))

	if err := client.Schema.Create(ctx); err != nil {
		return nil, err
	}

	return &Server{
		Client:      client,
		Broadcaster: NewBroadcaster(),
	}, nil
}

func (s *Server) Close() error {
	return s.Client.Close()
}

func (s *Server) Routes(staticFS fs.FS) *http.ServeMux {
	mux := http.NewServeMux()

	// Static assets from embedded FS
	staticSub, _ := fs.Sub(staticFS, "static")
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticSub))))

	// Page routes
	mux.HandleFunc("GET /{$}", s.BoardViewHandler)

	// Column content endpoint (for drag-drop refresh)
	mux.HandleFunc("GET /columns/{column}", s.ColumnContentHandler)

	// SSE endpoint for unified real-time updates (board + activity)
	mux.HandleFunc("GET /datastar/events", s.HandleEvents)

	// Datastar SSE routes for tasks
	mux.HandleFunc("GET /datastar/tasks/add-form", s.TaskAddFormHandler)
	mux.HandleFunc("POST /datastar/tasks", s.TaskCreateHandler)
	mux.HandleFunc("GET /datastar/tasks/details/{id}", s.TaskDetailsHandler)
	mux.HandleFunc("GET /datastar/tasks/edit/{id}", s.TaskEditFormHandler)
	mux.HandleFunc("PUT /datastar/tasks/{id}", s.TaskUpdateHandler)
	mux.HandleFunc("PATCH /datastar/tasks/{id}/column", s.TaskColumnUpdateHandler)
	mux.HandleFunc("PATCH /datastar/tasks/{id}/position", s.TaskPositionUpdateHandler)
	mux.HandleFunc("DELETE /datastar/tasks/{id}", s.TaskDeleteHandler)
	mux.HandleFunc("POST /datastar/tasks/{id}/move", s.TaskMoveHandler)
	mux.HandleFunc("POST /datastar/tasks/{id}/assign", s.TaskAssignHandler)
	mux.HandleFunc("POST /datastar/tasks/{id}/tag", s.TaskAddTagHandler)
	mux.HandleFunc("DELETE /datastar/tasks/{id}/tags/{tagId}", s.TaskRemoveTagHandler)

	return mux
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		slog.InfoContext(r.Context(), "request", "method", r.Method, "path", r.URL.Path)
		next.ServeHTTP(w, r)
	})
}
