package models

import (
	"time"

	"github.com/uptrace/bun"
)

type OrgRole string

const (
	OrgRoleOwner  OrgRole = "owner"
	OrgRoleAdmin  OrgRole = "admin"
	OrgRoleMember OrgRole = "member"
)

type Organization struct {
	bun.BaseModel `bun:"table:organizations,alias:o"`

	ID             int64     `bun:"id,pk,autoincrement"`
	Name           string    `bun:"name,notnull"`
	Slug           string    `bun:"slug,notnull,unique"`
	AllowedDomains []string  `bun:"allowed_domains,array,notnull,default:'{}'"`
	CreatedAt      time.Time `bun:"created_at,notnull,default:current_timestamp"`
	UpdatedAt      time.Time `bun:"updated_at,notnull,default:current_timestamp"`

	Members []OrganizationMember `bun:"rel:has-many,join:id=organization_id"`
}

type InvitationStatus string

const (
	InvitationStatusPending  InvitationStatus = "pending"
	InvitationStatusAccepted InvitationStatus = "accepted"
	InvitationStatusExpired  InvitationStatus = "expired"
)

type OrgInvitation struct {
	bun.BaseModel `bun:"table:org_invitations,alias:oi"`

	ID             int64            `bun:"id,pk,autoincrement"`
	OrganizationID int64            `bun:"organization_id,notnull"`
	Email          string           `bun:"email,notnull"`
	Role           OrgRole          `bun:"role,notnull,default:'member'"`
	Token          string           `bun:"token,notnull,unique"`
	InvitedBy      int64            `bun:"invited_by,notnull"`
	Status         InvitationStatus `bun:"status,notnull,default:'pending'"`
	ExpiresAt      time.Time        `bun:"expires_at,notnull"`
	CreatedAt      time.Time        `bun:"created_at,notnull,default:current_timestamp"`

	Organization *Organization `bun:"rel:belongs-to,join:organization_id=id"`
	Inviter      *User         `bun:"rel:belongs-to,join:invited_by=id"`
}

type OrganizationMember struct {
	bun.BaseModel `bun:"table:organization_members,alias:om"`

	ID             int64     `bun:"id,pk,autoincrement"`
	OrganizationID int64     `bun:"organization_id,notnull"`
	UserID         int64     `bun:"user_id,notnull"`
	Role           OrgRole   `bun:"role,notnull,default:'member'"`
	CreatedAt      time.Time `bun:"created_at,notnull,default:current_timestamp"`

	Organization *Organization `bun:"rel:belongs-to,join:organization_id=id"`
	User         *User         `bun:"rel:belongs-to,join:user_id=id"`
}
