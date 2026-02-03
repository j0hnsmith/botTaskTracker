package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// TaskTag holds the schema definition for the TaskTag entity (k:v pairs).
type TaskTag struct {
	ent.Schema
}

// Fields of the TaskTag.
func (TaskTag) Fields() []ent.Field {
	return []ent.Field{
		field.String("key").
			NotEmpty(), // e.g., "project", "priority", "readyToStart", "type"
		field.String("value").
			NotEmpty(), // e.g., "databacked", "high", "true", "feature"
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
	}
}

// Edges of the TaskTag.
func (TaskTag) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("task", Task.Type).
			Ref("tags").
			Unique().
			Required(),
	}
}
