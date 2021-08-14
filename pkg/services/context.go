package services

import (
	"context"

	"github.com/w-k-s/simple-budget-tracker/pkg/ledger"
)

type ContextKey string

const (
	CtxUserId ContextKey = "userId"
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
