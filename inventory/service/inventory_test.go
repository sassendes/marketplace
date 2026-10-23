package service

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"

	"mercatus/inventory/model"
	"mercatus/inventory/repository"
	"mercatus/platform/database/dbtest"
)

func TestNoOversellUnderConcurrency(t *testing.T) {
	db := dbtest.New(t)
	svc := NewInventoryService(repository.NewInventoryRepository(db))
	ctx := context.Background()
	if err := svc.SetStock(ctx, 1, 1); err != nil {
		t.Fatal(err)
	}
	var ok, rejected atomic.Int32
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(order int64) {
			defer wg.Done()
			err := svc.ReserveOrder(ctx, order, []model.Line{{VariantID: 1, Quantity: 1}})
			switch {
			case err == nil:
				ok.Add(1)
			case errors.Is(err, ErrInsufficientStock):
				rejected.Add(1)
			default:
				t.Error(err)
			}
		}(int64(100 + i))
	}
	wg.Wait()
	if ok.Load() != 1 || rejected.Load() != 19 {
		t.Fatalf("ok=%d rejected=%d, want 1 and 19", ok.Load(), rejected.Load())
	}
	st, _ := svc.GetStock(ctx, 1)
	if st.Reserved != 1 || st.Available() != 0 {
		t.Fatalf("stock %+v", st)
	}
}

func TestMultiLineReservationIsAllOrNothing(t *testing.T) {
	db := dbtest.New(t)
	svc := NewInventoryService(repository.NewInventoryRepository(db))
	ctx := context.Background()
	_ = svc.SetStock(ctx, 1, 5)
	_ = svc.SetStock(ctx, 2, 1)
	err := svc.ReserveOrder(ctx, 7, []model.Line{{VariantID: 1, Quantity: 2}, {VariantID: 2, Quantity: 3}})
	if !errors.Is(err, ErrInsufficientStock) {
		t.Fatalf("want insufficient stock, got %v", err)
	}
	a, _ := svc.GetStock(ctx, 1)
	if a.Reserved != 0 {
		t.Fatalf("partial reservation leaked: %+v", a)
	}
}

func TestCommitAndReleaseAreIdempotent(t *testing.T) {
	db := dbtest.New(t)
	svc := NewInventoryService(repository.NewInventoryRepository(db))
	ctx := context.Background()
	_ = svc.SetStock(ctx, 1, 5)
	if err := svc.ReserveOrder(ctx, 9, []model.Line{{VariantID: 1, Quantity: 2}}); err != nil {
		t.Fatal(err)
	}
	if err := svc.ReserveOrder(ctx, 9, []model.Line{{VariantID: 1, Quantity: 2}}); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		if err := svc.Commit(ctx, 9); err != nil {
			t.Fatal(err)
		}
	}
	st, _ := svc.GetStock(ctx, 1)
	if st.OnHand != 3 || st.Reserved != 0 {
		t.Fatalf("after commit %+v", st)
	}
	if err := svc.ReleaseOrder(ctx, 9); err != nil {
		t.Fatal(err)
	}
	st, _ = svc.GetStock(ctx, 1)
	if st.OnHand != 3 || st.Reserved != 0 {
		t.Fatalf("release after commit must be a no-op: %+v", st)
	}
	if err := svc.ReserveOrder(ctx, 10, []model.Line{{VariantID: 1, Quantity: 1}}); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`UPDATE reservations SET expires_at = now() - interval '1 minute' WHERE order_id = 10`); err != nil {
		t.Fatal(err)
	}
	n, err := svc.ExpireStale(ctx)
	if err != nil || n != 1 {
		t.Fatalf("expire n=%d err=%v", n, err)
	}
	if err := svc.Commit(ctx, 10); !errors.Is(err, ErrReservationGone) {
		t.Fatalf("commit after expiry should fail, got %v", err)
	}
}
