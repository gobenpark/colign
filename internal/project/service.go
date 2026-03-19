package project

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/gobenpark/CoSpec/internal/models"
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
			// TODO: handle invite for non-registered users
			return nil, fmt.Errorf("user not found: %s", input.Email)
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
