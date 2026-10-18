package repository

import (
	"context"
	"database/sql"
	"time"

	"marketplace/notification/model"
)

type NotificationRepository struct {
	db *sql.DB
}

func NewNotificationRepository(db *sql.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

func (r *NotificationRepository) Create(ctx context.Context, n *model.Notification) error {
	query := `
		INSERT INTO notifications (user_id, channel, subject, body, status, read, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING id`
	return r.db.QueryRowContext(ctx, query,
		n.UserID, n.Channel, n.Subject, n.Body, n.Status, n.Read, n.CreatedAt,
	).Scan(&n.ID)
}

func (r *NotificationRepository) MarkSent(ctx context.Context, id int64) error {
	now := time.Now()
	_, err := r.db.ExecContext(ctx,
		`UPDATE notifications SET status='sent', sent_at=$1 WHERE id=$2`, now, id)
	return err
}

func (r *NotificationRepository) MarkFailed(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `UPDATE notifications SET status='failed' WHERE id=$1`, id)
	return err
}

func (r *NotificationRepository) ListForUser(ctx context.Context, userID int64) ([]*model.Notification, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, user_id, channel, subject, body, status, read, created_at, sent_at
		 FROM notifications WHERE user_id=$1 ORDER BY created_at DESC LIMIT 50`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []*model.Notification{}
	for rows.Next() {
		n := &model.Notification{}
		if err := rows.Scan(&n.ID, &n.UserID, &n.Channel, &n.Subject, &n.Body,
			&n.Status, &n.Read, &n.CreatedAt, &n.SentAt); err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	return out, rows.Err()
}

func (r *NotificationRepository) UnreadCount(ctx context.Context, userID int64) (int, error) {
	var n int
	err := r.db.QueryRowContext(ctx,
		`SELECT count(*) FROM notifications WHERE user_id = $1 AND read = false AND channel = 'in_app'`, userID).Scan(&n)
	return n, err
}

func (r *NotificationRepository) MarkAllRead(ctx context.Context, userID int64) error {
	_, err := r.db.ExecContext(ctx, `UPDATE notifications SET read = true WHERE user_id = $1 AND read = false`, userID)
	return err
}

func (r *NotificationRepository) MarkRead(ctx context.Context, id, userID int64) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE notifications SET read=true WHERE id=$1 AND user_id=$2`, id, userID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}
