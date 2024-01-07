package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/w-k-s/simple-budget-tracker/pkg"
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

	userId, err := RequireUserId(ctx)
	if err != nil {
		return CategoriesResponse{}, err
	}

	tx, err := svc.categoryDao.BeginTx()
	if err != nil {
		return CategoriesResponse{}, err
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
			return CategoriesResponse{}, err
		}

		if category, err = ledger.NewCategory(
			categoryId,
			categoryReq.Name,
			ledger.MustMakeUpdatedByUserId(userId),
		); err != nil {
			return CategoriesResponse{}, err
		}

		categories = append(categories, category)
	}

	// Save Categories
	err = svc.categoryDao.SaveTx(ctx, userId, categories, tx)
	if _, duplicate := svc.categoryDao.IsDuplicateKeyError(err); duplicate {
		message := fmt.Sprintf("Category names must be unique. One of these is duplicated: %s", strings.Join(categories.Names(), ", "))
		if categories.Len() == 1 {
			message = fmt.Sprintf("Category named %q already exists", categories.Names()[0])
		}
		return CategoriesResponse{}, pkg.ValidationErrorWithError(pkg.ErrCategoryNameDuplicated, message, err)
	} else if err != nil {
		return CategoriesResponse{}, pkg.NewSystemError(pkg.ErrDatabaseState, "Failed to create category", err)
	}

	if err = dao.Commit(tx); err != nil {
		return CategoriesResponse{}, err
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

	userId, err := RequireUserId(ctx)
	if err != nil {
		return CategoriesResponse{}, err
	}

	tx, err := svc.categoryDao.BeginTx()
	if err != nil {
		return CategoriesResponse{}, err
	}

	defer dao.DeferRollback(tx, fmt.Sprintf("GetCategories: %d", userId))

	categories, err := svc.categoryDao.GetCategoriesForUser(ctx, userId, tx)
	if err != nil {
		return CategoriesResponse{}, err
	}

	if err = dao.Commit(tx); err != nil {
		return CategoriesResponse{}, err
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
