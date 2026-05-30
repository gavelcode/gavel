package payment

type Repository interface {
	Save(payment Payment) error
	FindByID(id int) (*Payment, error)
}
