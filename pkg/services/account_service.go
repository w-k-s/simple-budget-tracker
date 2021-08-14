package services

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/w-k-s/simple-budget-tracker/pkg/ledger"
	dao "github.com/w-k-s/simple-budget-tracker/pkg/persistence"
)

type CreateAccountsRequest struct{
	Accounts []struct{
		Name string `json:"name"`
		Currency string `json:"currency"`
	} `json:"accounts"`
}

type AccountResponse struct{
	Id uint64 `json:"id"`
	Name string `json:"name"`
	Currency string `json:"currency"`
}

type AccountsResponse struct{
	Accounts []AccountResponse `json:"accounts"`
}

type AccountService interface {
	CreateAccounts(ctx context.Context, request CreateAccountsRequest) (AccountsResponse, error)
	GetAccounts(ctx context.Context) (AccountsResponse, error)
}

type accountService struct {
	accountDao dao.AccountDao
}

func NewAccountService(accountDao dao.AccountDao) (AccountService, error) {
	if accountDao == nil {
		return nil, fmt.Errorf("can not create user service. accountDao is nil")
	}

	return &accountService{
		accountDao: accountDao,
	}, nil
}

func (svc accountService) CreateAccounts(ctx context.Context, request CreateAccountsRequest) (AccountsResponse, error){
	var (
		userId ledger.UserId
		tx *sql.Tx
		err error
	)

	if userId, err = RequireUserId(ctx); err != nil{
		return AccountsResponse{},err.(ledger.Error)
	}

	if tx, err = svc.accountDao.BeginTx(); err != nil{
		return AccountsResponse{},err.(ledger.Error)
	}

	defer dao.DeferRollback(tx, fmt.Sprintf("CreateAccounts: %d", userId))

	// Create Account models
	var accounts ledger.Accounts
	for _, accountReq := range request.Accounts{
		var(
			accountId ledger.AccountId
			account ledger.Account
			err error
		)
		// TODO: limit number of accounts that can be created
		if accountId, err = svc.accountDao.NewAccountId(tx); err != nil{
			return AccountsResponse{},err.(ledger.Error)
		}

		if account, err = ledger.NewAccount(
			accountId, 
			accountReq.Name, 
			accountReq.Currency, 
			ledger.MustMakeUpdatedByUserId(userId),
		); err != nil{
			return AccountsResponse{}, err.(ledger.Error)
		}

		accounts = append(accounts, account)
	}

	// Save Accounts
	if err = svc.accountDao.SaveTx(ctx, userId, accounts, tx); err != nil{
		return AccountsResponse{}, err.(ledger.Error)
	}

	if err = dao.Commit(tx); err != nil{
		return AccountsResponse{}, err.(ledger.Error)
	}

	// Return response
	response := AccountsResponse{}
	for _, account := range accounts{
		response.Accounts = append(response.Accounts, AccountResponse{
			Id: uint64(account.Id()),
			Name: account.Name(),
			Currency: account.Currency(),
		})
	}

	return response, nil
}

func (svc accountService) GetAccounts(ctx context.Context) (AccountsResponse, error){
	var (
		userId ledger.UserId
		tx *sql.Tx
		accounts []ledger.Account
		err error
	)

	if userId, err = RequireUserId(ctx); err != nil{
		return AccountsResponse{},err.(ledger.Error)
	}

	if tx, err = svc.accountDao.BeginTx(); err != nil{
		return AccountsResponse{},err.(ledger.Error)
	}

	defer dao.DeferRollback(tx, fmt.Sprintf("GetAccounts: %d", userId))

	if accounts, err = svc.accountDao.GetAccountsByUserId(ctx, userId, tx); err != nil{
		return AccountsResponse{}, err.(ledger.Error)
	}

	if err = dao.Commit(tx); err != nil{
		return AccountsResponse{}, err.(ledger.Error)
	}

	response := AccountsResponse{}
	for _, account := range accounts{
		response.Accounts = append(response.Accounts, AccountResponse{
			Id: uint64(account.Id()),
			Name: account.Name(),
			Currency: account.Currency(),
		})
	}

	return response, nil
}