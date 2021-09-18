package services

import (
	"context"

	"github.com/w-k-s/simple-budget-tracker/pkg/ledger"
)

type ContextKey string

const (
	CtxUserId    ContextKey = "userId"
	CtxAccountId ContextKey = "accountId"
)

func RequireUserId(ctx context.Context) (ledger.UserId, error) {
	var (
		userId ledger.UserId
		ok     bool
	)
	if userId, ok = ctx.Value(CtxUserId).(ledger.UserId); !ok {
		return 0, ledger.NewError(ledger.ErrServiceUserIdRequired, "User id is required", nil)
	}
	return userId, nil
}

func SetAccountId(ctx context.Context, accountId ledger.AccountId) context.Context {
	return context.WithValue(ctx, CtxAccountId, accountId)
}

func RequireAccountId(ctx context.Context) (ledger.AccountId, error) {
	var (
		accountId ledger.AccountId
		ok        bool
	)
	if accountId, ok = ctx.Value(CtxAccountId).(ledger.AccountId); !ok {
		return 0, ledger.NewError(ledger.ErrServiceUserIdRequired, "User id is required", nil)
	}
	return accountId, nil
}
