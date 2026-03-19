package models

import (
	"time"

	"github.com/uptrace/bun"
)

type User struct {
	bun.BaseModel `bun:"table:users,alias:u"`

	ID            int64     `bun:"id,pk,autoincrement"`
	Email         string    `bun:"email,notnull,unique"`
	PasswordHash  string    `bun:"password_hash"`
	Name          string    `bun:"name,notnull"`
	AvatarURL     string    `bun:"avatar_url"`
	EmailVerified bool      `bun:"email_verified,notnull,default:false"`
	CreatedAt     time.Time `bun:"created_at,notnull,default:current_timestamp"`
	UpdatedAt     time.Time `bun:"updated_at,notnull,default:current_timestamp"`

	Accounts []Account `bun:"rel:has-many,join:id=user_id"`
}

type Account struct {
	bun.BaseModel `bun:"table:accounts,alias:a"`

	ID                int64  `bun:"id,pk,autoincrement"`
	UserID            int64  `bun:"user_id,notnull"`
	Provider          string `bun:"provider,notnull"`
	ProviderAccountID string `bun:"provider_account_id,notnull"`
	AccessToken       string `bun:"access_token"`
	RefreshToken      string `bun:"refresh_token"`
	ExpiresAt         int64  `bun:"expires_at"`
	CreatedAt         time.Time `bun:"created_at,notnull,default:current_timestamp"`

	User *User `bun:"rel:belongs-to,join:user_id=id"`
}

type Session struct {
	bun.BaseModel `bun:"table:sessions,alias:s"`

	ID           int64     `bun:"id,pk,autoincrement"`
	UserID       int64     `bun:"user_id,notnull"`
	RefreshToken string    `bun:"refresh_token,notnull,unique"`
	UserAgent    string    `bun:"user_agent"`
	IP           string    `bun:"ip"`
	ExpiresAt    time.Time `bun:"expires_at,notnull"`
	CreatedAt    time.Time `bun:"created_at,notnull,default:current_timestamp"`

	User *User `bun:"rel:belongs-to,join:user_id=id"`
}

type EmailVerification struct {
	bun.BaseModel `bun:"table:email_verifications,alias:ev"`

	ID        int64     `bun:"id,pk,autoincrement"`
	UserID    int64     `bun:"user_id,notnull"`
	Token     string    `bun:"token,notnull,unique"`
	ExpiresAt time.Time `bun:"expires_at,notnull"`
	CreatedAt time.Time `bun:"created_at,notnull,default:current_timestamp"`

	User *User `bun:"rel:belongs-to,join:user_id=id"`
}
