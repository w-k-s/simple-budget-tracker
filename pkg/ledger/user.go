package ledger

import (
	"fmt"
	"net/mail"
)

type UserId uint64
type User struct {
	id    UserId
	email *mail.Address
}

func NewUser(id UserId, email *mail.Address) User {
	return User{
		id:    id,
		email: email,
	}
}

func NewUserWithEmailString(id UserId, emailString string) (User, error) {
	email, err := mail.ParseAddress(emailString)
	if err != nil {
		return User{}, NewError(ErrUserEmailInvalid, err.Error(), err)
	}
	return User{
		id:    id,
		email: email,
	}, nil
}

func (u User) Id() UserId {
	return u.id
}

func (u User) Email() *mail.Address {
	return u.email
}

func (u User) String() string {
	return fmt.Sprintf("User{id: %d, email: %s}", u.id, u.email.Address)
}
