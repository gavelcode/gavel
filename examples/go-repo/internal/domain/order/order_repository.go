package order

type Repository interface {
	Save(order Order) error
	FindByID(id int) (*Order, error)
	NextID() (int, error)
}
