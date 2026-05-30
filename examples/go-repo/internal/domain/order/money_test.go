package order_test

import (
	"testing"

	"github.com/example/go-repo/internal/domain/order"
)

func TestNewMoney(t *testing.T) {
	m, err := order.NewMoney(10.50, "USD")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.Amount() != 10.50 {
		t.Errorf("expected 10.50, got %f", m.Amount())
	}
	if m.Currency() != "USD" {
		t.Errorf("expected USD, got %s", m.Currency())
	}
}

func TestZeroMoney(t *testing.T) {
	z := order.ZeroMoney("EUR")
	if !z.IsZero() {
		t.Error("expected zero")
	}
	if z.IsNegative() {
		t.Error("zero should not be negative")
	}
}

func TestMoneyAdd(t *testing.T) {
	a, _ := order.NewMoney(10.00, "USD")
	b, _ := order.NewMoney(5.50, "USD")
	result, err := a.Add(b)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Amount() != 15.50 {
		t.Errorf("expected 15.50, got %f", result.Amount())
	}
}

func TestMoneySubtract(t *testing.T) {
	a, _ := order.NewMoney(10.00, "USD")
	b, _ := order.NewMoney(3.00, "USD")
	result, _ := a.Subtract(b)
	if result.Amount() != 7.00 {
		t.Errorf("expected 7.00, got %f", result.Amount())
	}
}

func TestMoneyMultiply(t *testing.T) {
	m, _ := order.NewMoney(29.99, "USD")
	result := m.Multiply(3)
	expected := 89.97
	if result.Amount() != expected {
		t.Errorf("expected %f, got %f", expected, result.Amount())
	}
}

func TestMoneyRejectDifferentCurrencies(t *testing.T) {
	usd, _ := order.NewMoney(10.00, "USD")
	eur, _ := order.NewMoney(5.00, "EUR")
	_, err := usd.Add(eur)
	if err == nil {
		t.Fatal("expected error adding different currencies")
	}
}

func TestMoneyNegativeDetection(t *testing.T) {
	a, _ := order.NewMoney(5.00, "USD")
	b, _ := order.NewMoney(10.00, "USD")
	diff, _ := a.Subtract(b)
	if !diff.IsNegative() {
		t.Error("expected negative result")
	}
}
