package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"sync/atomic"
)

var ErrDeclined = errors.New("payment declined")

type Provider interface {
	Charge(ctx context.Context, idempotencyKey string, amountCents int64, currency string) (providerRef string, err error)
	Refund(ctx context.Context, providerRef string, amountCents int64) error
}

type TestStripeProvider struct {
	decline atomic.Bool
}

func NewTestStripeProvider() *TestStripeProvider { return &TestStripeProvider{} }

func (p *TestStripeProvider) SetDecline(v bool) { p.decline.Store(v) }

func (p *TestStripeProvider) Charge(ctx context.Context, idempotencyKey string, amountCents int64, currency string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if p.decline.Load() {
		return "", ErrDeclined
	}
	b := make([]byte, 12)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return "pi_test_" + hex.EncodeToString(b), nil
}

func (p *TestStripeProvider) Refund(ctx context.Context, providerRef string, amountCents int64) error {
	return ctx.Err()
}
