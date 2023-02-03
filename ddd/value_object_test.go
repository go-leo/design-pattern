package ddd

type IAddress ValueObject[Address]

type Address struct {
	country  string
	province string
	city     string
}

func (a Address) SameValueAs(other Address) bool {
	return a.country == other.country && a.province == other.country && a.city == other.city
}
