package models

import (
	"time"

	"github.com/uptrace/bun"
)

type Role string

const (
	RoleOwner  Role = "owner"
	RoleEditor Role = "editor"
	RoleViewer Role = "viewer"
)

type Project struct {
	bun.BaseModel `bun:"table:projects,alias:p"`

	ID          int64     `bun:"id,pk,autoincrement"`
	Name        string    `bun:"name,notnull"`
	Slug        string    `bun:"slug,notnull,unique"`
	Description string    `bun:"description"`
	CreatedAt   time.Time `bun:"created_at,notnull,default:current_timestamp"`
	UpdatedAt   time.Time `bun:"updated_at,notnull,default:current_timestamp"`

	Members []ProjectMember `bun:"rel:has-many,join:id=project_id"`
}

type ProjectMember struct {
	bun.BaseModel `bun:"table:project_members,alias:pm"`

	ID        int64     `bun:"id,pk,autoincrement"`
	ProjectID int64     `bun:"project_id,notnull"`
	UserID    int64     `bun:"user_id,notnull"`
	Role      Role      `bun:"role,notnull,default:'viewer'"`
	CreatedAt time.Time `bun:"created_at,notnull,default:current_timestamp"`

	Project *Project `bun:"rel:belongs-to,join:project_id=id"`
	User    *User    `bun:"rel:belongs-to,join:user_id=id"`
}
