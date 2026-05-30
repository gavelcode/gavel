package inventory

import (
	"errors"
	"fmt"
)

type Stock struct {
	productID int
	quantity  int
}

func NewStock(productID, quantity int) (Stock, error) {
	if productID <= 0 {
		return Stock{}, errors.New("product ID must be positive")
	}
	if quantity < 0 {
		return Stock{}, errors.New("quantity must not be negative")
	}
	return Stock{productID: productID, quantity: quantity}, nil
}

func (s Stock) ProductID() int { return s.productID }
func (s Stock) Quantity() int  { return s.quantity }

func (s *Stock) Reserve(qty int) error {
	if qty <= 0 {
		return errors.New("reserve quantity must be positive")
	}
	if qty > s.quantity {
		return fmt.Errorf("insufficient stock: have %d, need %d", s.quantity, qty)
	}
	s.quantity -= qty
	return nil
}

func (s *Stock) Restock(qty int) error {
	if qty <= 0 {
		return errors.New("restock quantity must be positive")
	}
	s.quantity += qty
	return nil
}
