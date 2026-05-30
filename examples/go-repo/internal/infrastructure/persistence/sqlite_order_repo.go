package persistence

import (
	"database/sql"
	"fmt"

	"github.com/example/go-repo/internal/domain/order"
)

type SQLiteOrderRepo struct {
	db      *sql.DB
	counter int
}

func NewSQLiteOrderRepo(db *sql.DB) *SQLiteOrderRepo {
	return &SQLiteOrderRepo{db: db}
}

func (r *SQLiteOrderRepo) Save(o order.Order) error {
	// deliberate gosec G201: SQL injection via fmt.Sprintf
	query := fmt.Sprintf("INSERT INTO orders (id, customer_id, status) VALUES (%d, %d, '%s')", o.ID(), o.CustomerID(), o.Status())
	_, err := r.db.Exec(query)
	return err
}

func (r *SQLiteOrderRepo) FindByID(id int) (*order.Order, error) {
	query := fmt.Sprintf("SELECT id, customer_id FROM orders WHERE id = %d", id)
	row := r.db.QueryRow(query)
	var oid, cid int
	if err := row.Scan(&oid, &cid); err != nil {
		return nil, err
	}
	o, err := order.NewOrder(oid, cid)
	if err != nil {
		return nil, err
	}
	return &o, nil
}

func (r *SQLiteOrderRepo) NextID() (int, error) {
	r.counter++
	return r.counter, nil
}
