package core

import "database/sql"

type UserDao interface {
	Close() error
	NewUserId() (UserId,error)
	
	Save(u *User) error
	SaveTx(u *User, tx *sql.Tx) error

	GetUserById(id UserId) (*User, error)
}