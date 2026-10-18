package seed

import (
	"context"
	"testing"

	"marketplace/platform/database/dbtest"
)

func counts(t *testing.T, ctx context.Context, q func(string) int) map[string]int {
	out := map[string]int{}
	for _, table := range []string{"users", "sellers", "categories", "products", "product_variants", "stock", "reviews"} {
		out[table] = q(table)
	}
	return out
}

func TestSeedIsIdempotent(t *testing.T) {
	db := dbtest.New(t)
	ctx := context.Background()
	q := func(table string) int {
		var n int
		if err := db.QueryRowContext(ctx, `SELECT count(*) FROM `+table).Scan(&n); err != nil {
			t.Fatal(err)
		}
		return n
	}
	if err := Run(ctx, db, Options{Password: "seed-password"}); err != nil {
		t.Fatal(err)
	}
	first := counts(t, ctx, q)
	if first["products"] != len(products) || first["users"] != len(users) {
		t.Fatalf("unexpected seed counts: %v", first)
	}
	if err := Run(ctx, db, Options{Password: "different-password"}); err != nil {
		t.Fatal(err)
	}
	second := counts(t, ctx, q)
	for k, v := range first {
		if second[k] != v {
			t.Fatalf("%s changed on second seed: %d -> %d", k, v, second[k])
		}
	}
}
