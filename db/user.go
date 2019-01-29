package db

type (
	UserID string

	User struct {
		ID          UserID
		First, Last string
	}
)

func newUser(id UserID, first, last string) *User {
	return &User{
		ID:    id,
		First: first,
		Last:  last,
	}
}
