package core

type UserDao interface {
	Close() error
	NewUserId() (UserId,error)
	Save(u *User)
	GetUserById(id UserId) (*User, error)
}