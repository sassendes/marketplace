package repository

import (
	"context"
	"database/sql"
	"errors"

	"mercatus/catalog/model"
	"mercatus/platform/database"
)

var ErrDuplicateCategory = errors.New("category already exists")

type CategoryRepository struct {
	db *sql.DB
}

func NewCategoryRepository(db *sql.DB) *CategoryRepository {
	return &CategoryRepository{db: db}
}

func (r *CategoryRepository) Create(ctx context.Context, c *model.Category) error {
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO categories (parent_id, name, slug, created_at) VALUES ($1, $2, $3, $4) RETURNING id`,
		c.ParentID, c.Name, c.Slug, c.CreatedAt).Scan(&c.ID)
	if database.IsUniqueViolation(err) {
		return ErrDuplicateCategory
	}
	return err
}

func (r *CategoryRepository) Exists(ctx context.Context, id int64) (bool, error) {
	var ok bool
	err := r.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM categories WHERE id = $1)`, id).Scan(&ok)
	return ok, err
}

func (r *CategoryRepository) List(ctx context.Context) ([]*model.Category, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT c.id, c.parent_id, c.name, c.slug, c.created_at,
		        (SELECT count(*) FROM products p JOIN sellers s ON s.id = p.seller_id AND s.status = 'active'
		         LEFT JOIN categories pc ON pc.id = p.category_id
		         WHERE p.status = 'active' AND (p.category_id = c.id OR pc.parent_id = c.id))::int
		 FROM categories c ORDER BY c.name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []*model.Category{}
	for rows.Next() {
		c := &model.Category{}
		if err := rows.Scan(&c.ID, &c.ParentID, &c.Name, &c.Slug, &c.CreatedAt, &c.ProductCount); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}
