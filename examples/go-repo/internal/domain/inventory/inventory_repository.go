package inventory

type Repository interface {
	Save(stock Stock) error
	FindByProductID(productID int) (*Stock, error)
}
