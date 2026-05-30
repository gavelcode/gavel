package product

type Repository interface {
	FindByID(id int) (*Product, error)
	Add(product Product) error
}
