package payment_test

import (
	"testing"

	"github.com/example/go-repo/internal/domain/order"
	"github.com/example/go-repo/internal/domain/payment"
)

func TestNewPayment(t *testing.T) {
	amount, _ := order.NewMoney(50.00, "USD")
	p, err := payment.NewPayment(1, 100, amount, "credit_card")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.ID() != 1 {
		t.Errorf("expected ID 1, got %d", p.ID())
	}
	if p.Status() != payment.StatusPending {
		t.Errorf("expected pending, got %s", p.Status())
	}
}

func TestRejectZeroAmount(t *testing.T) {
	zero := order.ZeroMoney("USD")
	_, err := payment.NewPayment(1, 100, zero, "credit_card")
	if err == nil {
		t.Fatal("expected error for zero amount")
	}
}

func TestRejectEmptyMethod(t *testing.T) {
	amount, _ := order.NewMoney(50.00, "USD")
	_, err := payment.NewPayment(1, 100, amount, "")
	if err == nil {
		t.Fatal("expected error for empty method")
	}
}

func TestProcessAndComplete(t *testing.T) {
	amount, _ := order.NewMoney(50.00, "USD")
	p, _ := payment.NewPayment(1, 100, amount, "credit_card")
	if err := p.Process(); err != nil {
		t.Fatalf("process: %v", err)
	}
	if p.Status() != payment.StatusProcessing {
		t.Errorf("expected processing, got %s", p.Status())
	}
	if err := p.Complete(); err != nil {
		t.Fatalf("complete: %v", err)
	}
	if p.Status() != payment.StatusCompleted {
		t.Errorf("expected completed, got %s", p.Status())
	}
}

func TestProcessAndFail(t *testing.T) {
	amount, _ := order.NewMoney(50.00, "USD")
	p, _ := payment.NewPayment(1, 100, amount, "credit_card")
	p.Process()
	p.Fail()
	if p.Status() != payment.StatusFailed {
		t.Errorf("expected failed, got %s", p.Status())
	}
}

func TestRefundCompleted(t *testing.T) {
	amount, _ := order.NewMoney(50.00, "USD")
	p, _ := payment.NewPayment(1, 100, amount, "credit_card")
	p.Process()
	p.Complete()
	if err := p.Refund(); err != nil {
		t.Fatalf("refund: %v", err)
	}
	if p.Status() != payment.StatusRefunded {
		t.Errorf("expected refunded, got %s", p.Status())
	}
}

func TestCannotProcessNonPending(t *testing.T) {
	amount, _ := order.NewMoney(50.00, "USD")
	p, _ := payment.NewPayment(1, 100, amount, "credit_card")
	p.Process()
	if err := p.Process(); err == nil {
		t.Fatal("expected error processing non-pending payment")
	}
}

func TestCannotRefundNonCompleted(t *testing.T) {
	amount, _ := order.NewMoney(50.00, "USD")
	p, _ := payment.NewPayment(1, 100, amount, "credit_card")
	if err := p.Refund(); err == nil {
		t.Fatal("expected error refunding non-completed payment")
	}
}
