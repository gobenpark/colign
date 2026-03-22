package models

import (
	"time"

	"github.com/uptrace/bun"
)

type APIToken struct {
	bun.BaseModel `bun:"table:api_tokens,alias:at"`

	ID            int64      `bun:"id,pk,autoincrement"`
	UserID        int64      `bun:"user_id,notnull"`
	OrgID         int64      `bun:"org_id,notnull"`
	Name          string     `bun:"name,notnull"`
	TokenType     string     `bun:"token_type,notnull,default:'personal'"`
	OAuthClientID string     `bun:"oauth_client_id"`
	TokenHash     string     `bun:"token_hash,notnull,unique"`
	Prefix        string     `bun:"prefix,notnull"`
	LastUsedAt    *time.Time `bun:"last_used_at"`
	CreatedAt     time.Time  `bun:"created_at,notnull,default:current_timestamp"`

	User         *User         `bun:"rel:belongs-to,join:user_id=id"`
	Organization *Organization `bun:"rel:belongs-to,join:org_id=id"`
}
