package services

import (
	"context"
	"fmt"
	"log"

	"github.com/w-k-s/simple-budget-tracker/pkg"
	"github.com/w-k-s/simple-budget-tracker/pkg/ledger"
	dao "github.com/w-k-s/simple-budget-tracker/pkg/persistence"
)

type CategoryBudgetRequest struct {
	CategoryId uint64         `json:"categoryId"`
	MaxAmount  AmountResponse `json:"maxAmount"`
}

type CreateBudgetRequest struct {
	AccountIds      []uint64                `json:"accountIds"`
	PeriodType      string                  `json:"period"`
	CategoryBudgets []CategoryBudgetRequest `json:"categoryBudgets"`
}

type BudgetResponse struct {
	Id              uint64                  `json:"id"`
	AccountIds      []uint64                `json:"accountIds"`
	PeriodType      string                  `json:"period"`
	CategoryBudgets []CategoryBudgetRequest `json:"categoryBudgets"`
}

type BudgetService interface {
	CreateBudget(ctx context.Context, request CreateBudgetRequest) (BudgetResponse, error)
	GetBudget(ctx context.Context, budgetId ledger.BudgetId) (BudgetResponse, error)
}

type budgetService struct {
	uniqueIdService UniqueIdService
	accountDao      dao.AccountDao
	categoryDao     dao.CategoryDao
	daoFactory      dao.Factory
}

func NewBudgetService(
	uniqueIdService UniqueIdService,
	accountDao dao.AccountDao,
	categoryDao dao.CategoryDao,
	daoFactory dao.Factory,
) (BudgetService, error) {
	if uniqueIdService == nil {
		log.Fatalf("can not create budget service. uniqueIdService is nil")
	}
	if categoryDao == nil {
		log.Fatalf("can not create budget service. categoryDao is nil")
	}
	if daoFactory == nil {
		log.Fatalf("can not create budget service. daoFactory is nil")
	}

	return &budgetService{
		uniqueIdService: uniqueIdService,
		daoFactory:      daoFactory,
	}, nil
}

func (svc budgetService) CreateBudget(ctx context.Context, request CreateBudgetRequest) (BudgetResponse, error) {

	userId, err := RequireUserId(ctx)
	if err != nil {
		return BudgetResponse{}, err
	}

	tx, err := svc.daoFactory.BeginTx()
	if err != nil {
		return BudgetResponse{}, pkg.NewSystemError(pkg.ErrDatabaseState, "Failed to begin transaction", err)
	}

	defer dao.DeferRollback(tx, fmt.Sprintf("CreateBudget: %d", userId))

	accountIds := uint64ToAccountIds(request.AccountIds)

	// Ensure that the currency of the accounts = currency of the budget.
	// Currencies of the budget should be the same (validated later), so we'll just take the first one
	if len(request.CategoryBudgets) > 0 {
		budgetCurrency := request.CategoryBudgets[0].MaxAmount.Currency
		accountCurrenciesMap, err := svc.accountDao.GetCurrenciesOfAccounts(ctx, accountIds, userId, tx)
		if err != nil {
			return BudgetResponse{}, pkg.ValidationErrorWithError(pkg.ErrDatabaseState, "Budget currency must match account currencies", nil)
		}

		for _, currency := range accountCurrenciesMap {
			if currency.CurrencyCode() != budgetCurrency {
				return BudgetResponse{}, pkg.ValidationErrorWithError(pkg.ErrDatabaseState, "Budget currency must match account currencies", nil)
			}
		}
	}

	// Ensure the budget categories belong to the user.
	categories, err := svc.categoryDao.GetCategoriesForUser(ctx, userId, tx)
	if err != nil {
		return BudgetResponse{}, err
	}

	categoryIdMap := categories.MapById()
	categoryBudgets := ledger.CategoryBudgets{}
	for _, budget := range request.CategoryBudgets {
		if category, ok := categoryIdMap[ledger.CategoryId(budget.CategoryId)]; ok {
			limit, err := ledger.NewMoney(budget.MaxAmount.Currency, budget.MaxAmount.Value)
			if err != nil {
				return BudgetResponse{}, err
			}
			categoryBudget, err := ledger.NewCategoryBudget(
				category.Id(),
				limit,
			)
			categoryBudgets = append(categoryBudgets, categoryBudget)
		}
	}

	budget, err := ledger.NewBudget(
		ledger.BudgetId(svc.uniqueIdService.MustGetId(EntityBudget)),
		accountIds,
		ledger.BudgetPeriodType(request.PeriodType),
		nil,
		ledger.MustMakeUpdatedByUserId(userId),
	)
	if err != nil {
		return BudgetResponse{}, err
	}

	budgetDao := svc.daoFactory.GetBudgetDao(tx)
	err = budgetDao.Save(ctx, userId, budget)
	if err != nil {
		return BudgetResponse{}, err
	}

	if err = dao.Commit(tx); err != nil {
		return BudgetResponse{}, err
	}

	return BudgetResponse{
		Id:              uint64(budget.Id()),
		AccountIds:      request.AccountIds,
		PeriodType:      request.PeriodType,
		CategoryBudgets: request.CategoryBudgets,
	}, nil
}

func (svc budgetService) GetBudget(ctx context.Context, budgetId ledger.BudgetId) (BudgetResponse, error) {

	userId, err := RequireUserId(ctx)
	if err != nil {
		return BudgetResponse{}, err
	}

	tx, err := svc.daoFactory.BeginTx()
	if err != nil {
		return BudgetResponse{}, pkg.NewSystemError(pkg.ErrDatabaseState, "Failed to begin transaction", err)
	}

	defer dao.DeferRollback(tx, fmt.Sprintf("GetBudget: %d", userId))

	budgetDao := svc.daoFactory.GetBudgetDao(tx)
	budget, err := budgetDao.GetBudgetById(ctx, budgetId, userId)
	if err != nil {
		return BudgetResponse{}, err
	}

	categoryBudgets := []CategoryBudgetRequest{}
	for _, categoryBudget := range budget.CategoryBudgets() {
		categoryBudgets = append(categoryBudgets, CategoryBudgetRequest{
			CategoryId: uint64(categoryBudget.CategoryId()),
			MaxAmount: AmountResponse{
				Currency: categoryBudget.MaxLimit().Currency().CurrencyCode(),
				Value:    categoryBudget.MaxLimit().MustMinorUnits(),
			},
		})
	}

	return BudgetResponse{
		Id:              uint64(budget.Id()),
		AccountIds:      accountIdsToUint64(budget.AccountIds()),
		PeriodType:      string(budget.PeriodType()),
		CategoryBudgets: categoryBudgets,
	}, err
}

func uint64ToAccountIds(ids []uint64) ledger.AccountIds {
	accountIds := ledger.AccountIds{}
	for _, accountId := range ids {
		accountIds = append(accountIds, ledger.AccountId(accountId))
	}
	return accountIds
}

func accountIdsToUint64(accountIds ledger.AccountIds) []uint64 {
	ids := []uint64{}
	for _, accountId := range accountIds {
		ids = append(ids, uint64(accountId))
	}
	return ids
}
