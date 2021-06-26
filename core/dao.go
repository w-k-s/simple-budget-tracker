package core

import "database/sql"

type UserDao interface {
	Close() error
	NewUserId() (UserId, error)

	Save(u *User) error
	SaveTx(u *User, tx *sql.Tx) error

	GetUserById(id UserId) (*User, error)
}

type AccountDao interface {
	Close() error
	NewAccountId() (AccountId, error)

	Save(id UserId, a *Account) error
	SaveTx(id UserId, a *Account, tx *sql.Tx) error

	GetAccountById(id AccountId) (*Account, error)
	GetAccountsByUserId(id UserId) ([]*Account, error)
}

type CategoryDao interface {
	Close() error
	NewCategoryId() (CategoryId, error)

	Save(id UserId, c Categories) error
	SaveTx(id UserId, c Categories, tx *sql.Tx) error

	GetCategoriesForUser(id UserId) (Categories, error)
}
