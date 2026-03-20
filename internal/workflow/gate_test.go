package workflow

import (
	"testing"

	"github.com/gobenpark/colign/internal/models"
)

func TestGateConditions_Draft(t *testing.T) {
	gate := NewGateChecker()

	// Draft → Design: proposal must exist
	conditions := gate.Check(models.StageDraft, GateInput{
		HasProposal: false,
		HasDesign:   false,
	})

	if len(conditions) == 0 {
		t.Error("expected conditions for Draft gate")
	}
	if conditions[0].Met {
		t.Error("proposal condition should not be met")
	}

	// With proposal
	conditions = gate.Check(models.StageDraft, GateInput{
		HasProposal: true,
	})
	if !conditions[0].Met {
		t.Error("proposal condition should be met")
	}
}

func TestGateConditions_Design(t *testing.T) {
	gate := NewGateChecker()

	// Design → Review: design must exist
	conditions := gate.Check(models.StageDesign, GateInput{
		HasProposal: true,
		HasDesign:   false,
	})

	if len(conditions) == 0 {
		t.Error("expected conditions for Design gate")
	}
	if conditions[0].Name != "design" {
		t.Error("expected design condition")
	}
	if conditions[0].Met {
		t.Error("design condition should not be met")
	}

	// With design
	conditions = gate.Check(models.StageDesign, GateInput{
		HasProposal: true,
		HasDesign:   true,
	})

	for _, c := range conditions {
		if !c.Met {
			t.Errorf("condition %s should be met", c.Name)
		}
	}
}

func TestGateConditions_AllMet(t *testing.T) {
	gate := NewGateChecker()

	if !gate.AllMet(models.StageDraft, GateInput{HasProposal: true}) {
		t.Error("Draft gate should pass with proposal")
	}

	if gate.AllMet(models.StageDraft, GateInput{HasProposal: false}) {
		t.Error("Draft gate should fail without proposal")
	}
}
