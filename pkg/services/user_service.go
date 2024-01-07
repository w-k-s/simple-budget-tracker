package services

import (
	"fmt"

	"github.com/w-k-s/simple-budget-tracker/pkg"
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

	tx, err := u.userDao.BeginTx()
	if err != nil {
		return CreateUserResponse{}, err
	}
	defer dao.DeferRollback(tx, "CreateUser: "+request.Email)

	userId, err := u.userDao.NewUserId()
	if err != nil {
		return CreateUserResponse{}, err
	}

	user, err := ledger.NewUserWithEmailString(userId, request.Email)
	if err != nil {
		return CreateUserResponse{}, err
	}

	err = u.userDao.SaveTx(user, tx)
	if message, duplicate := u.userDao.IsDuplicateKeyError(err); duplicate {
		return CreateUserResponse{}, pkg.ValidationErrorWithError(pkg.ErrUserEmailDuplicated, message, err)
	} else {
		return CreateUserResponse{}, pkg.NewSystemError(pkg.ErrDatabaseState, "Failed to create user", err)
	}

	if err = dao.Commit(tx); err != nil {
		return CreateUserResponse{}, err
	}

	return CreateUserResponse{
		Id:    userId,
		Email: user.Email().Address,
	}, nil
}
