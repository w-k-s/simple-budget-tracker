package core

import (
	"net/mail"
)

type UserId int64
type User struct {
	id    UserId
	email *mail.Address
}

func NewUser(id UserId, email *mail.Address) *User {
	return &User{
		id:    id,
		email: email,
	}
}

func (u User) Id() UserId {
	return u.id
}

func (u User) Email() *mail.Address {
	return u.email
}
