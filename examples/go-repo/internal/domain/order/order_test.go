package order_test

import (
	"testing"

	"github.com/example/go-repo/internal/domain/order"
)

func TestNewOrder(t *testing.T) {
	o, err := order.NewOrder(1, 100)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if o.ID() != 1 {
		t.Errorf("expected ID 1, got %d", o.ID())
	}
	if o.CustomerID() != 100 {
		t.Errorf("expected customer ID 100, got %d", o.CustomerID())
	}
	if o.Status() != order.StatusPending {
		t.Errorf("expected status pending, got %s", o.Status())
	}
}

func TestNewOrderRejectsZeroID(t *testing.T) {
	_, err := order.NewOrder(0, 100)
	if err == nil {
		t.Fatal("expected error for zero ID")
	}
}

func TestNewOrderRejectsZeroCustomerID(t *testing.T) {
	_, err := order.NewOrder(1, 0)
	if err == nil {
		t.Fatal("expected error for zero customer ID")
	}
}

func TestOrderAddLineAndTotal(t *testing.T) {
	o, _ := order.NewOrder(1, 100)
	price, _ := order.NewMoney(29.99, "USD")
	line, _ := order.NewOrderLine(1, "Mouse", 3, price)
	o.AddLine(line)

	total := o.Total()
	if total.IsZero() {
		t.Error("expected non-zero total")
	}
}

func TestOrderConfirm(t *testing.T) {
	o, _ := order.NewOrder(1, 100)
	price, _ := order.NewMoney(999.99, "USD")
	line, _ := order.NewOrderLine(1, "Laptop", 1, price)
	o.AddLine(line)

	if err := o.Confirm(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if o.Status() != order.StatusConfirmed {
		t.Errorf("expected confirmed, got %s", o.Status())
	}
}

func TestOrderCannotConfirmTwice(t *testing.T) {
	o, _ := order.NewOrder(1, 100)
	price, _ := order.NewMoney(999.99, "USD")
	line, _ := order.NewOrderLine(1, "Laptop", 1, price)
	o.AddLine(line)
	o.Confirm()

	if err := o.Confirm(); err == nil {
		t.Fatal("expected error confirming already confirmed order")
	}
}

func TestOrderMarkPaid(t *testing.T) {
	o, _ := order.NewOrder(1, 100)
	price, _ := order.NewMoney(999.99, "USD")
	line, _ := order.NewOrderLine(1, "Laptop", 1, price)
	o.AddLine(line)
	o.Confirm()
	o.MarkPaid()

	if o.Status() != order.StatusPaid {
		t.Errorf("expected paid, got %s", o.Status())
	}
}

func TestOrderCancel(t *testing.T) {
	o, _ := order.NewOrder(1, 100)
	o.Cancel()
	if o.Status() != order.StatusCancelled {
		t.Errorf("expected cancelled, got %s", o.Status())
	}
}

func TestEmptyOrderTotalIsZero(t *testing.T) {
	o, _ := order.NewOrder(1, 100)
	if !o.Total().IsZero() {
		t.Error("expected zero total for empty order")
	}
}
