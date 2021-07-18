package core

import (
	"database/sql"
	"fmt"
)

type CreateUserRequest struct {
	Email string `json:"email"`
}

type CreateUserResponse struct {
	Id    UserId `json:"id"`
	Email string `json:"emai"`
}

type UserService interface {
	CreateUser(request CreateUserRequest) (*CreateUserResponse, error)
}

type userService struct {
	userDao UserDao
}

func NewUserService(userDao UserDao) (UserService, error) {
	if userDao == nil {
		return nil, fmt.Errorf("can not create user service. userDao is nil")
	}

	return &userService{
		userDao: userDao,
	}, nil
}

func (u userService) CreateUser(request CreateUserRequest) (*CreateUserResponse, error) {
	var (
		tx     *sql.Tx
		userId UserId
		user   *User
		err    error
	)
	if tx, err = u.userDao.BeginTx(); err != nil {
		return nil, err.(Error)
	}
	defer DeferRollback(tx, "CreateUser: "+request.Email)

	if userId, err = u.userDao.NewUserId(); err != nil {
		return nil, err.(Error)
	}
	if user, err = NewUserWithEmailString(userId, request.Email); err != nil {
		return nil, err.(Error)
	}
	if err = u.userDao.SaveTx(user, tx); err != nil {
		return nil, err.(Error)
	}
	if err = Commit(tx); err != nil {
		return nil, err.(Error)
	}
	return &CreateUserResponse{
		Id:    userId,
		Email: user.Email().Address,
	}, nil
}
