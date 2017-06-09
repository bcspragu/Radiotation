package app

type AccountType int

func (a AccountType) String() string {
	switch a {
	case GoogleAccount:
		return "Google"
	case FacebookAccount:
		return "Facebook"
	}
	return "Unknown"
}

const (
	GoogleAccount AccountType = iota
	FacebookAccount
)

type ID struct {
	AccountType AccountType
	ID          string
}

func (id ID) String() string {
	return id.AccountType.String() + ":" + id.ID
}

type User struct {
	ID          ID
	First, Last string
}

func GoogleUser(id, first, last string) *User {
	return newUser(ID{AccountType: GoogleAccount, ID: id}, first, last)
}

func newUser(id ID, first, last string) *User {
	return &User{
		ID:    id,
		First: first,
		Last:  last,
	}
}
