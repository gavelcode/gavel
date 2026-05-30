package customer

type Repository interface {
	Save(customer Customer) error
	FindByID(id int) (*Customer, error)
}
