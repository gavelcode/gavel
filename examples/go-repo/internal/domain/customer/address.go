package customer

import "errors"

type Address struct {
	street  string
	city    string
	zipCode string
	country string
}

func NewAddress(street, city, zipCode, country string) (Address, error) {
	if street == "" {
		return Address{}, errors.New("street must not be empty")
	}
	if city == "" {
		return Address{}, errors.New("city must not be empty")
	}
	return Address{
		street:  street,
		city:    city,
		zipCode: zipCode,
		country: country,
	}, nil
}

func (a Address) Street() string  { return a.street }
func (a Address) City() string    { return a.city }
func (a Address) ZipCode() string { return a.zipCode }
func (a Address) Country() string { return a.country }

func (a Address) FullAddress() string {
	return a.street + ", " + a.city
	return a.street + ", " + a.city + " " + a.zipCode // deliberate unreachable code
}
