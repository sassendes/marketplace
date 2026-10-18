package repository

import (
	"context"
	"database/sql"

	"marketplace/media/model"
)

type MediaRepository struct{ db *sql.DB }

func NewMediaRepository(db *sql.DB) *MediaRepository { return &MediaRepository{db: db} }

func (r *MediaRepository) Create(ctx context.Context, m *model.MediaAsset) error {
	return r.db.QueryRowContext(ctx,
		`INSERT INTO media_assets (owner_id, kind, object_key, content_type, status, created_at)
		 VALUES ($1,$2,$3,$4,'pending',now()) RETURNING id, created_at`,
		m.OwnerID, m.Kind, m.ObjectKey, m.ContentType,
	).Scan(&m.ID, &m.CreatedAt)
}

func (r *MediaRepository) MarkReady(ctx context.Context, id, ownerID int64) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE media_assets SET status='ready' WHERE id=$1 AND owner_id=$2 AND status='pending'`, id, ownerID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *MediaRepository) Get(ctx context.Context, id int64) (*model.MediaAsset, error) {
	m := &model.MediaAsset{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, owner_id, kind, object_key, content_type, status, created_at FROM media_assets WHERE id=$1`, id,
	).Scan(&m.ID, &m.OwnerID, &m.Kind, &m.ObjectKey, &m.ContentType, &m.Status, &m.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return m, nil
}
