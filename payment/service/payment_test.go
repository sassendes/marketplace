package service

import (
	"context"
	"errors"
	"sync"
	"testing"

	"marketplace/payment/model"
	"marketplace/payment/repository"
	"marketplace/platform/database/dbtest"
)

func charge() *model.ChargeRequest {
	return &model.ChargeRequest{
		IdempotencyKey: "order-1", OrderID: 1, AmountCents: 4500, Currency: "USD",
		Splits: []model.SellerSplit{
			{SellerID: 10, SubtotalCents: 4000, CommissionCents: 400},
			{SellerID: 20, SubtotalCents: 500, CommissionCents: 50},
		},
	}
}

func TestConcurrentChargesWithSameKeyChargeOnce(t *testing.T) {
	db := dbtest.New(t)
	svc := NewPaymentService(repository.NewPaymentRepository(db), NewTestStripeProvider())
	ctx := context.Background()
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := svc.Charge(ctx, 5, charge())
			if err != nil && !errors.Is(err, ErrInProgress) {
				t.Error(err)
			}
		}()
	}
	wg.Wait()
	var payments, entries int
	db.QueryRow(`SELECT count(*) FROM payments`).Scan(&payments)
	db.QueryRow(`SELECT count(*) FROM ledger_entries`).Scan(&entries)
	if payments != 1 || entries != 4 {
		t.Fatalf("payments=%d entries=%d", payments, entries)
	}
	total, _ := svc.LedgerTotal(ctx)
	if total != 0 {
		t.Fatalf("ledger sum %d", total)
	}
	if bal, _ := svc.LedgerBalance(ctx, "seller:10"); bal != 3600 {
		t.Fatalf("seller balance %d", bal)
	}
}

func TestKeyReuseWithDifferentAmountIsRejected(t *testing.T) {
	db := dbtest.New(t)
	svc := NewPaymentService(repository.NewPaymentRepository(db), NewTestStripeProvider())
	ctx := context.Background()
	if _, err := svc.Charge(ctx, 5, charge()); err != nil {
		t.Fatal(err)
	}
	req := charge()
	req.AmountCents = 1
	req.Splits = []model.SellerSplit{{SellerID: 10, SubtotalCents: 1}}
	if _, err := svc.Charge(ctx, 5, req); !errors.Is(err, ErrKeyReuse) {
		t.Fatalf("want key reuse error, got %v", err)
	}
}

func TestRefundReversesLedgerAndIsIdempotent(t *testing.T) {
	db := dbtest.New(t)
	svc := NewPaymentService(repository.NewPaymentRepository(db), NewTestStripeProvider())
	ctx := context.Background()
	p, err := svc.Charge(ctx, 5, charge())
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		if err := svc.RefundOrder(ctx, p.OrderID); err != nil {
			t.Fatal(err)
		}
	}
	var entries int
	db.QueryRow(`SELECT count(*) FROM ledger_entries`).Scan(&entries)
	if entries != 8 {
		t.Fatalf("entries=%d want 8", entries)
	}
	if bal, _ := svc.LedgerBalance(ctx, "seller:10"); bal != 0 {
		t.Fatalf("seller balance after refund %d", bal)
	}
}

func TestDeclinedChargeFailsPayment(t *testing.T) {
	db := dbtest.New(t)
	prov := NewTestStripeProvider()
	prov.SetDecline(true)
	svc := NewPaymentService(repository.NewPaymentRepository(db), prov)
	p, err := svc.Charge(context.Background(), 5, charge())
	if !errors.Is(err, ErrPaymentFailed) || p.Status != model.StatusFailed {
		t.Fatalf("p=%+v err=%v", p, err)
	}
	var entries int
	db.QueryRow(`SELECT count(*) FROM ledger_entries`).Scan(&entries)
	if entries != 0 {
		t.Fatal("declined charge must not write ledger entries")
	}
}
