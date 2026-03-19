package models

import "testing"

func TestChangeStages(t *testing.T) {
	stages := []ChangeStage{StageDraft, StageDesign, StageReview, StageReady}
	expected := []string{"draft", "design", "review", "ready"}

	for i, stage := range stages {
		if string(stage) != expected[i] {
			t.Errorf("expected stage '%s', got '%s'", expected[i], stage)
		}
	}
}

func TestChangeModel(t *testing.T) {
	c := &Change{
		ProjectID: 1,
		Name:      "add-user-auth",
		Stage:     StageDraft,
	}

	if c.Stage != StageDraft {
		t.Errorf("expected stage '%s', got '%s'", StageDraft, c.Stage)
	}
	if c.Name != "add-user-auth" {
		t.Errorf("expected name 'add-user-auth', got '%s'", c.Name)
	}
}

func TestChangeStageOrder(t *testing.T) {
	order := StageOrder()
	if len(order) != 4 {
		t.Fatalf("expected 4 stages, got %d", len(order))
	}
	if order[0] != StageDraft {
		t.Errorf("first stage should be Draft")
	}
	if order[3] != StageReady {
		t.Errorf("last stage should be Ready")
	}
}
