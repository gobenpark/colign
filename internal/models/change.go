package models

import (
	"time"

	"github.com/uptrace/bun"
)

type ChangeStage string

const (
	StageDraft  ChangeStage = "draft"
	StageDesign ChangeStage = "design"
	StageReview ChangeStage = "review"
	StageReady  ChangeStage = "ready"
)

func StageOrder() []ChangeStage {
	return []ChangeStage{StageDraft, StageDesign, StageReview, StageReady}
}

type ChangeType string

const (
	ChangeFeature  ChangeType = "feature"
	ChangeBugfix   ChangeType = "bugfix"
	ChangeRefactor ChangeType = "refactor"
)

type Change struct {
	bun.BaseModel `bun:"table:changes,alias:ch"`

	ID         int64       `bun:"id,pk,autoincrement"`
	ProjectID  int64       `bun:"project_id,notnull"`
	Name       string      `bun:"name,notnull"`
	Stage      ChangeStage `bun:"stage,notnull,default:'draft'"`
	ChangeType ChangeType  `bun:"change_type,notnull,default:'feature'"`
	CreatedAt  time.Time   `bun:"created_at,notnull,default:current_timestamp"`
	UpdatedAt  time.Time   `bun:"updated_at,notnull,default:current_timestamp"`

	Project *Project `bun:"rel:belongs-to,join:project_id=id"`
}
