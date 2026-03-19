package workflow

import "github.com/gobenpark/colign/internal/models"

// ShouldTriggerTaskGeneration returns true when a change transitions to Ready stage.
func ShouldTriggerTaskGeneration(from, to models.ChangeStage) bool {
	return from == models.StageReview && to == models.StageReady
}
