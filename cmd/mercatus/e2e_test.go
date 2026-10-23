package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"mercatus/platform/config"
	"mercatus/platform/database/dbtest"
	"mercatus/seed"
)

type client struct {
	t     *testing.T
	base  string
	token string
}

func (c *client) do(method, path string, body any, out any, headers ...string) int {
	c.t.Helper()
	var buf bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&buf).Encode(body)
	}
	req, _ := http.NewRequest(method, c.base+path, &buf)
	req.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	for i := 0; i+1 < len(headers); i += 2 {
		req.Header.Set(headers[i], headers[i+1])
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		c.t.Fatal(err)
	}
	defer resp.Body.Close()
	if out != nil {
		_ = json.NewDecoder(resp.Body).Decode(out)
	}
	return resp.StatusCode
}

func setup(t *testing.T) (*App, *httptest.Server) {
	db := dbtest.New(t)
	if err := seed.Run(context.Background(), db, seed.Options{Password: "seed-password"}); err != nil {
		t.Fatal(err)
	}
	cfg, _ := config.Load()
	cfg.JWTSecret = "test-secret-that-is-long-enough-for-hs256"
	cfg.SearchURL, cfg.FraudURL = "", ""
	cfg.RateLimitRPS, cfg.RateLimitBurst = 1000, 1000
	app, err := NewApp(cfg, db)
	if err != nil {
		t.Fatal(err)
	}
	app.authRL = app.limiter
	app.ready.Store(true)
	h, err := app.Router()
	if err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	return app, srv
}

func login(t *testing.T, base, email string) *client {
	c := &client{t: t, base: base}
	var auth struct {
		AccessToken string `json:"access_token"`
	}
	if code := c.do("POST", "/api/auth/login", map[string]string{"email": email, "password": "seed-password"}, &auth); code != 200 {
		t.Fatalf("login %s: %d", email, code)
	}
	c.token = auth.AccessToken
	return c
}

func variantID(t *testing.T, c *client, slug string) int64 {
	var p struct {
		Variants []struct {
			ID int64 `json:"id"`
		} `json:"variants"`
	}
	if code := c.do("GET", "/api/catalog/products/"+slug, nil, &p); code != 200 {
		t.Fatalf("product %s: %d", slug, code)
	}
	return p.Variants[0].ID
}

func scalar(t *testing.T, app *App, q string, args ...any) int64 {
	var n int64
	if err := app.db.QueryRow(q, args...).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

func TestCheckoutEndToEnd(t *testing.T) {
	app, srv := setup(t)
	buyer := login(t, srv.URL, "buyer@example.com")
	guide := variantID(t, buyer, "seller-guide")
	if code := buyer.do("POST", "/api/cart/items", map[string]any{"variant_id": guide, "quantity": 2, "unit_price": 1}, nil); code != 200 {
		t.Fatalf("add guide %d", code)
	}
	var res struct {
		OrderID int64  `json:"order_id"`
		Status  string `json:"status"`
	}
	if code := buyer.do("POST", "/api/checkout", nil, &res, "Idempotency-Key", "e2e-checkout-1"); code != 201 || res.Status != "completed" {
		t.Fatalf("checkout %d %+v", code, res)
	}
	var again struct {
		OrderID int64 `json:"order_id"`
	}
	buyer.do("POST", "/api/checkout", nil, &again, "Idempotency-Key", "e2e-checkout-1")
	if again.OrderID != res.OrderID {
		t.Fatal("idempotent checkout replay returned a different order")
	}
	if n := scalar(t, app, `SELECT count(*) FROM orders`); n != 1 {
		t.Fatalf("orders=%d", n)
	}
	if total := scalar(t, app, `SELECT total_cents FROM orders WHERE id = $1`, res.OrderID); total != 200 {
		t.Fatalf("server must price the order, got %d", total)
	}
	if on := scalar(t, app, `SELECT on_hand FROM stock WHERE variant_id = $1`, guide); on != 99998 {
		t.Fatalf("guide stock %d", on)
	}
	if sum := scalar(t, app, `SELECT COALESCE(sum(amount_cents), 0) FROM ledger_entries`); sum != 0 {
		t.Fatalf("ledger unbalanced %d", sum)
	}
	var cart struct {
		ItemCount int `json:"item_count"`
	}
	buyer.do("GET", "/api/cart", nil, &cart)
	if cart.ItemCount != 0 {
		t.Fatal("cart should be cleared after checkout")
	}
	var unread struct {
		Unread int `json:"unread"`
	}
	buyer.do("GET", "/api/notifications/unread-count", nil, &unread)
	if unread.Unread != 1 {
		t.Fatalf("unread=%d", unread.Unread)
	}
}

func TestDeclinedPaymentCompensates(t *testing.T) {
	app, srv := setup(t)
	buyer := login(t, srv.URL, "buyer@example.com")
	v := variantID(t, buyer, "seller-guide")
	before := scalar(t, app, `SELECT on_hand - reserved FROM stock WHERE variant_id = $1`, v)
	buyer.do("POST", "/api/cart/items", map[string]any{"variant_id": v, "quantity": 3}, nil)
	app.provider.SetDecline(true)
	var res struct {
		OrderID int64  `json:"order_id"`
		Status  string `json:"status"`
		Error   string `json:"error"`
	}
	if code := buyer.do("POST", "/api/checkout", nil, &res, "Idempotency-Key", "e2e-decline-1"); code != http.StatusConflict {
		t.Fatalf("declined checkout code %d", code)
	}
	if res.Status != "failed" || res.Error != "your payment was declined" {
		t.Fatalf("result %+v", res)
	}
	if after := scalar(t, app, `SELECT on_hand - reserved FROM stock WHERE variant_id = $1`, v); after != before {
		t.Fatalf("stock not released: before=%d after=%d", before, after)
	}
	var status string
	app.db.QueryRow(`SELECT status FROM orders WHERE id = $1`, res.OrderID).Scan(&status)
	if status != "cancelled" {
		t.Fatalf("order status %s", status)
	}
	app.provider.SetDecline(false)
	if code := buyer.do("POST", "/api/checkout", nil, &res, "Idempotency-Key", "e2e-decline-2"); code != 201 {
		t.Fatalf("retry after decline should succeed, got %d", code)
	}
}

func TestSecurityBoundaries(t *testing.T) {
	_, srv := setup(t)
	anon := &client{t: t, base: srv.URL}
	if code := anon.do("POST", "/api/auth/register", map[string]string{"username": "Mallory", "email": "m@example.com", "password": "password123", "role": "admin"}, nil); code != 400 {
		t.Fatalf("self-registering as admin must fail, got %d", code)
	}
	for _, tc := range []struct{ method, path string }{
		{"PUT", "/api/stock"},
		{"POST", "/api/notifications/send"},
		{"GET", "/api/admin/stats"},
		{"GET", "/api/payments/ledger/integrity"},
		{"POST", "/api/checkout"},
	} {
		if code := anon.do(tc.method, tc.path, map[string]any{}, nil); code != 401 {
			t.Fatalf("%s %s unauthenticated: %d", tc.method, tc.path, code)
		}
	}
	buyer := login(t, srv.URL, "buyer@example.com")
	if code := buyer.do("GET", "/api/admin/stats", nil, nil); code != 403 {
		t.Fatalf("buyer on admin: %d", code)
	}
	v := variantID(t, buyer, "seller-guide")
	other := &client{t: t, base: srv.URL}
	var reg struct {
		AccessToken string `json:"access_token"`
	}
	other.do("POST", "/api/auth/register", map[string]string{"username": "Other Shop", "email": "other@example.com", "password": "password123", "role": "seller"}, &reg)
	seller := &client{t: t, base: srv.URL, token: reg.AccessToken}
	if code := seller.do("PUT", "/api/stock", map[string]any{"variant_id": v, "on_hand": 9999}, nil); code != 403 {
		t.Fatalf("seller editing another store's stock: %d", code)
	}
	forged := &client{t: t, base: srv.URL, token: "eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0.eyJ1c2VyX2lkIjoxLCJyb2xlIjoiYWRtaW4ifQ."}
	if code := forged.do("GET", "/api/admin/stats", nil, nil); code != 401 {
		t.Fatalf("alg=none token accepted: %d", code)
	}
}

func TestRefreshTokenRotationDetectsReuse(t *testing.T) {
	_, srv := setup(t)
	c := &client{t: t, base: srv.URL}
	var a struct {
		RefreshToken string `json:"refresh_token"`
	}
	c.do("POST", "/api/auth/login", map[string]string{"email": "buyer@example.com", "password": "seed-password"}, &a)
	var b struct {
		RefreshToken string `json:"refresh_token"`
	}
	if code := c.do("POST", "/api/auth/refresh", map[string]string{"refresh_token": a.RefreshToken}, &b); code != 200 {
		t.Fatalf("refresh %d", code)
	}
	if code := c.do("POST", "/api/auth/refresh", map[string]string{"refresh_token": a.RefreshToken}, nil); code != 401 {
		t.Fatalf("reused token must be rejected, got %d", code)
	}
	if code := c.do("POST", "/api/auth/refresh", map[string]string{"refresh_token": b.RefreshToken}, nil); code != 401 {
		t.Fatal(fmt.Sprintf("reuse must revoke the whole session family, got %d", code))
	}
}
