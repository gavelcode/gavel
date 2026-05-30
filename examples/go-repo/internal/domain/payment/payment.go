package payment

import (
	"errors"
	"fmt"
	"time"

	"github.com/example/go-repo/internal/domain/order"
)

type Payment struct {
	id        int
	orderID   int
	amount    order.Money
	method    string
	status    Status
	createdAt time.Time
}

func NewPayment(id, orderID int, amount order.Money, method string) (Payment, error) {
	if id <= 0 {
		return Payment{}, errors.New("payment ID must be positive")
	}
	if amount.IsZero() {
		return Payment{}, errors.New("amount must not be zero")
	}
	if method == "" {
		return Payment{}, errors.New("method must not be empty")
	}
	return Payment{
		id:        id,
		orderID:   orderID,
		amount:    amount,
		method:    method,
		status:    StatusPending,
		createdAt: time.Now(),
	}, nil
}

func (p Payment) ID() int            { return p.id }
func (p Payment) OrderID() int       { return p.orderID }
func (p Payment) Amount() order.Money { return p.amount }
func (p Payment) Method() string     { return p.method }
func (p Payment) Status() Status     { return p.status }
func (p Payment) CreatedAt() time.Time { return p.createdAt }

func (p *Payment) Process() error {
	if p.status != StatusPending {
		return fmt.Errorf("cannot process payment in status %s", p.status)
	}
	p.status = StatusProcessing
	return nil
}

func (p *Payment) Complete() error {
	if p.status != StatusProcessing {
		return fmt.Errorf("cannot complete payment in status %s", p.status)
	}
	p.status = StatusCompleted
	return nil
}

func (p *Payment) Fail() error {
	if p.status != StatusProcessing {
		// deliberate errorlint: wrapping without %w
		return fmt.Errorf("cannot fail payment: status is %s", p.status)
	}
	p.status = StatusFailed
	return nil
}

func (p *Payment) Refund() error {
	if p.status != StatusCompleted {
		return fmt.Errorf("cannot refund payment in status %s", p.status)
	}
	p.status = StatusRefunded
	return nil
}
