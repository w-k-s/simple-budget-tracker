package ledger

import (
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/w-k-s/simple-budget-tracker/pkg"
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
	category, err := NewCategory(categoryId, "Shopping", MustMakeUpdatedByUserId(UserId(1)))

	// THEN
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), Category{}, category)
	assert.Equal(suite.T(), pkg.ErrCategoryValidation, errorCode(err, 0))
	assert.Equal(suite.T(), "Id must be greater than 0", err.Error())
	assert.Equal(suite.T(), "Id must be greater than 0", errorFields(err)["id"])
}

func (suite *CategoryTestSuite) Test_GIVEN_emptyCategoryName_WHEN_CategoryIsCreated_THEN_errorIsReturned() {

	// WHEN
	category, err := NewCategory(2, "", MustMakeUpdatedByUserId(UserId(1)))

	// THEN
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), Category{}, category)
	assert.Equal(suite.T(), pkg.ErrCategoryValidation, errorCode(err, 0))
	assert.Equal(suite.T(), "Name must be 1 and 25 characters long", err.Error())
	assert.Equal(suite.T(), "Name must be 1 and 25 characters long", errorFields(err)["name"])
}

func (suite *CategoryTestSuite) Test_GIVEN_validParameters_WHEN_CategoryIsCreated_THEN_noErrorsAreReturned() {

	// WHEN
	category, err := NewCategory(2, "Shopping", MustMakeUpdatedByUserId(UserId(1)))

	// THEN
	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), category)
	assert.Equal(suite.T(), CategoryId(2), category.Id())
	assert.Equal(suite.T(), "Shopping", category.Name())
	assert.Equal(suite.T(), "UserId: 1", category.CreatedBy().String())
	assert.Equal(suite.T(), Version(1), category.Version())
	assert.True(suite.T(), time.Now().UTC().Sub(category.CreatedAtUTC()) < time.Duration(1)*time.Second)
}

func (suite *CategoryTestSuite) Test_GIVEN_aCategory_WHEN_stringIsCalled_THEN_stringIsReadable() {

	// WHEN
	category, _ := NewCategory(2, "Shopping", MustMakeUpdatedByUserId(UserId(1)))

	// THEN
	assert.Equal(suite.T(), "Category{id: 2, name: Shopping}", category.String())
}

func (suite *CategoryTestSuite) Test_GIVEN_categories_WHEN_namesIsCalled_THEN_sliceOfSortedCategoryNamesIsReturned() {

	// WHEN
	category1, _ := NewCategory(1, "Health", MustMakeUpdatedByUserId(UserId(1)))
	category2, _ := NewCategory(1, "Entertainment", MustMakeUpdatedByUserId(UserId(1)))

	categories := Categories([]Category{
		category1,
		category2,
	})

	// THEN
	assert.Equal(suite.T(), []string{"Entertainment", "Health"}, categories.Names())
}

func (suite *CategoryTestSuite) Test_GIVEN_categories_WHEN_sortIsCalled_THEN_categoriesAreSortedInPlace() {

	// WHEN
	category1, _ := NewCategory(1, "Health", MustMakeUpdatedByUserId(UserId(1)))
	category2, _ := NewCategory(1, "Entertainment", MustMakeUpdatedByUserId(UserId(1)))

	categories := Categories([]Category{
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
	category1, _ := NewCategory(1, "Health", MustMakeUpdatedByUserId(UserId(1)))
	category2, _ := NewCategory(2, "Entertainment", MustMakeUpdatedByUserId(UserId(1)))

	categories := Categories([]Category{
		category1,
		category2,
	})

	// THEN
	assert.Equal(suite.T(), "Categories{Category{id: 2, name: Entertainment}, Category{id: 1, name: Health}}", categories.String())
}

func (suite *CategoryTestSuite) Test_GIVEN_categoryNames_WHEN_categoriesAreSaved_THEN_categoryNamesAreCapitalized() {

	// WHEN
	category1, _ := NewCategory(1, "health", MustMakeUpdatedByUserId(UserId(1)))
	category2, _ := NewCategory(2, "ENTERTAINMENT", MustMakeUpdatedByUserId(UserId(1)))

	categories := Categories([]Category{
		category1,
		category2,
	})

	// THEN
	assert.Equal(suite.T(), "Categories{Category{id: 2, name: Entertainment}, Category{id: 1, name: Health}}", categories.String())
}
