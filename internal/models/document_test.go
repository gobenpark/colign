package models

import "testing"

func TestDocumentTypes(t *testing.T) {
	types := []DocumentType{DocProposal, DocDesign, DocTasks}
	expected := []string{"proposal", "design", "tasks"}

	for i, dt := range types {
		if string(dt) != expected[i] {
			t.Errorf("expected %s, got %s", expected[i], dt)
		}
	}
}

func TestDocumentModel(t *testing.T) {
	doc := &Document{
		ChangeID: 1,
		Type:     DocProposal,
		Content:  "## Why\n\nTest content",
	}

	if doc.Type != DocProposal {
		t.Errorf("expected DocProposal, got %s", doc.Type)
	}
}
