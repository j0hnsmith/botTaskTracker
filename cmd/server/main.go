package main

import (
	"context"
	"embed"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/j0hnsmith/botTaskTracker/ent"
	"github.com/j0hnsmith/botTaskTracker/ent/task"
	"github.com/j0hnsmith/botTaskTracker/ent/taskhistory"
	"github.com/j0hnsmith/botTaskTracker/ent/tasktag"

	_ "modernc.org/sqlite"
)

//go:embed templates/*.html
var templatesFS embed.FS

type columnConfig struct {
	Key   string
	Title string
}

type columnView struct {
	Key   string
	Title string
	Tasks []*ent.Task
}

type boardPage struct {
	Columns          []columnView
	Assignees        []string
	SelectedAssignee string
	Activity         []*ent.TaskHistory
	Now              time.Time
}

type server struct {
	client *ent.Client
	tmpl   *template.Template
}

func main() {
	ctx := context.Background()

	if err := os.MkdirAll("data", 0o755); err != nil {
		log.Fatalf("create data dir: %v", err)
	}

	client, err := ent.Open("sqlite", "file:data/bot_task_tracker.db?_fk=1")
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer func() {
		if cerr := client.Close(); cerr != nil {
			log.Printf("close db: %v", cerr)
		}
	}()

	if err := client.Schema.Create(ctx); err != nil {
		log.Fatalf("migrate db: %v", err)
	}

	if err := seedData(ctx, client); err != nil {
		log.Fatalf("seed data: %v", err)
	}

	funcs := template.FuncMap{
		"formatTime": func(t time.Time) string {
			return t.Format("Jan 2, 15:04")
		},
		"badgeStyle": func(key string) string {
			switch strings.ToLower(key) {
			case "priority":
				return "badge-error"
			case "project":
				return "badge-primary"
			case "type":
				return "badge-secondary"
			default:
				return "badge-ghost"
			}
		},
	}

	tmpl := template.Must(template.New("layout.html").Funcs(funcs).ParseFS(templatesFS, "templates/*.html"))

	srv := &server{client: client, tmpl: tmpl}

	mux := http.NewServeMux()
	mux.HandleFunc("/", srv.handleBoard)
	mux.HandleFunc("/tasks", srv.handleCreateTask)
	mux.HandleFunc("/tasks/", srv.handleTaskAction)

	log.Printf("botTaskTracker running at http://localhost:7002")
	if err := http.ListenAndServe(":7002", loggingMiddleware(mux)); err != nil {
		log.Fatalf("serve: %v", err)
	}
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}

func (s *server) handleBoard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()
	selectedAssignee := strings.TrimSpace(r.URL.Query().Get("assignee"))
	if selectedAssignee == "" {
		selectedAssignee = "all"
	}

	query := s.client.Task.Query().WithTags().WithHistory(func(q *ent.TaskHistoryQuery) {
		q.Order(ent.Desc(taskhistory.FieldCreatedAt))
	})
	if strings.ToLower(selectedAssignee) != "all" {
		query = query.Where(task.AssigneeEQ(selectedAssignee))
	}

	tasks, err := query.Order(ent.Asc(task.FieldPosition), ent.Asc(task.FieldCreatedAt)).All(ctx)
	if err != nil {
		http.Error(w, fmt.Sprintf("load tasks: %v", err), http.StatusInternalServerError)
		return
	}

	columns := buildColumns(tasks)

	assignees, err := s.listAssignees(ctx)
	if err != nil {
		http.Error(w, fmt.Sprintf("load assignees: %v", err), http.StatusInternalServerError)
		return
	}

	activity, err := s.client.TaskHistory.Query().
		WithTask().
		Order(ent.Desc(taskhistory.FieldCreatedAt)).
		Limit(30).
		All(ctx)
	if err != nil {
		http.Error(w, fmt.Sprintf("load activity: %v", err), http.StatusInternalServerError)
		return
	}

	page := boardPage{
		Columns:          columns,
		Assignees:        assignees,
		SelectedAssignee: selectedAssignee,
		Activity:         activity,
		Now:              time.Now(),
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.tmpl.ExecuteTemplate(w, "layout", page); err != nil {
		log.Printf("render template: %v", err)
	}
}

func (s *server) handleCreateTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}

	title := strings.TrimSpace(r.FormValue("title"))
	if title == "" {
		http.Error(w, "title required", http.StatusBadRequest)
		return
	}

	description := strings.TrimSpace(r.FormValue("description"))
	column := sanitizeColumn(r.FormValue("column"))
	assignee := strings.TrimSpace(r.FormValue("assignee"))
	position := nextPosition(ctx, s.client, column)

	newTask, err := s.client.Task.Create().
		SetTitle(title).
		SetDescription(description).
		SetColumn(column).
		SetAssignee(assignee).
		SetPosition(position).
		Save(ctx)
	if err != nil {
		http.Error(w, fmt.Sprintf("create task: %v", err), http.StatusInternalServerError)
		return
	}

	_, _ = s.client.TaskHistory.Create().
		SetTaskID(newTask.ID).
		SetAction("created").
		SetDetails(fmt.Sprintf("created in %s", column)).
		SetActor(assignee).
		Save(ctx)

	if tags := parseTags(r.FormValue("tags")); len(tags) > 0 {
		for _, tag := range tags {
			_, _ = s.client.TaskTag.Create().
				SetTaskID(newTask.ID).
				SetKey(tag.Key).
				SetValue(tag.Value).
				Save(ctx)
		}
	}

	redirectBack(w, r)
}

func (s *server) handleTaskAction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 3 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(parts[1])
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	action := parts[2]
	ctx := r.Context()
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}

	switch action {
	case "move":
		column := sanitizeColumn(r.FormValue("column"))
		if err := s.client.Task.UpdateOneID(id).SetColumn(column).Save(ctx); err != nil {
			http.Error(w, "unable to move", http.StatusInternalServerError)
			return
		}
		_, _ = s.client.TaskHistory.Create().
			SetTaskID(id).
			SetAction("moved").
			SetDetails(fmt.Sprintf("moved to %s", column)).
			Save(ctx)
	case "assign":
		assignee := strings.TrimSpace(r.FormValue("assignee"))
		if err := s.client.Task.UpdateOneID(id).SetAssignee(assignee).Save(ctx); err != nil {
			http.Error(w, "unable to assign", http.StatusInternalServerError)
			return
		}
		_, _ = s.client.TaskHistory.Create().
			SetTaskID(id).
			SetAction("assigned").
			SetDetails(fmt.Sprintf("assigned to %s", assignee)).
			SetActor(assignee).
			Save(ctx)
	case "tag":
		key := strings.TrimSpace(r.FormValue("key"))
		value := strings.TrimSpace(r.FormValue("value"))
		if key == "" || value == "" {
			http.Error(w, "key and value required", http.StatusBadRequest)
			return
		}
		if _, err := s.client.TaskTag.Create().SetTaskID(id).SetKey(key).SetValue(value).Save(ctx); err != nil {
			http.Error(w, "unable to add tag", http.StatusInternalServerError)
			return
		}
		_, _ = s.client.TaskHistory.Create().
			SetTaskID(id).
			SetAction("tagged").
			SetDetails(fmt.Sprintf("tagged %s:%s", key, value)).
			Save(ctx)
	case "delete":
		if _, err := s.client.TaskHistory.Delete().Where(taskhistory.HasTaskWith(task.ID(id))).Exec(ctx); err != nil {
			http.Error(w, "unable to delete history", http.StatusInternalServerError)
			return
		}
		if _, err := s.client.TaskTag.Delete().Where(tasktag.HasTaskWith(task.ID(id))).Exec(ctx); err != nil {
			http.Error(w, "unable to delete tags", http.StatusInternalServerError)
			return
		}
		if err := s.client.Task.DeleteOneID(id).Exec(ctx); err != nil {
			http.Error(w, "unable to delete task", http.StatusInternalServerError)
			return
		}
	default:
		http.Error(w, "unknown action", http.StatusBadRequest)
		return
	}

	redirectBack(w, r)
}

func redirectBack(w http.ResponseWriter, r *http.Request) {
	location := r.Header.Get("Referer")
	if location == "" {
		location = "/"
	}
	http.Redirect(w, r, location, http.StatusSeeOther)
}

func buildColumns(tasks []*ent.Task) []columnView {
	configs := []columnConfig{{Key: "backlog", Title: "Backlog"}, {Key: "in_progress", Title: "In Progress"}, {Key: "review", Title: "Review"}, {Key: "done", Title: "Done"}}
	columns := make([]columnView, 0, len(configs))
	byColumn := make(map[string][]*ent.Task)
	for _, taskItem := range tasks {
		byColumn[taskItem.Column] = append(byColumn[taskItem.Column], taskItem)
	}
	for _, cfg := range configs {
		columns = append(columns, columnView{Key: cfg.Key, Title: cfg.Title, Tasks: byColumn[cfg.Key]})
	}
	return columns
}

func (s *server) listAssignees(ctx context.Context) ([]string, error) {
	list, err := s.client.Task.Query().Select(task.FieldAssignee).Strings(ctx)
	if err != nil {
		return nil, err
	}

	seen := make(map[string]struct{})
	assignees := make([]string, 0, len(list))
	for _, item := range list {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		if _, exists := seen[item]; exists {
			continue
		}
		seen[item] = struct{}{}
		assignees = append(assignees, item)
	}

	sort.Strings(assignees)
	return assignees, nil
}

func sanitizeColumn(value string) string {
	clean := strings.ToLower(strings.TrimSpace(value))
	switch clean {
	case "backlog", "in_progress", "review", "done":
		return clean
	default:
		return "backlog"
	}
}

func nextPosition(ctx context.Context, client *ent.Client, column string) int {
	max, err := client.Task.Query().Where(task.ColumnEQ(column)).Max(ctx, task.FieldPosition)
	if err != nil {
		if ent.IsNotFound(err) {
			return 0
		}
		return 0
	}
	return max + 1
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

func seedData(ctx context.Context, client *ent.Client) error {
	count, err := client.Task.Query().Count(ctx)
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	seedTasks := []struct {
		Title       string
		Description string
		Column      string
		Assignee    string
		Tags        []tagInput
		History     []string
	}{
		{
			Title:       "Initialize ent schema",
			Description: "Set up Task, TaskTag, TaskHistory schemas and run migrations.",
			Column:      "backlog",
			Assignee:    "", 
			Tags: []tagInput{{Key: "project", Value: "foundation"}, {Key: "priority", Value: "high"}},
			History:     []string{"created"},
		},
		{
			Title:       "Build server scaffold",
			Description: "Wire up HTTP server, routing, templates, and data access.",
			Column:      "in_progress",
			Assignee:    "peter",
			Tags:        []tagInput{{Key: "type", Value: "feature"}},
			History:     []string{"created", "assigned"},
		},
		{
			Title:       "Ship first UI pass",
			Description: "Design board layout with Tailwind/daisyUI and Datastar hooks.",
			Column:      "review",
			Assignee:    "john",
			Tags:        []tagInput{{Key: "priority", Value: "medium"}, {Key: "readyToStart", Value: "true"}},
			History:     []string{"created", "moved"},
		},
		{
			Title:       "Announce deployment",
			Description: "Broadcast the new botTaskTracker status page.",
			Column:      "done",
			Assignee:    "jane",
			Tags:        []tagInput{{Key: "type", Value: "ops"}},
			History:     []string{"created", "moved", "assigned"},
		},
	}

	positionByColumn := make(map[string]int)
	for _, seed := range seedTasks {
		positionByColumn[seed.Column]++
		createdTask, err := client.Task.Create().
			SetTitle(seed.Title).
			SetDescription(seed.Description).
			SetColumn(seed.Column).
			SetAssignee(seed.Assignee).
			SetPosition(positionByColumn[seed.Column]).
			Save(ctx)
		if err != nil {
			return err
		}

		for _, tag := range seed.Tags {
			if _, err := client.TaskTag.Create().SetTaskID(createdTask.ID).SetKey(tag.Key).SetValue(tag.Value).Save(ctx); err != nil {
				return err
			}
		}

		for _, action := range seed.History {
			details := action
			if action == "moved" {
				details = fmt.Sprintf("moved to %s", seed.Column)
			}
			if action == "assigned" {
				details = fmt.Sprintf("assigned to %s", seed.Assignee)
			}
			if _, err := client.TaskHistory.Create().
				SetTaskID(createdTask.ID).
				SetAction(action).
				SetDetails(details).
				SetActor(seed.Assignee).
				Save(ctx); err != nil {
				return err
			}
		}
	}

	return nil
}
