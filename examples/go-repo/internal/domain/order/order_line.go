package order

import "errors"

type OrderLine struct {
	productID   int
	productName string
	quantity    int
	unitPrice   Money
}

func NewOrderLine(productID int, productName string, quantity int, unitPrice Money) (OrderLine, error) {
	if productID <= 0 {
		return OrderLine{}, errors.New("product ID must be positive")
	}
	if quantity <= 0 {
		return OrderLine{}, errors.New("quantity must be positive")
	}
	return OrderLine{
		productID:   productID,
		productName: productName,
		quantity:    quantity,
		unitPrice:   unitPrice,
	}, nil
}

func (l OrderLine) ProductID() int     { return l.productID }
func (l OrderLine) ProductName() string { return l.productName }
func (l OrderLine) Quantity() int      { return l.quantity }
func (l OrderLine) UnitPrice() Money   { return l.unitPrice }

func (l OrderLine) LineTotal() Money {
	return l.unitPrice.Multiply(l.quantity)
}
