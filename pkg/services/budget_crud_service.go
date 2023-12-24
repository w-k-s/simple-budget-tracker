package services

import (
	"context"
	"fmt"
	"log"

	"github.com/w-k-s/simple-budget-tracker/pkg/ledger"
	dao "github.com/w-k-s/simple-budget-tracker/pkg/persistence"
)

type CategoryBudgetRequest struct{
	CategoryId uint64 `json:"categoryId"`
	MaxAmount  AmountResponse `json:"maxAmount"`
	AmountSpent AmountResponse `json:"amountSpent"`
}

type CreateBudgetRequest struct {
	AccountIds     []uint64 `json:"accountIds"`
	PeriodType     string `json:"period"`
	CategoryBudgets []CategoryBudgetRequest`json:"categoryBudgets"`
}

type BudgetResponse struct {
	Id       uint64 `json:"id"`
	AccountIds     []uint64 `json:"accountIds"`
	PeriodType     string `json:"period"`
	CategoryBudgets []CategoryBudgetRequest`json:"categoryBudgets"`
}


type BudgetService interface {
	CreateBudget(ctx context.Context, request CreateBudgetRequest) (BudgetResponse, error)
	GetBudget(ctx context.Context, budgetId ledger.BudgetId) (BudgetResponse, error)
}

type budgetService struct {
	uniqueIdService UniqueIdService
	categoryDao dao.CategoryDao
	budgetDao dao.BudgetDao
}
 
func NewBudgetService(
	uniqueIdService UniqueIdService,
	categoryDao dao.CategoryDao,
	budgetDao dao.BudgetDao,
) (BudgetService, error) {
	if uniqueIdService == nil {
		log.Fatalf("can not create budget service. uniqueIdService is nil")
	}
	if categoryDao == nil {
		log.Fatalf("can not create budget service. categoryDao is nil")
	}
	if budgetDao == nil {
		log.Fatalf("can not create budget service. budgetDao is nil")
	}

	return &budgetService{
		uniqueIdService: uniqueIdService,
		budgetDao: budgetDao,
	}, nil
}

func (svc budgetService) CreateBudget(ctx context.Context, request CreateBudgetRequest) (BudgetResponse, error) {

	userId, err := RequireUserId(ctx); 
	if err != nil {
		return BudgetResponse{}, err
	}

	tx, err := svc.budgetDao.BeginTx(); 
	if err != nil {
		return BudgetResponse{}, err
	}

	defer dao.DeferRollback(tx, fmt.Sprintf("CreateBudget: %d", userId))

	accountIds := uint64ToAccountIds(request.AccountIds)

	categories, err := svc.categoryDao.GetCategoriesForUser(ctx, userId, tx)
	if err != nil{
		return BudgetResponse{}, err
	}

	categoryIdMap := categories.MayById()
	categoryBudgets := ledger.CategoryBudgets{}
	for _, budget := range request.CategoryBudgets{
		if category,ok := categoryIdMap[ledger.CategoryId(budget.CategoryId)]; ok{
			limit, err := ledger.NewMoney(budget.MaxAmount.Currency, budget.MaxAmount.Value)
			if err != nil{
				return BudgetResponse{}, err
			}
			zero,_ := ledger.NewMoney(budget.MaxAmount.Currency, 0)
			categoryBudget, err := ledger.NewCategoryBudget(
				category.Id(),
				limit,
				zero,
			)
			categoryBudgets = append(categoryBudgets, categoryBudget)
		}	
	}

	budget,err := ledger.NewBudget(
		ledger.BudgetId(svc.uniqueIdService.MustGetId()),
		accountIds,
		ledger.BudgetPeriodType(request.PeriodType),
		nil,
		ledger.MustMakeUpdatedByUserId(userId),
	)
	if err != nil{
		return BudgetResponse{}, err
	}

	err = svc.budgetDao.SaveTx(ctx, userId, budget, tx)
	if err != nil{
		return BudgetResponse{}, err
	}

	if err = dao.Commit(tx); err != nil{
		return BudgetResponse{}, err
	}
	
	return BudgetResponse{
		Id: uint64(budget.Id()),
		AccountIds: request.AccountIds,
		PeriodType: request.PeriodType,
		CategoryBudgets: request.CategoryBudgets,
	},nil
}

func (svc budgetService) GetBudget(ctx context.Context, budgetId ledger.BudgetId) (BudgetResponse, error) {

	userId, err := RequireUserId(ctx)
	if err != nil {
		return BudgetResponse{}, err
	}

	tx, err := svc.budgetDao.BeginTx()
	if err != nil {
		return BudgetResponse{}, err
	}

	defer dao.DeferRollback(tx, fmt.Sprintf("GetBudget: %d", userId))

	budget,err := svc.budgetDao.GetBudgetById(ctx, budgetId, userId, tx)
	if err != nil{
		return BudgetResponse{}, err
	}

	categoryBudgets := []CategoryBudgetRequest{}
	for _,categoryBudget := range budget.CategoryBudgets(){
		categoryBudgets = append(categoryBudgets, CategoryBudgetRequest{
			CategoryId: uint64(categoryBudget.CategoryId()),
			MaxAmount: AmountResponse{
				Currency: categoryBudget.MaxLimit().Currency().CurrencyCode(),
				Value: categoryBudget.MaxLimit().MustMinorUnits(),
			},
			AmountSpent: AmountResponse{
				Currency: categoryBudget.AmountSpent().Currency().CurrencyCode(),
				Value: categoryBudget.AmountSpent().MustMinorUnits(),
			},
		})
	}

	return BudgetResponse{
		Id: uint64(budget.Id()),
		AccountIds: accountIdsToUint64(budget.AccountIds()),
		PeriodType: string(budget.PeriodType()),
		CategoryBudgets: categoryBudgets,
	},err
}

func uint64ToAccountIds(ids []uint64) ledger.AccountIds{
	accountIds := ledger.AccountIds{}
	for _,accountId := range ids{
		accountIds = append(accountIds, ledger.AccountId(accountId))
	}
	return accountIds
}

func accountIdsToUint64(accountIds ledger.AccountIds) []uint64{
	ids := []uint64{}
	for _,accountId := range accountIds{
		ids = append(ids, uint64(accountId))
	}
	return ids
}