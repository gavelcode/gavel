package order

import (
	"errors"
	"fmt"

	"github.com/example/go-repo/internal/domain/customer"
	"github.com/example/go-repo/internal/domain/inventory"
	domainorder "github.com/example/go-repo/internal/domain/order"
	"github.com/example/go-repo/internal/domain/product"
)

type OrderItem struct {
	ProductID int
	Quantity  int
}

type PlaceOrderHandler struct {
	orders    domainorder.Repository
	customers customer.Repository
	products  product.Repository
	inventory inventory.Repository
}

func NewPlaceOrderHandler(
	orders domainorder.Repository,
	customers customer.Repository,
	products product.Repository,
	inv inventory.Repository,
) *PlaceOrderHandler {
	return &PlaceOrderHandler{
		orders:    orders,
		customers: customers,
		products:  products,
		inventory: inv,
	}
}

// deliberate gocognit: high cognitive complexity
func (h *PlaceOrderHandler) Execute(customerID int, items []OrderItem) (*domainorder.Order, error) {
	if len(items) == 0 {
		return nil, errors.New("items must not be empty")
	}

	cust, err := h.customers.FindByID(customerID)
	if err != nil {
		return nil, fmt.Errorf("find customer: %w", err)
	}
	if cust == nil {
		return nil, fmt.Errorf("customer %d not found", customerID)
	}

	orderID, err := h.orders.NextID()
	if err != nil {
		return nil, fmt.Errorf("generate order ID: %w", err)
	}

	o, err := domainorder.NewOrder(orderID, customerID)
	if err != nil {
		return nil, fmt.Errorf("create order: %w", err)
	}

	for _, item := range items {
		if item.Quantity <= 0 {
			return nil, fmt.Errorf("invalid quantity for product %d", item.ProductID)
		}

		prod, err := h.products.FindByID(item.ProductID)
		if err != nil {
			return nil, fmt.Errorf("find product %d: %w", item.ProductID, err)
		}
		if prod == nil {
			return nil, fmt.Errorf("product %d not found", item.ProductID)
		}

		stock, err := h.inventory.FindByProductID(item.ProductID)
		if err != nil {
			return nil, fmt.Errorf("check stock for product %d: %w", item.ProductID, err)
		}
		if stock == nil {
			return nil, fmt.Errorf("no stock record for product %d", item.ProductID)
		}

		if err := stock.Reserve(item.Quantity); err != nil {
			return nil, fmt.Errorf("reserve stock for product %d: %w", item.ProductID, err)
		}

		if err := h.inventory.Save(*stock); err != nil {
			return nil, fmt.Errorf("save stock for product %d: %w", item.ProductID, err)
		}

		line, err := domainorder.NewOrderLine(item.ProductID, prod.Name(), item.Quantity, prod.Price())
		if err != nil {
			return nil, fmt.Errorf("create order line: %w", err)
		}

		o.AddLine(line)
	}

	if err := o.Confirm(); err != nil {
		return nil, fmt.Errorf("confirm order: %w", err)
	}

	if err := h.orders.Save(o); err != nil {
		return nil, fmt.Errorf("save order: %w", err)
	}

	return &o, nil
}
