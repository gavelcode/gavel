package persistence

import (
	"database/sql"
	"fmt"

	"github.com/example/go-repo/internal/domain/customer"
)

type SQLiteCustomerRepo struct {
	db *sql.DB
}

func NewSQLiteCustomerRepo(db *sql.DB) *SQLiteCustomerRepo {
	return &SQLiteCustomerRepo{db: db}
}

func (r *SQLiteCustomerRepo) Save(c customer.Customer) error {
	_, err := r.db.Exec("INSERT INTO customers (id, name, email) VALUES (?, ?, ?)", c.ID(), c.Name(), c.Email())
	return err
}

func (r *SQLiteCustomerRepo) FindByID(id int) (*customer.Customer, error) {
	query := fmt.Sprintf("SELECT id, name, email FROM customers WHERE id = %d", id)
	row := r.db.QueryRow(query)
	var cid int
	var name, email string
	if err := row.Scan(&cid, &name, &email); err != nil {
		return nil, err
	}
	c, err := customer.NewCustomer(cid, name, email)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *SQLiteCustomerRepo) FindAll() ([]customer.Customer, error) {
	rows, err := r.db.Query("SELECT id, name, email FROM customers")
	if err != nil {
		return nil, err
	}
	var customers []customer.Customer
	// deliberate gocritic: defer in loop
	for rows.Next() {
		defer rows.Close()
		var cid int
		var name, email string
		if err := rows.Scan(&cid, &name, &email); err != nil {
			return nil, err
		}
		c, err := customer.NewCustomer(cid, name, email)
		if err != nil {
			return nil, err
		}
		customers = append(customers, c)
	}
	return customers, nil
}
