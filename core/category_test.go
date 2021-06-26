package core

import (
	"sort"
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
	category, err := NewCategory(categoryId, "Shopping")

	// THEN
	assert.Nil(suite.T(), category)
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), ErrCategoryValidation, err.(Error).Code())
	assert.Equal(suite.T(), "Id must be greater than 0", err.(Error).Error())
	assert.Equal(suite.T(), "Id must be greater than 0", err.(Error).Fields()["id"])
}

func (suite *CategoryTestSuite) Test_GIVEN_emptyCategoryName_WHEN_CategoryIsCreated_THEN_errorIsReturned() {

	// WHEN
	category, err := NewCategory(2, "")

	// THEN
	assert.Nil(suite.T(), category)
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), ErrCategoryValidation, err.(Error).Code())
	assert.Equal(suite.T(), "Name must be 1 and 255 characters long", err.(Error).Error())
	assert.Equal(suite.T(), "Name must be 1 and 255 characters long", err.(Error).Fields()["name"])
}

func (suite *CategoryTestSuite) Test_GIVEN_validParameters_WHEN_CategoryIsCreated_THEN_noErrorsAreReturned() {

	// WHEN
	category, err := NewCategory(2, "Shopping")

	// THEN
	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), category)
	assert.Equal(suite.T(), CategoryId(2), category.Id())
	assert.Equal(suite.T(), "Shopping", category.Name())
}

func (suite *CategoryTestSuite) Test_GIVEN_aCategory_WHEN_stringIsCalled_THEN_stringIsReadable() {

	// WHEN
	category, _ := NewCategory(2, "Shopping")

	// THEN
	assert.Equal(suite.T(), "Category{id: 2, name: Shopping}", category.String())
}

func (suite *CategoryTestSuite) Test_GIVEN_categories_WHEN_namesIsCalled_THEN_sliceOfSortedCategoryNamesIsReturned() {

	// WHEN
	category1, _ := NewCategory(1, "Health")
	category2, _ := NewCategory(1, "Entertainment")

	categories := Categories([]*Category{
		category1,
		category2,
	})

	// THEN
	assert.Equal(suite.T(), []string{"Entertainment", "Health"}, categories.Names())
}

func (suite *CategoryTestSuite) Test_GIVEN_categories_WHEN_sortIsCalled_THEN_categoriesAreSortedInPlace() {

	// WHEN
	category1, _ := NewCategory(1, "Health")
	category2, _ := NewCategory(1, "Entertainment")

	categories := Categories([]*Category{
		category1,
		category2,
	})
	sort.Sort(categories)

	// THEN
	assert.Equal(suite.T(), "Entertainment", categories[0].Name())
	assert.Equal(suite.T(), "Health", categories[1].Name())
}

func (suite *CategoryTestSuite) Test_GIVEN_categories_WHEN_stringIsCalled_THEN_stringOfEachCategoryIsPrintedInAlphabeticalOrder() {

	// WHEN
	category1, _ := NewCategory(1, "Health")
	category2, _ := NewCategory(2, "Entertainment")

	categories := Categories([]*Category{
		category1,
		category2,
	})

	// THEN
	assert.Equal(suite.T(), "Categories{Category{id: 2, name: Entertainment}, Category{id: 1, name: Health}}", categories.String())
}
