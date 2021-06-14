package core

type UserDao interface {
	Save(u *User)
}
