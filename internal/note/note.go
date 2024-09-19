package note

import "fmt"

type Note struct {
	id   uint64
	text string
}

// NewNote creates a new note
func NewNote(id uint64, text string) *Note {
	return &Note{id, text}
}

func (n *Note) ID() uint64 {
	return n.id
}

func (n *Note) Text() string {
	return n.text
}

func (n *Note) ToJSON() string {
	return fmt.Sprintf(`{"id": %d, "text": %q}`, n.id, n.text)
}
