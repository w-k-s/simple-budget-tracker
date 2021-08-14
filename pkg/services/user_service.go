package services

import (
	"database/sql"
	"fmt"

	"github.com/w-k-s/simple-budget-tracker/pkg/ledger"
	dao "github.com/w-k-s/simple-budget-tracker/pkg/persistence"
)

type CreateUserRequest struct {
	Email string `json:"email"`
}

type CreateUserResponse struct {
	Id    ledger.UserId `json:"id"`
	Email string        `json:"email"`
}

type UserService interface {
	CreateUser(request CreateUserRequest) (CreateUserResponse, error)
}

type userService struct {
	userDao dao.UserDao
}

func NewUserService(userDao dao.UserDao) (UserService, error) {
	if userDao == nil {
		return nil, fmt.Errorf("can not create user service. userDao is nil")
	}

	return &userService{
		userDao: userDao,
	}, nil
}

func (u userService) CreateUser(request CreateUserRequest) (CreateUserResponse, error) {
	var (
		tx     *sql.Tx
		userId ledger.UserId
		user   ledger.User
		err    error
	)
	if tx, err = u.userDao.BeginTx(); err != nil {
		return CreateUserResponse{}, err.(ledger.Error)
	}
	defer dao.DeferRollback(tx, "CreateUser: "+request.Email)

	if userId, err = u.userDao.NewUserId(); err != nil {
		return CreateUserResponse{}, err.(ledger.Error)
	}
	if user, err = ledger.NewUserWithEmailString(userId, request.Email); err != nil {
		return CreateUserResponse{}, err.(ledger.Error)
	}
	if err = u.userDao.SaveTx(user, tx); err != nil {
		return CreateUserResponse{}, err.(ledger.Error)
	}
	if err = dao.Commit(tx); err != nil {
		return CreateUserResponse{}, err.(ledger.Error)
	}
	return CreateUserResponse{
		Id:    userId,
		Email: user.Email().Address,
	}, nil
}
