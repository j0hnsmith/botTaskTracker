package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Task holds the schema definition for the Task entity.
type Task struct {
	ent.Schema
}

// Fields of the Task.
func (Task) Fields() []ent.Field {
	return []ent.Field{
		field.String("title").
			NotEmpty(),
		field.Text("description").
			Optional(),
		field.String("column").
			Default("backlog"), // backlog, in_progress, review, done
		field.String("assignee").
			Default(""), // empty, "peter", "john"
		field.Int("position").
			Default(0), // for ordering within column
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

// Edges of the Task.
func (Task) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("tags", TaskTag.Type),
		edge.To("history", TaskHistory.Type),
	}
}
