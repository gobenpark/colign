package memory

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/uptrace/bun"

	"github.com/gobenpark/colign/internal/models"
)

type Service struct {
	db *bun.DB
}

func NewService(db *bun.DB) *Service {
	return &Service{db: db}
}

func (s *Service) Get(ctx context.Context, projectID int64) (*models.ProjectMemory, error) {
	mem := new(models.ProjectMemory)
	err := s.db.NewSelect().Model(mem).
		Where("project_id = ?", projectID).
		Relation("User").
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return mem, nil
}

func (s *Service) Save(ctx context.Context, projectID int64, content string, userID int64) (*models.ProjectMemory, error) {
	mem := new(models.ProjectMemory)
	err := s.db.NewSelect().Model(mem).
		Where("project_id = ?", projectID).
		Scan(ctx)

	if errors.Is(err, sql.ErrNoRows) {
		// Create
		mem = &models.ProjectMemory{
			ProjectID: projectID,
			Content:   content,
			UpdatedBy: userID,
		}
		if _, err := s.db.NewInsert().Model(mem).Exec(ctx); err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	} else {
		// Update
		mem.Content = content
		mem.UpdatedBy = userID
		mem.UpdatedAt = time.Now()
		if _, err := s.db.NewUpdate().Model(mem).WherePK().Exec(ctx); err != nil {
			return nil, err
		}
	}

	return mem, nil
}
