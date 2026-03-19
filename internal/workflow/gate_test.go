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
		SpecsCount:  0,
		SpecsDone:   0,
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

	// Design → Review: design + all specs must exist
	conditions := gate.Check(models.StageDesign, GateInput{
		HasProposal: true,
		HasDesign:   false,
		SpecsCount:  3,
		SpecsDone:   1,
	})

	var designCond, specsCond *GateCondition
	for i := range conditions {
		switch conditions[i].Name {
		case "design":
			designCond = &conditions[i]
		case "specs":
			specsCond = &conditions[i]
		}
	}

	if designCond == nil || designCond.Met {
		t.Error("design condition should exist and not be met")
	}
	if specsCond == nil || specsCond.Met {
		t.Error("specs condition should exist and not be met")
	}

	// All done
	conditions = gate.Check(models.StageDesign, GateInput{
		HasProposal: true,
		HasDesign:   true,
		SpecsCount:  3,
		SpecsDone:   3,
	})

	allMet := true
	for _, c := range conditions {
		if !c.Met {
			allMet = false
		}
	}
	if !allMet {
		t.Error("all design gate conditions should be met")
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
