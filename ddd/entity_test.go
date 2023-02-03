package ddd

type IPerson Entity[Person, string]

type Person struct {
	id string
}

func (p Person) SameIdentityAs(other Person) bool {
	return p.Identity() == other.Identity()
}

func (p Person) Identity() string {
	return p.id
}
