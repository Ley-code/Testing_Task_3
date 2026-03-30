package integration_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "github.com/lib/pq"

	"github.com/example/order-processing-service/testing/setup"
)

// Order is a minimal aggregate for tests.
type Order struct {
	ID          string
	CustomerID  string
	Status      string
	AmountCents int
}

// OrderRepo exercises real SQL against Postgres (integration).
type OrderRepo struct {
	db *sql.DB
}

func NewOrderRepo(db *sql.DB) *OrderRepo { return &OrderRepo{db: db} }

// CreateMut inserts a row; validates SQL and serialization.
func (r *OrderRepo) CreateMut(ctx context.Context, o *Order) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO orders (id, customer_id, status, amount_cents)
		VALUES ($1, $2, $3, $4)`,
		o.ID, o.CustomerID, o.Status, o.AmountCents,
	)
	return err
}

// UpdateMut updates only columns marked dirty (partial update).
func (r *OrderRepo) UpdateMut(ctx context.Context, o *Order, dirty map[string]bool) error {
	if len(dirty) == 0 {
		return nil
	}
	// Only update requested fields — avoids overwriting unchanged columns.
	if dirty["status"] {
		_, err := r.db.ExecContext(ctx, `UPDATE orders SET status = $1 WHERE id = $2`, o.Status, o.ID)
		return err
	}
	return nil
}

func TestOrderRepo_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in -short mode")
	}
	ctx := context.Background()
	pc, err := setup.StartPostgres(ctx)
	if err != nil {
		t.Skipf("integration: postgres container (need Docker): %v", err)
	}
	defer func() { _ = pc.Container.Terminate(ctx) }()

	db, err := sql.Open("postgres", pc.DSN)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(5)
	db.SetConnMaxLifetime(2 * time.Minute)

	_, err = db.ExecContext(ctx, `
		CREATE TABLE orders (
			id TEXT PRIMARY KEY,
			customer_id TEXT NOT NULL,
			status TEXT NOT NULL,
			amount_cents INT NOT NULL
		)`)
	if err != nil {
		t.Fatalf("schema: %v", err)
	}

	repo := NewOrderRepo(db)
	o := &Order{ID: "ord_1", CustomerID: "cust_a", Status: "pending", AmountCents: 999}
	if err := repo.CreateMut(ctx, o); err != nil {
		t.Fatalf("CreateMut: %v", err)
	}

	var got Order
	row := db.QueryRowContext(ctx, `SELECT id, customer_id, status, amount_cents FROM orders WHERE id = $1`, "ord_1")
	if err := row.Scan(&got.ID, &got.CustomerID, &got.Status, &got.AmountCents); err != nil {
		t.Fatal(err)
	}
	if got != *o {
		t.Fatalf("db row mismatch: %+v want %+v", got, o)
	}

	o.Status = "paid"
	if err := repo.UpdateMut(ctx, o, map[string]bool{"status": true}); err != nil {
		t.Fatalf("UpdateMut: %v", err)
	}
	row = db.QueryRowContext(ctx, `SELECT status FROM orders WHERE id = $1`, "ord_1")
	var st string
	if err := row.Scan(&st); err != nil {
		t.Fatal(err)
	}
	if st != "paid" {
		t.Fatalf("status: %q want paid", st)
	}

	// Dirty map empty: no-op should not error
	if err := repo.UpdateMut(ctx, o, nil); err != nil {
		t.Fatal(err)
	}
}
