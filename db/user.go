package db

type (
	AccountType int

	UserID struct {
		AccountType AccountType
		ID          string
	}

	User struct {
		ID          UserID
		First, Last string
	}
)

const (
	GoogleAccount AccountType = iota
	FacebookAccount
)

func (a AccountType) String() string {
	switch a {
	case GoogleAccount:
		return "Google"
	case FacebookAccount:
		return "Facebook"
	}
	return "Unknown"
}

func (id UserID) String() string {
	return id.AccountType.String() + ":" + id.ID
}

func GoogleUser(id, first, last string) *User {
	return newUser(UserID{AccountType: GoogleAccount, ID: id}, first, last)
}

func newUser(id UserID, first, last string) *User {
	return &User{
		ID:    id,
		First: first,
		Last:  last,
	}
}
