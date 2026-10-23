package seed

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	authservice "mercatus/auth/service"
)

type Options struct {
	Password string
}

type Result struct {
	Users, Sellers, Categories, Products, Variants, Reviews int
}

func Run(ctx context.Context, db *sql.DB, opts Options) error {
	password := opts.Password
	generated := false
	if password == "" {
		b := make([]byte, 12)
		if _, err := rand.Read(b); err != nil {
			return err
		}
		password = base64.RawURLEncoding.EncodeToString(b)
		generated = true
	}
	if len(password) < 8 {
		return errors.New("SEED_PASSWORD must be at least 8 characters")
	}
	hash, err := authservice.HashPassword(password)
	if err != nil {
		return err
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(727274003)`); err != nil {
		return err
	}

	var res Result
	userIDs := map[string]int64{}
	for _, u := range users {
		r, err := tx.ExecContext(ctx,
			`INSERT INTO users (username, email, password_hash, role, email_verified, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, true, now() - interval '30 days', now()) ON CONFLICT (email) DO NOTHING`,
			u.Name, u.Email, hash, u.Role)
		if err != nil {
			return fmt.Errorf("seed user %s: %w", u.Email, err)
		}
		res.Users += affected(r)
		var id int64
		if err := tx.QueryRowContext(ctx, `SELECT id FROM users WHERE email = $1`, u.Email).Scan(&id); err != nil {
			return err
		}
		userIDs[u.Email] = id
	}

	sellerIDs := map[string]int64{}
	for i, s := range sellers {
		r, err := tx.ExecContext(ctx,
			`INSERT INTO sellers (user_id, store_name, slug, description, status, commission_rate, stripe_account_id, payout_enabled, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, 'active', 1000, $5, true, now(), now()) ON CONFLICT (user_id) DO NOTHING`,
			userIDs[s.Email], s.Store, s.Slug, s.Description, fmt.Sprintf("acct_seed%08d", i+1))
		if err != nil {
			return fmt.Errorf("seed seller %s: %w", s.Slug, err)
		}
		res.Sellers += affected(r)
		var id int64
		if err := tx.QueryRowContext(ctx, `SELECT id FROM sellers WHERE user_id = $1`, userIDs[s.Email]).Scan(&id); err != nil {
			return err
		}
		sellerIDs[s.Email] = id
	}

	categoryIDs := map[string]int64{}
	for _, c := range categories {
		var parent any
		if c.Parent != "" {
			parent = categoryIDs[c.Parent]
		}
		r, err := tx.ExecContext(ctx,
			`INSERT INTO categories (parent_id, name, slug, created_at) VALUES ($1, $2, $3, now()) ON CONFLICT (slug) DO NOTHING`,
			parent, c.Name, c.Slug)
		if err != nil {
			return fmt.Errorf("seed category %s: %w", c.Slug, err)
		}
		res.Categories += affected(r)
		var id int64
		if err := tx.QueryRowContext(ctx, `SELECT id FROM categories WHERE slug = $1`, c.Slug).Scan(&id); err != nil {
			return err
		}
		categoryIDs[c.Slug] = id
	}

	productIDs := map[string]int64{}
	for i, p := range products {
		r, err := tx.ExecContext(ctx,
			`INSERT INTO products (seller_id, title, slug, description, category_id, brand, image_url, status, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, $6, '', 'active', now() - make_interval(hours => $7), now())
			 ON CONFLICT (slug) DO NOTHING`,
			sellerIDs[p.Seller], p.Title, p.Slug, p.Description, categoryIDs[p.Category], p.Brand, i*7)
		if err != nil {
			return fmt.Errorf("seed product %s: %w", p.Slug, err)
		}
		res.Products += affected(r)
		var id int64
		if err := tx.QueryRowContext(ctx, `SELECT id FROM products WHERE slug = $1`, p.Slug).Scan(&id); err != nil {
			return err
		}
		productIDs[p.Slug] = id

		for _, vr := range p.Variants {
			attrs := vr.Attrs
			if attrs == nil {
				attrs = map[string]any{}
			}
			b, _ := json.Marshal(attrs)
			var vid int64
			err := tx.QueryRowContext(ctx,
				`INSERT INTO product_variants (product_id, sku, price, currency, stock, attributes, active, created_at, updated_at)
				 VALUES ($1, $2, $3, 'USD', $4, $5, true, now(), now())
				 ON CONFLICT (product_id, sku) DO NOTHING RETURNING id`,
				id, vr.SKU, vr.Price, vr.Stock, string(b)).Scan(&vid)
			if errors.Is(err, sql.ErrNoRows) {
				continue
			}
			if err != nil {
				return fmt.Errorf("seed variant %s: %w", vr.SKU, err)
			}
			res.Variants++
			if _, err := tx.ExecContext(ctx,
				`INSERT INTO stock (variant_id, on_hand, reserved, updated_at) VALUES ($1, $2, 0, now())
				 ON CONFLICT (variant_id) DO NOTHING`, vid, vr.Stock); err != nil {
				return err
			}
		}
	}

	for _, rv := range reviews {
		r, err := tx.ExecContext(ctx,
			`INSERT INTO reviews (product_id, user_id, rating, title, body, verified_purchase, helpful, created_at)
			 VALUES ($1, $2, $3, $4, $5, false, 0, now() - interval '3 days') ON CONFLICT (product_id, user_id) DO NOTHING`,
			productIDs[rv.Product], userIDs[rv.Author], rv.Rating, rv.Title, rv.Body)
		if err != nil {
			return fmt.Errorf("seed review: %w", err)
		}
		res.Reviews += affected(r)
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	slog.Info("seed complete", "users", res.Users, "sellers", res.Sellers, "categories", res.Categories,
		"products", res.Products, "variants", res.Variants, "reviews", res.Reviews)
	if generated && res.Users > 0 {
		slog.Warn("seed accounts were created with a generated password, set SEED_PASSWORD to choose one",
			"password", password, "admin", "admin@example.com", "buyer", "buyer@example.com")
	}
	return nil
}

func affected(r sql.Result) int {
	n, _ := r.RowsAffected()
	return int(n)
}
