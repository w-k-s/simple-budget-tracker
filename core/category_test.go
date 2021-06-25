package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type CategoryTestSuite struct {
	suite.Suite
}

func TestCategoryTestSuite(t *testing.T) {
	suite.Run(t, new(CategoryTestSuite))
}

// -- SUITE

func (suite *CategoryTestSuite) Test_GIVEN_invalidCategoryId_WHEN_CategoryIsCreated_THEN_errorIsReturned() {
	// GIVEN
	categoryId := CategoryId(0)

	// WHEN
	category, err := NewCategory(categoryId, "Shopping", Expense)

	// THEN
	assert.Nil(suite.T(), category)
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), ErrCategoryValidation, err.(Error).Code())
	assert.Equal(suite.T(), "Id must be greater than 0", err.(Error).Error())
	assert.Equal(suite.T(), "Id must be greater than 0", err.(Error).Fields()["id"])
}

func (suite *CategoryTestSuite) Test_GIVEN_emptyCategoryName_WHEN_CategoryIsCreated_THEN_errorIsReturned() {

	// WHEN
	category, err := NewCategory(2, "", Expense)

	// THEN
	assert.Nil(suite.T(), category)
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), ErrCategoryValidation, err.(Error).Code())
	assert.Equal(suite.T(), "Name must be 1 and 255 characters long", err.(Error).Error())
	assert.Equal(suite.T(), "Name must be 1 and 255 characters long", err.(Error).Fields()["name"])
}

func (suite *CategoryTestSuite) Test_GIVEN_noCategoryType_WHEN_CategoryIsCreated_THEN_errorIsReturned() {

	// WHEN
	category, err := NewCategory(2, "Shopping", "")

	// THEN
	assert.Nil(suite.T(), category)
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), ErrCategoryValidation, err.(Error).Code())
	assert.Equal(suite.T(), "No such category type \"\". Must be either INCOME or EXPENSE", err.(Error).Error())
	assert.Equal(suite.T(), "No such category type \"\". Must be either INCOME or EXPENSE", err.(Error).Fields()["category_type"])
}

func (suite *CategoryTestSuite) Test_GIVEN_anInvalidCategoryType_WHEN_CategoryIsCreated_THEN_errorIsReturned() {

	// WHEN
	category, err := NewCategory(2, "Shopping", "Shopping")

	// THEN
	assert.Nil(suite.T(), category)
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), ErrCategoryValidation, err.(Error).Code())
	assert.Equal(suite.T(), "No such category type \"Shopping\". Must be either INCOME or EXPENSE", err.(Error).Error())
	assert.Equal(suite.T(), "No such category type \"Shopping\". Must be either INCOME or EXPENSE", err.(Error).Fields()["category_type"])
}

func (suite *CategoryTestSuite) Test_GIVEN_validParameters_WHEN_CategoryIsCreated_THEN_noErrorsAreReturned() {

	// WHEN
	category, err := NewCategory(2, "Shopping", Expense)

	// THEN
	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), category)
	assert.Equal(suite.T(), CategoryId(2), category.Id())
	assert.Equal(suite.T(), "Shopping", category.Name())
	assert.Equal(suite.T(), Expense, category.Type())
}

func (suite *CategoryTestSuite) Test_GIVEN_aCategory_WHEN_stringIsCalled_THEN_stringIsReadable() {

	// WHEN
	category, _ := NewCategory(2, "Shopping", Expense)

	// THEN
	assert.Equal(suite.T(), "Category{id: 2, name: Shopping, type: EXPENSE}", category.String())
}
