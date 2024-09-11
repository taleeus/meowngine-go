package meowngine

import (
	"slices"

	"github.com/taleeus/commons-go/data"
)

// ID is the unique identifier type for Entities
type ID uint64

// Entity represents a unique game object.
// It's basically just an ID with a set of tags and components.
type Entity[S any] struct {
	id      ID
	deleted bool

	tags data.Set[string]

	world *World[S]
}

// Tags returns all entity tags
func (ent *Entity[S]) Tags() []string {
	return slices.Collect(ent.tags.Values())
}

// PushTag adds a tag to the entity's set (if not present)
func (ent *Entity[S]) PushTag(tag string) {
	ent.tags.Push(tag)
}

// RemoveTag removes a tag from the entity's set (if present)
func (ent *Entity[S]) RemoveTag(tag string) {
	ent.tags.Delete(tag)
}

// HasTag checks if the entity has the given tag
func (ent *Entity[S]) HasTag(tag string) bool {
	return ent.tags.Contains(tag)
}
