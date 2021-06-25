package core

import (
	"fmt"
	"strings"

	"github.com/gobuffalo/validate"
	"github.com/gobuffalo/validate/validators"
)

type CategoryId uint64
type CategoryType string

const (
	Income  CategoryType = "INCOME"
	Expense CategoryType = "EXPENSE"
)

type Category struct {
	id           CategoryId
	name         string
	categoryType CategoryType
}

func NewCategory(id CategoryId, name string, categoryType CategoryType) (*Category, error) {
	category := &Category{
		id:           id,
		name:         name,
		categoryType: categoryType,
	}

	errors := validate.Validate(
		&validators.IntIsGreaterThan{Name: "Id", Field: int(category.id), Compared: 0, Message: "Id must be greater than 0"},
		&validators.StringLengthInRange{Name: "Name", Field: category.name, Min: 1, Max: 255, Message: "Name must be 1 and 255 characters long"},
		&validators.FuncValidator{Name: "CategoryType", Field: string(category.categoryType), Message: "No such category type %q. Must be either INCOME or EXPENSE", Fn: func() bool { return category.categoryType == Income || category.categoryType == Expense }},
	)

	if errors.HasAny() {
		flatErrors := map[string]string{}
		for field, violations := range errors.Errors {
			flatErrors[field] = strings.Join(violations, ", ")
		}
		listErrors := []string{}
		for _, violations := range flatErrors {
			listErrors = append(listErrors, violations)
		}
		return nil, NewErrorWithFields(ErrCategoryValidation, strings.Join(listErrors, ", "), errors, flatErrors)
	}
	return category, nil
}

func (c Category) Id() CategoryId {
	return c.id
}

func (c Category) Name() string {
	return c.name
}

func (c Category) Type() CategoryType {
	return c.categoryType
}

func (c Category) String() string {
	return fmt.Sprintf("Category{id: %d, name: %s, type: %s}", c.id, c.name, c.categoryType)
}
