package services

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/w-k-s/simple-budget-tracker/pkg/ledger"
	dao "github.com/w-k-s/simple-budget-tracker/pkg/persistence"
)

type CreateCategoriesRequest struct {
	Categories []struct {
		Name string `json:"name"`
	} `json:"categories"`
}

type CategoryResponse struct {
	Id   uint64 `json:"id"`
	Name string `json:"name"`
}

type CategoriesResponse struct {
	Categories []CategoryResponse `json:"categories"`
}

type CategoriesService interface {
	CreateCategories(ctx context.Context, request CreateCategoriesRequest) (CategoriesResponse, error)
	GetCategories(ctx context.Context) (CategoriesResponse, error)
}

type categoriesService struct {
	categoryDao dao.CategoryDao
}

func NewCategoriesService(categoryDao dao.CategoryDao) (CategoriesService, error) {
	if categoryDao == nil {
		return nil, fmt.Errorf("can not create category service. categoryDao is nil")
	}

	return &categoriesService{
		categoryDao: categoryDao,
	}, nil
}

func (svc categoriesService) CreateCategories(ctx context.Context, request CreateCategoriesRequest) (CategoriesResponse, error) {
	var (
		userId ledger.UserId
		tx     *sql.Tx
		err    error
	)

	if userId, err = RequireUserId(ctx); err != nil {
		return CategoriesResponse{}, err.(ledger.Error)
	}

	if tx, err = svc.categoryDao.BeginTx(); err != nil {
		return CategoriesResponse{}, err.(ledger.Error)
	}

	defer dao.DeferRollback(tx, fmt.Sprintf("CreateCategories: %d", userId))

	// Create Categories models
	var categories ledger.Categories
	for _, categoryReq := range request.Categories {
		var (
			categoryId ledger.CategoryId
			category   ledger.Category
			err        error
		)
		// TODO: limit number of accounts that can be created
		if categoryId, err = svc.categoryDao.NewCategoryId(tx); err != nil {
			return CategoriesResponse{}, err.(ledger.Error)
		}

		if category, err = ledger.NewCategory(
			categoryId,
			categoryReq.Name,
			ledger.MustMakeUpdatedByUserId(userId),
		); err != nil {
			return CategoriesResponse{}, err.(ledger.Error)
		}

		categories = append(categories, category)
	}

	// Save Categories
	if err = svc.categoryDao.SaveTx(ctx, userId, categories, tx); err != nil {
		return CategoriesResponse{}, err.(ledger.Error)
	}

	if err = dao.Commit(tx); err != nil {
		return CategoriesResponse{}, err.(ledger.Error)
	}

	// Return response
	response := CategoriesResponse{}
	for _, category := range categories {
		response.Categories = append(response.Categories, CategoryResponse{
			Id:   uint64(category.Id()),
			Name: category.Name(),
		})
	}

	return response, nil
}

func (svc categoriesService) GetCategories(ctx context.Context) (CategoriesResponse, error) {
	var (
		tx         *sql.Tx
		userId     ledger.UserId
		categories ledger.Categories
		err        error
	)

	if userId, err = RequireUserId(ctx); err != nil {
		return CategoriesResponse{}, err.(ledger.Error)
	}

	if tx, err = svc.categoryDao.BeginTx(); err != nil {
		return CategoriesResponse{}, err.(ledger.Error)
	}

	defer dao.DeferRollback(tx, fmt.Sprintf("GetCategories: %d", userId))

	if categories, err = svc.categoryDao.GetCategoriesForUser(ctx, userId, tx); err != nil {
		return CategoriesResponse{}, err.(ledger.Error)
	}

	if err = dao.Commit(tx); err != nil {
		return CategoriesResponse{}, err.(ledger.Error)
	}

	response := CategoriesResponse{}
	for _, category := range categories {
		response.Categories = append(response.Categories, CategoryResponse{
			Id:   uint64(category.Id()),
			Name: category.Name(),
		})
	}

	return response, nil
}
