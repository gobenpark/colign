package project

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/gobenpark/colign/internal/models"
	"github.com/uptrace/bun"
)

var (
	ErrProjectNotFound = errors.New("project not found")
	ErrNotAuthorized   = errors.New("not authorized")
	slugRegexp         = regexp.MustCompile(`[^a-z0-9-]+`)
)

type Service struct {
	db *bun.DB
}

func NewService(db *bun.DB) *Service {
	return &Service{db: db}
}

func GenerateSlug(name string) string {
	slug := strings.ToLower(strings.TrimSpace(name))
	slug = slugRegexp.ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")
	// Collapse multiple dashes
	for strings.Contains(slug, "--") {
		slug = strings.ReplaceAll(slug, "--", "-")
	}
	return slug
}

func (s *Service) ensureUniqueSlug(ctx context.Context, slug string) (string, error) {
	baseSlug := slug
	for i := 0; ; i++ {
		candidate := baseSlug
		if i > 0 {
			candidate = fmt.Sprintf("%s-%d", baseSlug, i+1)
		}
		exists, err := s.db.NewSelect().Model((*models.Project)(nil)).Where("slug = ?", candidate).Exists(ctx)
		if err != nil {
			return "", err
		}
		if !exists {
			return candidate, nil
		}
	}
}

type CreateProjectInput struct {
	Name        string
	Description string
	UserID      int64
}

func (s *Service) Create(ctx context.Context, input CreateProjectInput) (*models.Project, error) {
	slug := GenerateSlug(input.Name)
	uniqueSlug, err := s.ensureUniqueSlug(ctx, slug)
	if err != nil {
		return nil, err
	}

	project := &models.Project{
		Name:        input.Name,
		Slug:        uniqueSlug,
		Description: input.Description,
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	if _, err := tx.NewInsert().Model(project).Exec(ctx); err != nil {
		return nil, err
	}

	member := &models.ProjectMember{
		ProjectID: project.ID,
		UserID:    input.UserID,
		Role:      models.RoleOwner,
	}
	if _, err := tx.NewInsert().Model(member).Exec(ctx); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return project, nil
}

func (s *Service) GetBySlug(ctx context.Context, slug string) (*models.Project, []models.ProjectMember, error) {
	project := new(models.Project)
	err := s.db.NewSelect().Model(project).Where("slug = ?", slug).Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil, ErrProjectNotFound
		}
		return nil, nil, err
	}

	var members []models.ProjectMember
	err = s.db.NewSelect().Model(&members).
		Relation("User").
		Where("pm.project_id = ?", project.ID).
		Scan(ctx)
	if err != nil {
		return nil, nil, err
	}

	return project, members, nil
}

func (s *Service) ListByUser(ctx context.Context, userID int64) ([]models.Project, error) {
	var projects []models.Project
	err := s.db.NewSelect().Model(&projects).
		Join("JOIN project_members AS pm ON pm.project_id = p.id").
		Where("pm.user_id = ?", userID).
		OrderExpr("p.updated_at DESC").
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return projects, nil
}

type UpdateProjectInput struct {
	ID          int64
	Name        string
	Description string
}

func (s *Service) Update(ctx context.Context, input UpdateProjectInput) (*models.Project, error) {
	project := new(models.Project)
	err := s.db.NewSelect().Model(project).Where("id = ?", input.ID).Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrProjectNotFound
		}
		return nil, err
	}

	project.Name = input.Name
	project.Description = input.Description
	project.UpdatedAt = time.Now()

	if _, err := s.db.NewUpdate().Model(project).WherePK().Exec(ctx); err != nil {
		return nil, err
	}
	return project, nil
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	_, err := s.db.NewDelete().Model((*models.Project)(nil)).Where("id = ?", id).Exec(ctx)
	return err
}

func (s *Service) CreateChange(ctx context.Context, projectID int64, name string) (*models.Change, error) {
	change := &models.Change{
		ProjectID: projectID,
		Name:      name,
		Stage:     models.StageDraft,
	}

	if _, err := s.db.NewInsert().Model(change).Exec(ctx); err != nil {
		return nil, err
	}
	return change, nil
}

func (s *Service) ListChanges(ctx context.Context, projectID int64) ([]models.Change, error) {
	var changes []models.Change
	err := s.db.NewSelect().Model(&changes).
		Where("project_id = ?", projectID).
		OrderExpr("created_at DESC").
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return changes, nil
}

func (s *Service) GetChange(ctx context.Context, id int64) (*models.Change, error) {
	change := new(models.Change)
	err := s.db.NewSelect().Model(change).Where("id = ?", id).Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrProjectNotFound
		}
		return nil, err
	}
	return change, nil
}

func (s *Service) DeleteChange(ctx context.Context, id int64) error {
	_, err := s.db.NewDelete().Model((*models.Change)(nil)).Where("id = ?", id).Exec(ctx)
	return err
}

type InviteMemberInput struct {
	ProjectID int64
	Email     string
	Role      models.Role
}

func (s *Service) InviteMember(ctx context.Context, input InviteMemberInput) (*models.ProjectMember, error) {
	// Find user by email
	user := new(models.User)
	err := s.db.NewSelect().Model(user).Where("email = ?", input.Email).Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// User not registered — create pending invitation
			tokenBytes := make([]byte, 32)
			if _, err := rand.Read(tokenBytes); err != nil {
				return nil, err
			}
			invitation := &models.PendingInvitation{
				ProjectID: input.ProjectID,
				Email:     input.Email,
				Role:      input.Role,
				Token:     hex.EncodeToString(tokenBytes),
				ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
			}
			if _, err := s.db.NewInsert().Model(invitation).
				On("CONFLICT (project_id, email) DO UPDATE").
				Set("role = EXCLUDED.role").
				Set("token = EXCLUDED.token").
				Set("expires_at = EXCLUDED.expires_at").
				Exec(ctx); err != nil {
				return nil, err
			}
			// TODO: send invitation email with token
			return nil, nil
		}
		return nil, err
	}

	member := &models.ProjectMember{
		ProjectID: input.ProjectID,
		UserID:    user.ID,
		Role:      input.Role,
	}

	if _, err := s.db.NewInsert().Model(member).
		On("CONFLICT (project_id, user_id) DO UPDATE").
		Set("role = EXCLUDED.role").
		Exec(ctx); err != nil {
		return nil, err
	}

	return member, nil
}

// ProcessPendingInvitations converts pending invitations to project memberships
// after a new user registers. Called during user registration flow.
func (s *Service) ProcessPendingInvitations(ctx context.Context, userID int64, email string) error {
	var invitations []models.PendingInvitation
	err := s.db.NewSelect().Model(&invitations).
		Where("email = ?", email).
		Where("expires_at > ?", time.Now()).
		Scan(ctx)
	if err != nil {
		return err
	}

	for _, inv := range invitations {
		member := &models.ProjectMember{
			ProjectID: inv.ProjectID,
			UserID:    userID,
			Role:      inv.Role,
		}
		if _, err := s.db.NewInsert().Model(member).
			On("CONFLICT (project_id, user_id) DO NOTHING").
			Exec(ctx); err != nil {
			return err
		}
	}

	// Clean up processed invitations
	if len(invitations) > 0 {
		_, err = s.db.NewDelete().Model((*models.PendingInvitation)(nil)).
			Where("email = ?", email).
			Exec(ctx)
	}

	return err
}
