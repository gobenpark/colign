package models

import (
	"time"

	"github.com/uptrace/bun"
)

type ProjectMemory struct {
	bun.BaseModel `bun:"table:project_memories,alias:pm"`

	ID        int64     `bun:"id,pk,autoincrement"`
	ProjectID int64     `bun:"project_id,notnull"`
	Content   string    `bun:"content,notnull,type:text"`
	UpdatedBy int64     `bun:"updated_by"`
	CreatedAt time.Time `bun:"created_at,notnull,default:current_timestamp"`
	UpdatedAt time.Time `bun:"updated_at,notnull,default:current_timestamp"`

	Project *Project `bun:"rel:belongs-to,join:project_id=id"`
	User    *User    `bun:"rel:belongs-to,join:updated_by=id"`
}
