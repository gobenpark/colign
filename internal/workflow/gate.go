package workflow

import "github.com/gobenpark/colign/internal/models"

type GateInput struct {
	HasProposal     bool
	HasDesign       bool
	ApprovalsNeeded int
	ApprovalsDone   int
}

type GateCondition struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Met         bool   `json:"met"`
}

type GateChecker struct{}

func NewGateChecker() *GateChecker {
	return &GateChecker{}
}

// Check returns the gate conditions for the current stage.
func (g *GateChecker) Check(stage models.ChangeStage, input GateInput) []GateCondition {
	switch stage {
	case models.StageDraft:
		return []GateCondition{
			{
				Name:        "proposal",
				Description: "Proposal document saved",
				Met:         input.HasProposal,
			},
		}
	case models.StageDesign:
		return []GateCondition{
			{
				Name:        "design",
				Description: "Design document saved",
				Met:         input.HasDesign,
			},
		}
	case models.StageReview:
		return []GateCondition{
			{
				Name:        "approvals",
				Description: "Required approvals received",
				Met:         input.ApprovalsDone >= input.ApprovalsNeeded,
			},
		}
	default:
		return nil
	}
}

// AllMet returns true if all gate conditions for the stage are satisfied.
func (g *GateChecker) AllMet(stage models.ChangeStage, input GateInput) bool {
	conditions := g.Check(stage, input)
	for _, c := range conditions {
		if !c.Met {
			return false
		}
	}
	return true
}
