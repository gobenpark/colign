package models

import (
	"time"

	"github.com/uptrace/bun"
)

type DocumentType string

const (
	DocProposal DocumentType = "proposal"
	DocDesign   DocumentType = "design"
	DocTasks    DocumentType = "tasks"
)

type Document struct {
	bun.BaseModel `bun:"table:documents,alias:d"`

	ID        int64        `bun:"id,pk,autoincrement"`
	ChangeID  int64        `bun:"change_id,notnull"`
	Type      DocumentType `bun:"type,notnull"`
	Title     string       `bun:"title"`
	Content   string       `bun:"content,notnull,type:text"`
	Version   int          `bun:"version,notnull,default:1"`
	CreatedAt time.Time    `bun:"created_at,notnull,default:current_timestamp"`
	UpdatedAt time.Time    `bun:"updated_at,notnull,default:current_timestamp"`

	Change *Change `bun:"rel:belongs-to,join:change_id=id"`
}

type DocumentVersion struct {
	bun.BaseModel `bun:"table:document_versions,alias:dv"`

	ID         int64     `bun:"id,pk,autoincrement"`
	DocumentID int64     `bun:"document_id,notnull"`
	Content    string    `bun:"content,notnull,type:text"`
	Version    int       `bun:"version,notnull"`
	UserID     int64     `bun:"user_id"`
	CreatedAt  time.Time `bun:"created_at,notnull,default:current_timestamp"`

	Document *Document `bun:"rel:belongs-to,join:document_id=id"`
	User     *User     `bun:"rel:belongs-to,join:user_id=id"`
}
