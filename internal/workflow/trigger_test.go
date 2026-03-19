package workflow

import (
	"testing"

	"github.com/gobenpark/colign/internal/models"
)

func TestShouldTriggerTaskGeneration(t *testing.T) {
	tests := []struct {
		from    models.ChangeStage
		to      models.ChangeStage
		trigger bool
	}{
		{models.StageReview, models.StageReady, true},
		{models.StageDraft, models.StageDesign, false},
		{models.StageDesign, models.StageReview, false},
		{models.StageReady, models.StageReview, false}, // revert
	}

	for _, tt := range tests {
		got := ShouldTriggerTaskGeneration(tt.from, tt.to)
		if got != tt.trigger {
			t.Errorf("ShouldTriggerTaskGeneration(%s, %s) = %v, want %v", tt.from, tt.to, got, tt.trigger)
		}
	}
}
