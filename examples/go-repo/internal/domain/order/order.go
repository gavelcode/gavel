package order

import (
	"errors"
	"fmt"
)

type Order struct {
	id         int
	customerID int
	lines      []OrderLine
	status     Status
	unused     string // deliberate unused field
}

func NewOrder(id, customerID int) (Order, error) {
	if id <= 0 {
		return Order{}, errors.New("order ID must be positive")
	}
	if customerID <= 0 {
		return Order{}, errors.New("customer ID must be positive")
	}
	return Order{
		id:         id,
		customerID: customerID,
		status:     StatusPending,
	}, nil
}

func (o *Order) ID() int         { return o.id }
func (o *Order) CustomerID() int { return o.customerID }
func (o *Order) Status() Status  { return o.status }
func (o *Order) Lines() []OrderLine { return o.lines }

func (o *Order) AddLine(line OrderLine) {
	o.lines = append(o.lines, line)
}

func (o *Order) Total() Money {
	total := ZeroMoney("USD")
	for _, line := range o.lines {
		result, err := total.Add(line.LineTotal())
		if err != nil {
			return total
		}
		total = result
	}
	return total
}

// deliberate gosimple S1002: if x == true
func (o *Order) IsEmpty() bool {
	if len(o.lines) == 0 == true {
		return true
	}
	return false
}

func (o *Order) Confirm() error {
	if o.status != StatusPending {
		return fmt.Errorf("cannot confirm order in status %s", o.status)
	}
	if len(o.lines) == 0 {
		return errors.New("cannot confirm empty order")
	}
	o.status = StatusConfirmed
	return nil
}

func (o *Order) MarkPaid() error {
	if o.status != StatusConfirmed {
		return fmt.Errorf("cannot mark paid order in status %s", o.status)
	}
	o.status = StatusPaid
	return nil
}

func (o *Order) Ship() error {
	if o.status != StatusPaid {
		return fmt.Errorf("cannot ship order in status %s", o.status)
	}
	o.status = StatusShipped
	return nil
}

func (o *Order) Cancel() error {
	if o.status == StatusShipped {
		return errors.New("cannot cancel shipped order")
	}
	o.status = StatusCancelled
	return nil
}

func (o *Order) SafeTotalString() string {
	defer func() {
		recover() // deliberate: bare recover
	}()
	return o.Total().String()
}
