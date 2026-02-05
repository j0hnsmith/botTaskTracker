package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/starfederation/datastar-go/datastar"

	"github.com/j0hnsmith/botTaskTracker/ent"
	"github.com/j0hnsmith/botTaskTracker/ent/task"
	"github.com/j0hnsmith/botTaskTracker/ent/taskhistory"
	"github.com/j0hnsmith/botTaskTracker/ent/tasktag"
	"github.com/j0hnsmith/botTaskTracker/templates"
	"github.com/j0hnsmith/botTaskTracker/templates/fragments"
	"github.com/j0hnsmith/botTaskTracker/templates/pages"
)

// BoardViewHandler handles the main kanban board page.
func (s *Server) BoardViewHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get filter parameters
	selectedAssignee := r.URL.Query().Get("assignee")

	// Build query
	query := s.Client.Task.Query().
		WithTags().
		WithHistory(func(q *ent.TaskHistoryQuery) {
			q.Order(ent.Desc(taskhistory.FieldCreatedAt))
		}).
		Order(ent.Asc(task.FieldPosition), ent.Asc(task.FieldCreatedAt))

	// Apply assignee filter
	if selectedAssignee != "" && selectedAssignee != "all" {
		query = query.Where(task.AssigneeEQ(selectedAssignee))
	}

	// Get all tasks
	tasks, err := query.All(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get tasks", "error", err)
		http.Error(w, "Failed to load tasks", http.StatusInternalServerError)
		return
	}

	// Get activity feed
	activity, err := s.Client.TaskHistory.Query().
		WithTask().
		Order(ent.Desc(taskhistory.FieldCreatedAt)).
		Limit(30).
		All(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get activity", "error", err)
		http.Error(w, "Failed to load activity", http.StatusInternalServerError)
		return
	}

	// Get unique assignees
	assignees := []string{"peter", "john"}

	// Render page
	metaTags := pages.BoardMetaTags()
	bodyContent := pages.BoardContent(tasks, activity, assignees, selectedAssignee)
	boardTemplate := templates.Layout("Bot Task Tracker", metaTags, bodyContent)

	err = boardTemplate.Render(ctx, w)
	if err != nil {
		slog.ErrorContext(ctx, "render template", "error", err)
		http.Error(w, "Failed to render page", http.StatusInternalServerError)
		return
	}

	slog.InfoContext(ctx, "render page", "method", r.Method, "status", http.StatusOK, "path", r.URL.Path)
}

// TaskAddFormHandler returns an empty add task form via SSE.
func (s *Server) TaskAddFormHandler(w http.ResponseWriter, r *http.Request) {
	sse := datastar.NewSSE(w, r)

	// Render modal with empty form
	var htmlBuilder strings.Builder
	err := fragments.TaskAddModalWithForm().Render(r.Context(), &htmlBuilder)
	if err != nil {
		slog.ErrorContext(r.Context(), "failed to render add modal", "error", err)
		_ = sse.ConsoleError(err)
		return
	}

	// Insert modal and show it
	_ = sse.PatchElements(`<div id="modal-container">` + htmlBuilder.String() + `</div>`)

	// Set default signals
	signals := map[string]interface{}{
		"title":       "",
		"description": "",
		"column":      "backlog",
		"assignee":    "",
		"tags":        "",
	}
	signalsJSON, _ := json.Marshal(signals)
	_ = sse.PatchSignals(signalsJSON)

	_ = sse.ExecuteScript("document.getElementById('add-task-modal').showModal()")
}

// TaskCreateHandler creates a new task via SSE.
func (s *Server) TaskCreateHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Read signals BEFORE creating SSE
	type TaskCreateSignals struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Column      string `json:"column"`
		Assignee    string `json:"assignee"`
		Tags        string `json:"tags"`
	}
	signals := &TaskCreateSignals{}
	err := datastar.ReadSignals(r, signals)
	if err != nil {
		slog.ErrorContext(ctx, "failed to read signals", "error", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Create SSE AFTER reading signals
	sse := datastar.NewSSE(w, r)

	// Validate required fields
	if signals.Title == "" {
		_ = sse.PatchElements(`<div id="add-error" class="text-error text-sm">Title is required</div>`)
		return
	}

	// Sanitize column
	column := sanitizeColumn(signals.Column)

	// Get next position in column
	position, err := getNextPosition(ctx, s.Client, column)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get next position", "error", err)
		_ = sse.ConsoleError(err)
		return
	}

	// Create task
	newTask, err := s.Client.Task.Create().
		SetTitle(signals.Title).
		SetDescription(signals.Description).
		SetColumn(column).
		SetAssignee(signals.Assignee).
		SetPosition(position).
		Save(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "failed to create task", "error", err)
		_ = sse.ConsoleError(err)
		return
	}

	// Create history entry
	_, err = s.Client.TaskHistory.Create().
		SetTaskID(newTask.ID).
		SetAction("created").
		SetDetails(fmt.Sprintf("created in %s", column)).
		SetActor(signals.Assignee).
		Save(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "failed to create history", "error", err)
	}

	// Parse and add tags
	if signals.Tags != "" {
		tags := parseTags(signals.Tags)
		for _, tag := range tags {
			_, err = s.Client.TaskTag.Create().
				SetTaskID(newTask.ID).
				SetKey(tag.Key).
				SetValue(tag.Value).
				Save(ctx)
			if err != nil {
				slog.ErrorContext(ctx, "failed to create tag", "error", err)
			}
		}
	}

	// Reload task with edges
	newTask, err = s.Client.Task.Query().
		Where(task.IDEQ(newTask.ID)).
		WithTags().
		WithHistory().
		Only(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "failed to reload task", "error", err)
		_ = sse.ConsoleError(err)
		return
	}

	// Render task card
	var htmlBuilder strings.Builder
	err = fragments.TaskCard(newTask, column).Render(ctx, &htmlBuilder)
	if err != nil {
		slog.ErrorContext(ctx, "failed to render task card", "error", err)
		_ = sse.ConsoleError(err)
		return
	}

	// Clear error, append card to column, close modal
	_ = sse.PatchElements(`<div id="add-error" class="text-error text-sm hidden"></div>`)
	_ = sse.PatchElements(htmlBuilder.String(),
		datastar.WithModeAppend(),
		datastar.WithSelector("#column-"+column))
	_ = sse.PatchElements(`<div id="modal-container"></div>`)
}

// TaskEditFormHandler returns a populated edit form via SSE.
func (s *Server) TaskEditFormHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid task ID: "+err.Error(), http.StatusBadRequest)
		return
	}

	sse := datastar.NewSSE(w, r)

	// Get task from database
	t, err := s.Client.Task.Query().
		Where(task.IDEQ(id)).
		WithTags().
		WithHistory().
		Only(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get task", "error", err)
		_ = sse.ConsoleError(err)
		return
	}

	// Render modal with populated form
	var htmlBuilder strings.Builder
	err = fragments.TaskEditModalWithForm(t).Render(ctx, &htmlBuilder)
	if err != nil {
		slog.ErrorContext(ctx, "failed to render edit modal", "error", err)
		_ = sse.ConsoleError(err)
		return
	}

	// Insert modal and show it
	_ = sse.PatchElements(`<div id="modal-container">` + htmlBuilder.String() + `</div>`)

	// Set signals with existing values
	tagsStr := ""
	if len(t.Edges.Tags) > 0 {
		tagParts := make([]string, len(t.Edges.Tags))
		for i, tag := range t.Edges.Tags {
			tagParts[i] = fmt.Sprintf("%s:%s", tag.Key, tag.Value)
		}
		tagsStr = strings.Join(tagParts, ", ")
	}

	signals := map[string]interface{}{
		"task_id":     t.ID,
		"title":       t.Title,
		"description": t.Description,
		"column":      t.Column,
		"assignee":    t.Assignee,
		"tags":        tagsStr,
	}
	signalsJSON, _ := json.Marshal(signals)
	_ = sse.PatchSignals(signalsJSON)

	_ = sse.ExecuteScript("document.getElementById(\"edit-task-modal\").showModal()")
}

// TaskUpdateHandler updates an existing task via SSE.
func (s *Server) TaskUpdateHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Read signals BEFORE creating SSE
	type TaskUpdateSignals struct {
		TaskID      int    `json:"task_id"`
		Title       string `json:"title"`
		Description string `json:"description"`
		Column      string `json:"column"`
		Assignee    string `json:"assignee"`
		Tags        string `json:"tags"`
	}
	signals := &TaskUpdateSignals{}
	err := datastar.ReadSignals(r, signals)
	if err != nil {
		slog.ErrorContext(ctx, "failed to read signals", "error", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Create SSE AFTER reading signals
	sse := datastar.NewSSE(w, r)

	// Validate required fields
	if signals.Title == "" {
		_ = sse.PatchElements(`<div id="edit-error" class="text-error text-sm">Title is required</div>`)
		return
	}

	// Check if task exists
	existingTask, err := s.Client.Task.Query().
		Where(task.IDEQ(signals.TaskID)).
		Only(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find task for update", "error", err)
		_ = sse.ConsoleError(err)
		return
	}

	// Update task
	updatedTask, err := s.Client.Task.UpdateOneID(existingTask.ID).
		SetTitle(signals.Title).
		SetDescription(signals.Description).
		SetColumn(sanitizeColumn(signals.Column)).
		SetAssignee(signals.Assignee).
		Save(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "failed to update task", "error", err)
		_ = sse.ConsoleError(err)
		return
	}

	// Create history entry
	_, _ = s.Client.TaskHistory.Create().
		SetTaskID(updatedTask.ID).
		SetAction("updated").
		SetDetails("updated task").
		SetActor(signals.Assignee).
		Save(ctx)

	// Update tags
	_, _ = s.Client.TaskTag.Delete().Where(tasktag.HasTaskWith(task.IDEQ(updatedTask.ID))).Exec(ctx)
	if signals.Tags != "" {
		tags := parseTags(signals.Tags)
		for _, tag := range tags {
			_, _ = s.Client.TaskTag.Create().
				SetTaskID(updatedTask.ID).
				SetKey(tag.Key).
				SetValue(tag.Value).
				Save(ctx)
		}
	}

	// Reload task
	updatedTask, err = s.Client.Task.Query().
		Where(task.IDEQ(updatedTask.ID)).
		WithTags().
		WithHistory().
		Only(ctx)
	if err != nil {
		_ = sse.ConsoleError(err)
		return
	}

	// Render updated card
	var htmlBuilder strings.Builder
	err = fragments.TaskCard(updatedTask, updatedTask.Column).Render(ctx, &htmlBuilder)
	if err != nil {
		_ = sse.ConsoleError(err)
		return
	}

	_ = sse.PatchElements(`<div id="edit-error" class="text-error text-sm hidden"></div>`)
	_ = sse.PatchElements(htmlBuilder.String())
	_ = sse.PatchElements(`<div id="modal-container"></div>`)
}

// TaskDeleteHandler deletes a task via SSE.
func (s *Server) TaskDeleteHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	sse := datastar.NewSSE(w, r)

	// Delete related records
	_, _ = s.Client.TaskHistory.Delete().Where(taskhistory.HasTaskWith(task.IDEQ(id))).Exec(ctx)
	_, _ = s.Client.TaskTag.Delete().Where(tasktag.HasTaskWith(task.IDEQ(id))).Exec(ctx)
	err = s.Client.Task.DeleteOneID(id).Exec(ctx)
	if err != nil {
		_ = sse.ConsoleError(err)
		return
	}

	_ = sse.RemoveElement("#task-card-" + strconv.Itoa(id))
}

// Stub handlers for move, assign, tag operations
func (s *Server) TaskMoveHandler(w http.ResponseWriter, r *http.Request)      {}
func (s *Server) TaskAssignHandler(w http.ResponseWriter, r *http.Request)    {}
func (s *Server) TaskAddTagHandler(w http.ResponseWriter, r *http.Request)    {}
func (s *Server) TaskRemoveTagHandler(w http.ResponseWriter, r *http.Request) {}

// Helper functions
func sanitizeColumn(value string) string {
	clean := strings.ToLower(strings.TrimSpace(value))
	switch clean {
	case "backlog", "in_progress", "review", "done":
		return clean
	default:
		return "backlog"
	}
}

func getNextPosition(ctx context.Context, client *ent.Client, column string) (int, error) {
	max, err := client.Task.Query().
		Where(task.ColumnEQ(column)).
		Aggregate(ent.Max(task.FieldPosition)).
		Int(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return 0, nil
		}
		return 0, err
	}
	return max + 1, nil
}

type tagInput struct {
	Key   string
	Value string
}

func parseTags(input string) []tagInput {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil
	}
	parts := strings.Split(input, ",")
	results := make([]tagInput, 0, len(parts))
	for _, raw := range parts {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}
		kv := strings.SplitN(raw, ":", 2)
		if len(kv) != 2 {
			continue
		}
		key := strings.TrimSpace(kv[0])
		value := strings.TrimSpace(kv[1])
		if key == "" || value == "" {
			continue
		}
		results = append(results, tagInput{Key: key, Value: value})
	}
	return results
}
