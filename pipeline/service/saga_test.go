package service

import (
	"context"
	"slices"
	"testing"

	"marketplace/pipeline/repository"
	"marketplace/platform/database/dbtest"
)

func TestSagaHappyPath(t *testing.T) {
	db := dbtest.New(t)
	svc := NewSagaService(repository.NewSagaRepository(db))
	saga, err := svc.NewSaga(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	var order []string
	mk := func(name string) func(context.Context) error {
		return func(context.Context) error { order = append(order, name); return nil }
	}
	steps := BuildCheckoutSteps(Downstream{
		Reserve: mk("do:reserve"), Release: mk("undo:release"),
		Charge: mk("do:charge"), Refund: mk("undo:refund"),
		Confirm: mk("do:confirm"),
	}, "")
	if err := svc.Run(context.Background(), saga, steps); err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if !slices.Equal(order, []string{"do:reserve", "do:charge", "do:confirm"}) {
		t.Fatalf("wrong order: %v", order)
	}
	var st string
	db.QueryRow(`SELECT status FROM sagas WHERE id = $1`, saga.ID).Scan(&st)
	if st != "completed" {
		t.Fatalf("saga status = %s, want completed", st)
	}
}

func TestSagaCompensatesInReverse(t *testing.T) {
	db := dbtest.New(t)
	svc := NewSagaService(repository.NewSagaRepository(db))
	saga, err := svc.NewSaga(context.Background(), 2)
	if err != nil {
		t.Fatal(err)
	}
	var order []string
	mk := func(name string) func(context.Context) error {
		return func(context.Context) error { order = append(order, name); return nil }
	}
	steps := BuildCheckoutSteps(Downstream{
		Reserve: mk("do:reserve"), Release: mk("undo:release"),
		Charge: mk("do:charge"), Refund: mk("undo:refund"),
		Confirm: mk("do:confirm"),
	}, "confirm")
	if err := svc.Run(context.Background(), saga, steps); err == nil {
		t.Fatal("expected failure")
	}
	if !slices.Equal(order, []string{"do:reserve", "do:charge", "undo:refund", "undo:release"}) {
		t.Fatalf("compensation order wrong: %v", order)
	}
	var st, reserve string
	db.QueryRow(`SELECT status FROM sagas WHERE id = $1`, saga.ID).Scan(&st)
	db.QueryRow(`SELECT status FROM saga_steps WHERE saga_id = $1 AND name = 'reserve'`, saga.ID).Scan(&reserve)
	if st != "failed" || reserve != "compensated" {
		t.Fatalf("saga=%s reserve=%s", st, reserve)
	}
}
