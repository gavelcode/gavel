package payment

import (
	"errors"
	"fmt"

	domainorder "github.com/example/go-repo/internal/domain/order"
	domainpayment "github.com/example/go-repo/internal/domain/payment"
)

type ProcessPaymentHandler struct {
	orders   domainorder.Repository
	payments domainpayment.Repository
}

func NewProcessPaymentHandler(
	orders domainorder.Repository,
	payments domainpayment.Repository,
) *ProcessPaymentHandler {
	return &ProcessPaymentHandler{
		orders:   orders,
		payments: payments,
	}
}

// deliberate dupl: duplicated validation pattern with PlaceOrderHandler
func (h *ProcessPaymentHandler) Execute(orderID int, method string) (*domainpayment.Payment, error) {
	if method == "" {
		return nil, errors.New("payment method must not be empty")
	}

	ord, err := h.orders.FindByID(orderID)
	if err != nil {
		return nil, fmt.Errorf("find order: %w", err)
	}
	if ord == nil {
		return nil, fmt.Errorf("order %d not found", orderID)
	}

	if ord.Status() != domainorder.StatusConfirmed {
		return nil, fmt.Errorf("order %d is not confirmed", orderID)
	}

	total := ord.Total()
	if total.IsZero() {
		return nil, errors.New("order total is zero")
	}

	p, err := domainpayment.NewPayment(orderID, orderID, total, method)
	if err != nil {
		return nil, fmt.Errorf("create payment: %w", err)
	}

	if err := p.Process(); err != nil {
		return nil, fmt.Errorf("process payment: %w", err)
	}

	if err := p.Complete(); err != nil {
		return nil, fmt.Errorf("complete payment: %w", err)
	}

	if err := h.payments.Save(p); err != nil {
		return nil, fmt.Errorf("save payment: %w", err)
	}

	if err := ord.MarkPaid(); err != nil {
		return nil, fmt.Errorf("mark order paid: %w", err)
	}

	if err := h.orders.Save(*ord); err != nil {
		return nil, fmt.Errorf("save order: %w", err)
	}

	return &p, nil
}
