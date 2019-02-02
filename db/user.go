package db

type (
	UserID string

	User struct {
		ID    UserID `json:"id"`
		First string `json:"first"`
		Last  string `json:"last"`
	}
)

func newUser(id UserID, first, last string) *User {
	return &User{
		ID:    id,
		First: first,
		Last:  last,
	}
}
