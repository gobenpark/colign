package workflow

import "github.com/gobenpark/colign/internal/models"

var stageIndex = map[models.ChangeStage]int{
	models.StageDraft:  0,
	models.StageDesign: 1,
	models.StageReview: 2,
	models.StageReady:  3,
}

var stages = models.StageOrder()

// CanTransition checks if a transition from one stage to another is allowed.
// Only adjacent stage transitions are permitted (forward or backward by 1).
func CanTransition(from, to models.ChangeStage) bool {
	fi, ok1 := stageIndex[from]
	ti, ok2 := stageIndex[to]
	if !ok1 || !ok2 {
		return false
	}
	diff := ti - fi
	return diff == 1 || diff == -1
}

// NextStage returns the next stage in the workflow, if one exists.
func NextStage(current models.ChangeStage) (models.ChangeStage, bool) {
	idx, ok := stageIndex[current]
	if !ok || idx >= len(stages)-1 {
		return "", false
	}
	return stages[idx+1], true
}

// PrevStage returns the previous stage in the workflow, if one exists.
func PrevStage(current models.ChangeStage) (models.ChangeStage, bool) {
	idx, ok := stageIndex[current]
	if !ok || idx <= 0 {
		return "", false
	}
	return stages[idx-1], true
}
