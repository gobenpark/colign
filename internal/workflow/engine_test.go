package workflow

import (
	"testing"

	"github.com/gobenpark/colign/internal/models"
)

func TestCanTransition(t *testing.T) {
	tests := []struct {
		from    models.ChangeStage
		to      models.ChangeStage
		allowed bool
	}{
		{models.StageDraft, models.StageDesign, true},
		{models.StageDesign, models.StageReview, true},
		{models.StageReview, models.StageReady, true},
		// Can revert
		{models.StageDesign, models.StageDraft, true},
		{models.StageReview, models.StageDesign, true},
		{models.StageReady, models.StageReview, true},
		// Cannot skip
		{models.StageDraft, models.StageReview, false},
		{models.StageDraft, models.StageReady, false},
		{models.StageDesign, models.StageReady, false},
		// Cannot self-transition
		{models.StageDraft, models.StageDraft, false},
	}

	for _, tt := range tests {
		got := CanTransition(tt.from, tt.to)
		if got != tt.allowed {
			t.Errorf("CanTransition(%s, %s) = %v, want %v", tt.from, tt.to, got, tt.allowed)
		}
	}
}

func TestNextStage(t *testing.T) {
	tests := []struct {
		current  models.ChangeStage
		expected models.ChangeStage
		hasNext  bool
	}{
		{models.StageDraft, models.StageDesign, true},
		{models.StageDesign, models.StageReview, true},
		{models.StageReview, models.StageReady, true},
		{models.StageReady, "", false},
	}

	for _, tt := range tests {
		next, ok := NextStage(tt.current)
		if ok != tt.hasNext {
			t.Errorf("NextStage(%s) hasNext = %v, want %v", tt.current, ok, tt.hasNext)
		}
		if ok && next != tt.expected {
			t.Errorf("NextStage(%s) = %s, want %s", tt.current, next, tt.expected)
		}
	}
}

func TestPrevStage(t *testing.T) {
	tests := []struct {
		current  models.ChangeStage
		expected models.ChangeStage
		hasPrev  bool
	}{
		{models.StageDraft, "", false},
		{models.StageDesign, models.StageDraft, true},
		{models.StageReview, models.StageDesign, true},
		{models.StageReady, models.StageReview, true},
	}

	for _, tt := range tests {
		prev, ok := PrevStage(tt.current)
		if ok != tt.hasPrev {
			t.Errorf("PrevStage(%s) hasPrev = %v, want %v", tt.current, ok, tt.hasPrev)
		}
		if ok && prev != tt.expected {
			t.Errorf("PrevStage(%s) = %s, want %s", tt.current, prev, tt.expected)
		}
	}
}
