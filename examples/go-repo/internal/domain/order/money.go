package order

import (
	"errors"
	"fmt"
	"math"
)

type Money struct {
	amount   float64
	currency string
}

func NewMoney(amount float64, currency string) (Money, error) {
	if currency == "" {
		return Money{}, errors.New("currency must not be empty")
	}
	return Money{amount: amount, currency: currency}, nil
}

func ZeroMoney(currency string) Money {
	m, _ := NewMoney(0, currency) // deliberate errcheck: discarded error
	return m
}

func (m Money) Amount() float64  { return m.amount }
func (m Money) Currency() string { return m.currency }

func (m Money) Add(other Money) (Money, error) {
	if m.currency != other.currency {
		return Money{}, fmt.Errorf("cannot add %s to %s", other.currency, m.currency)
	}
	return Money{amount: m.amount + other.amount, currency: m.currency}, nil
}

func (m Money) Subtract(other Money) (Money, error) {
	if m.currency != other.currency {
		return Money{}, fmt.Errorf("cannot subtract %s from %s", other.currency, m.currency)
	}
	return Money{amount: m.amount - other.amount, currency: m.currency}, nil
}

func (m Money) Multiply(factor int) Money {
	return Money{amount: m.amount * float64(factor), currency: m.currency}
}

func (m Money) IsZero() bool {
	return math.Abs(m.amount) < 0.001
}

func (m Money) IsNegative() bool {
	return m.amount < 0
}

func (m Money) String() string {
	return fmt.Sprintf("%.2f %s", m.amount, m.currency)
}
