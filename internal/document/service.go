package document

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/gobenpark/colign/internal/models"
	"github.com/uptrace/bun"
)

var ErrDocumentNotFound = errors.New("document not found")

type Service struct {
	db *bun.DB
}

func NewService(db *bun.DB) *Service {
	return &Service{db: db}
}

type SaveInput struct {
	ChangeID int64
	Type     models.DocumentType
	Title    string
	Content  string
	UserID   int64
}

func (s *Service) Save(ctx context.Context, input SaveInput) (*models.Document, error) {
	doc := new(models.Document)
	err := s.db.NewSelect().Model(doc).
		Where("change_id = ?", input.ChangeID).
		Where("type = ?", input.Type).
		Where("COALESCE(title, '') = COALESCE(?, '')", input.Title).
		Scan(ctx)

	if errors.Is(err, sql.ErrNoRows) {
		// Create new document
		doc = &models.Document{
			ChangeID: input.ChangeID,
			Type:     input.Type,
			Title:    input.Title,
			Content:  input.Content,
			Version:  1,
		}
		if _, err := s.db.NewInsert().Model(doc).Exec(ctx); err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	} else {
		// Update existing
		doc.Content = input.Content
		doc.Version++
		doc.UpdatedAt = time.Now()
		if _, err := s.db.NewUpdate().Model(doc).WherePK().Exec(ctx); err != nil {
			return nil, err
		}
	}

	// Save version history
	version := &models.DocumentVersion{
		DocumentID: doc.ID,
		Content:    input.Content,
		Version:    doc.Version,
		UserID:     input.UserID,
	}
	s.db.NewInsert().Model(version).Exec(ctx)

	return doc, nil
}

func (s *Service) GetHistory(ctx context.Context, documentID int64) ([]models.DocumentVersion, error) {
	var versions []models.DocumentVersion
	err := s.db.NewSelect().Model(&versions).
		Where("document_id = ?", documentID).
		OrderExpr("version DESC").
		Scan(ctx)
	return versions, err
}

func (s *Service) Restore(ctx context.Context, documentID int64, version int, userID int64) (*models.Document, error) {
	dv := new(models.DocumentVersion)
	err := s.db.NewSelect().Model(dv).
		Where("document_id = ?", documentID).
		Where("version = ?", version).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrDocumentNotFound
		}
		return nil, err
	}

	doc := new(models.Document)
	if err := s.db.NewSelect().Model(doc).Where("id = ?", documentID).Scan(ctx); err != nil {
		return nil, err
	}

	doc.Content = dv.Content
	doc.Version++
	doc.UpdatedAt = time.Now()
	if _, err := s.db.NewUpdate().Model(doc).WherePK().Exec(ctx); err != nil {
		return nil, err
	}

	// Record restore as new version
	newVersion := &models.DocumentVersion{
		DocumentID: doc.ID,
		Content:    dv.Content,
		Version:    doc.Version,
		UserID:     userID,
	}
	s.db.NewInsert().Model(newVersion).Exec(ctx)

	return doc, nil
}
