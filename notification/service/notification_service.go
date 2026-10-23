package service

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"time"
	"unicode/utf8"

	"mercatus/notification/model"
	"mercatus/notification/repository"
)

var ErrInvalidInput = errors.New("invalid notification")

type Deliverer interface {
	Deliver(ctx context.Context, n *model.Notification) error
}

type LogDeliverer struct{}

func (LogDeliverer) Deliver(ctx context.Context, n *model.Notification) error {
	slog.InfoContext(ctx, "notification delivered", "channel", n.Channel, "user_id", n.UserID, "notification_id", n.ID)
	return nil
}

type NotificationService struct {
	notifications *repository.NotificationRepository
	deliverer     Deliverer
}

func NewNotificationService(n *repository.NotificationRepository, d Deliverer) *NotificationService {
	if d == nil {
		d = LogDeliverer{}
	}
	return &NotificationService{notifications: n, deliverer: d}
}

func (s *NotificationService) Send(ctx context.Context, req *model.SendRequest) (*model.Notification, error) {
	switch req.Channel {
	case model.ChannelEmail, model.ChannelSMS, model.ChannelInApp:
	case "":
		req.Channel = model.ChannelInApp
	default:
		return nil, ErrInvalidInput
	}
	req.Subject = strings.TrimSpace(req.Subject)
	if req.UserID <= 0 || req.Subject == "" || utf8.RuneCountInString(req.Subject) > 200 || utf8.RuneCountInString(req.Body) > 5000 {
		return nil, ErrInvalidInput
	}
	n := &model.Notification{
		UserID: req.UserID, Channel: req.Channel, Subject: req.Subject, Body: req.Body,
		Status: model.StatusPending, CreatedAt: time.Now(),
	}
	if err := s.notifications.Create(ctx, n); err != nil {
		return nil, err
	}
	if err := s.deliverer.Deliver(ctx, n); err != nil {
		_ = s.notifications.MarkFailed(ctx, n.ID)
		n.Status = model.StatusFailed
		return n, nil
	}
	if err := s.notifications.MarkSent(ctx, n.ID); err != nil {
		return nil, err
	}
	n.Status = model.StatusSent
	return n, nil
}

func (s *NotificationService) ListForUser(ctx context.Context, userID int64) ([]*model.Notification, error) {
	return s.notifications.ListForUser(ctx, userID)
}

func (s *NotificationService) UnreadCount(ctx context.Context, userID int64) (int, error) {
	return s.notifications.UnreadCount(ctx, userID)
}

func (s *NotificationService) MarkRead(ctx context.Context, id, userID int64) error {
	return s.notifications.MarkRead(ctx, id, userID)
}

func (s *NotificationService) MarkAllRead(ctx context.Context, userID int64) error {
	return s.notifications.MarkAllRead(ctx, userID)
}
