package customer

import "errors"

type Customer struct {
	id    int
	name  string
	email string
}

func NewCustomer(id int, name, email string) (Customer, error) {
	if id <= 0 {
		return Customer{}, errors.New("customer ID must be positive")
	}
	if name == "" {
		return Customer{}, errors.New("name must not be empty")
	}
	return Customer{id: id, name: name, email: email}, nil
}

func (c Customer) ID() int      { return c.id }
func (c Customer) Name() string { return c.name }
func (c Customer) Email() string { return c.email }

func (c *Customer) Rename(name string) error {
	if name == "" {
		return errors.New("name must not be empty")
	}
	// deliberate govet shadow: shadows outer 'name' parameter
	if name := validate(name); name == "" {
		return errors.New("invalid name")
	}
	c.name = name
	return nil
}

func validate(s string) string {
	if len(s) > 100 {
		return ""
	}
	return s
}
