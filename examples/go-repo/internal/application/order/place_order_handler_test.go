package order_test

import (
	"testing"

	apporder "github.com/example/go-repo/internal/application/order"
	"github.com/example/go-repo/internal/domain/customer"
	"github.com/example/go-repo/internal/domain/inventory"
	domainorder "github.com/example/go-repo/internal/domain/order"
	"github.com/example/go-repo/internal/domain/product"
)

type inMemoryOrderRepo struct {
	orders  map[int]domainorder.Order
	counter int
}

func (r *inMemoryOrderRepo) Save(o domainorder.Order) error {
	r.orders[o.ID()] = o
	return nil
}
func (r *inMemoryOrderRepo) FindByID(id int) (*domainorder.Order, error) {
	o, ok := r.orders[id]
	if !ok {
		return nil, nil
	}
	return &o, nil
}
func (r *inMemoryOrderRepo) NextID() (int, error) {
	r.counter++
	return r.counter, nil
}

type inMemoryCustomerRepo struct {
	customers map[int]customer.Customer
}

func (r *inMemoryCustomerRepo) Save(c customer.Customer) error {
	r.customers[c.ID()] = c
	return nil
}
func (r *inMemoryCustomerRepo) FindByID(id int) (*customer.Customer, error) {
	c, ok := r.customers[id]
	if !ok {
		return nil, nil
	}
	return &c, nil
}

type inMemoryProductRepo struct {
	products map[int]product.Product
}

func (r *inMemoryProductRepo) FindByID(id int) (*product.Product, error) {
	p, ok := r.products[id]
	if !ok {
		return nil, nil
	}
	return &p, nil
}
func (r *inMemoryProductRepo) Add(p product.Product) error {
	r.products[p.ID()] = p
	return nil
}

type inMemoryInventoryRepo struct {
	stocks map[int]inventory.Stock
}

func (r *inMemoryInventoryRepo) Save(s inventory.Stock) error {
	r.stocks[s.ProductID()] = s
	return nil
}
func (r *inMemoryInventoryRepo) FindByProductID(id int) (*inventory.Stock, error) {
	s, ok := r.stocks[id]
	if !ok {
		return nil, nil
	}
	return &s, nil
}

func setupHandler() *apporder.PlaceOrderHandler {
	orders := &inMemoryOrderRepo{orders: make(map[int]domainorder.Order)}
	customers := &inMemoryCustomerRepo{customers: make(map[int]customer.Customer)}
	products := &inMemoryProductRepo{products: make(map[int]product.Product)}
	inv := &inMemoryInventoryRepo{stocks: make(map[int]inventory.Stock)}

	c, _ := customer.NewCustomer(1, "Alice", "alice@example.com")
	customers.Save(c)

	laptopPrice, _ := domainorder.NewMoney(999.99, "USD")
	mousePrice, _ := domainorder.NewMoney(29.99, "USD")
	laptop, _ := product.NewProduct(1, "Laptop", "Gaming laptop", laptopPrice)
	mouse, _ := product.NewProduct(2, "Mouse", "Wireless mouse", mousePrice)
	products.Add(laptop)
	products.Add(mouse)

	s1, _ := inventory.NewStock(1, 10)
	s2, _ := inventory.NewStock(2, 50)
	inv.Save(s1)
	inv.Save(s2)

	return apporder.NewPlaceOrderHandler(orders, customers, products, inv)
}

func TestPlaceOrderSuccessfully(t *testing.T) {
	handler := setupHandler()
	items := []apporder.OrderItem{
		{ProductID: 1, Quantity: 2},
		{ProductID: 2, Quantity: 3},
	}
	result, err := handler.Execute(1, items)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.CustomerID() != 1 {
		t.Errorf("expected customer ID 1, got %d", result.CustomerID())
	}
	if result.Status() != domainorder.StatusConfirmed {
		t.Errorf("expected confirmed, got %s", result.Status())
	}
	if len(result.Lines()) != 2 {
		t.Errorf("expected 2 lines, got %d", len(result.Lines()))
	}
}

func TestRejectUnknownCustomer(t *testing.T) {
	handler := setupHandler()
	_, err := handler.Execute(999, []apporder.OrderItem{{ProductID: 1, Quantity: 1}})
	if err == nil {
		t.Fatal("expected error for unknown customer")
	}
}

func TestRejectEmptyItems(t *testing.T) {
	handler := setupHandler()
	_, err := handler.Execute(1, []apporder.OrderItem{})
	if err == nil {
		t.Fatal("expected error for empty items")
	}
}

func TestRejectInsufficientStock(t *testing.T) {
	handler := setupHandler()
	_, err := handler.Execute(1, []apporder.OrderItem{{ProductID: 1, Quantity: 100}})
	if err == nil {
		t.Fatal("expected error for insufficient stock")
	}
}
