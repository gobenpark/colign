package workflow

import (
	"testing"

	"github.com/gobenpark/colign/internal/models"
)

func TestShouldAutoAdvance(t *testing.T) {
	gate := NewGateChecker()

	// Draft with proposal done → should advance
	if !gate.AllMet(models.StageDraft, GateInput{HasProposal: true}) {
		t.Error("should auto-advance from Draft when proposal exists")
	}

	// Design with design done → should advance
	if !gate.AllMet(models.StageDesign, GateInput{
		HasDesign: true,
	}) {
		t.Error("should auto-advance from Design when design complete")
	}

	// Review with approvals done → should advance
	if !gate.AllMet(models.StageReview, GateInput{
		ApprovalsNeeded: 2,
		ApprovalsDone:   2,
	}) {
		t.Error("should auto-advance from Review when approvals met")
	}

	// Review without enough approvals → should not advance
	if gate.AllMet(models.StageReview, GateInput{
		ApprovalsNeeded: 2,
		ApprovalsDone:   1,
	}) {
		t.Error("should not auto-advance from Review with insufficient approvals")
	}
}
