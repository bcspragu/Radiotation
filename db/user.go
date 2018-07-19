package db

import (
	"fmt"
	"strings"
)

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
	UnknownAccountType AccountType = iota
	FacebookAccount
	GoogleAccount
	NonLoggedInAccount
)

func (a AccountType) String() string {
	switch a {
	case GoogleAccount:
		return "Google"
	case FacebookAccount:
		return "Facebook"
	case NonLoggedInAccount:
		return "NonLoggedIn"
	}
	return "Unknown"
}

func accountTypeFromString(at string) (AccountType, error) {
	switch at {
	case "Google":
		return GoogleAccount, nil
	case "Facebook":
		return FacebookAccount, nil
	case "NonLoggedIn":
		return NonLoggedInAccount, nil
	default:
		return UnknownAccountType, fmt.Errorf("unrecognized account type %q", at)
	}
}

func (id UserID) String() string {
	return id.AccountType.String() + ":" + id.ID
}

func UserIDFromString(uid string) (UserID, error) {
	idp := strings.SplitN(uid, ":", 2)
	if len(idp) != 2 {
		return UserID{}, fmt.Errorf("malformed uid %q", uid)
	}
	at, err := accountTypeFromString(idp[0])
	if err != nil {
		return UserID{}, err
	}
	return UserID{
		AccountType: at,
		ID:          idp[1],
	}, nil
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
