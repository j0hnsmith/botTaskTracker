package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// TaskHistory holds the schema definition for the TaskHistory entity.
type TaskHistory struct {
	ent.Schema
}

// Fields of the TaskHistory.
func (TaskHistory) Fields() []ent.Field {
	return []ent.Field{
		field.String("action").
			NotEmpty(), // "created", "moved", "updated", "deleted", "tagged", "assigned"
		field.String("details").
			Optional(), // e.g., "moved from backlog to in_progress", "assigned to john"
		field.String("actor").
			Default(""), // who made the change: "peter", "john"
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
	}
}

// Edges of the TaskHistory.
func (TaskHistory) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("task", Task.Type).
			Ref("history").
			Unique().
			Required(),
	}
}
