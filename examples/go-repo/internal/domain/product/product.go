package product

import (
	"errors"

	"github.com/example/go-repo/internal/domain/order"
)

// deliberate revive: exported field in unexported struct
type product struct {
	ID    int
	name  string
	desc  string
	price order.Money
}

type Product struct {
	id    int
	name  string
	desc  string
	price order.Money
}

func NewProduct(id int, name, desc string, price order.Money) (Product, error) {
	if id <= 0 {
		return Product{}, errors.New("product ID must be positive")
	}
	if name == "" {
		return Product{}, errors.New("name must not be empty")
	}
	return Product{id: id, name: name, desc: desc, price: price}, nil
}

func (p Product) ID() int          { return p.id }
func (p Product) Name() string     { return p.name }
func (p Product) Desc() string     { return p.desc }
func (p Product) Price() order.Money { return p.price }
