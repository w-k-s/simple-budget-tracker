package ledger

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/w-k-s/simple-budget-tracker/pkg"
)

type BudgetTestSuite struct {
	suite.Suite
	testCategoryBudgets CategoryBudgets
}

func TestBudgetTestSuite(t *testing.T) {
	suite.Run(t, new(BudgetTestSuite))
}

func (suite *BudgetTestSuite) SetupTest() {
	shopping, _ := NewCategory(CategoryId(1), "Shopping", MustMakeUpdatedByUserId(1))
	shoppingLimit := MustMoney(NewMoney("AED", 2000_00))

	bills, _ := NewCategory(CategoryId(2), "Bills", MustMakeUpdatedByUserId(1))
	billsLimit := MustMoney(NewMoney("AED", 1000_00))

	suite.testCategoryBudgets = []CategoryBudget{
		MustCategoryBudget(NewCategoryBudget(shopping.id, shoppingLimit)),
		MustCategoryBudget(NewCategoryBudget(bills.id, billsLimit)),
	}
}

// -- SUITE

func (suite *BudgetTestSuite) Test_GIVEN_invalidBudgetId_WHEN_CreatingBudget_THEN_errorIsReturned() {
	// WHEN
	budget, err := NewBudget(
		0,
		AccountIds{AccountId(1)},
		BudgetPeriodTypeMonth,
		suite.testCategoryBudgets,
		MustMakeUpdatedByUserId(1),
	)

	// THEN
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), Budget{}, budget)
	assert.Equal(suite.T(), pkg.ErrBudgetValidation, errorCode(err, 0))
	assert.Equal(suite.T(), "Id must be greater than 0", err.Error())
	assert.Equal(suite.T(), "Id must be greater than 0", errorFields(err)["id"])
}

func (suite *BudgetTestSuite) Test_GIVEN_invalidAccountId_WHEN_CreatingBudget_THEN_errorIsReturned() {

	// WHEN
	budget, err := NewBudget(
		1,
		AccountIds{},
		BudgetPeriodTypeMonth,
		suite.testCategoryBudgets,
		MustMakeUpdatedByUserId(1),
	)

	// THEN
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), Budget{}, budget)
	assert.Equal(suite.T(), pkg.ErrBudgetValidation, errorCode(err, 0))
	assert.Equal(suite.T(), "AccountIds must not be empty", err.Error())
	assert.Equal(suite.T(), "AccountIds must not be empty", errorFields(err)["account_ids"])
}

func (suite *BudgetTestSuite) Test_GIVEN_noLimits_WHEN_CreatingBudget_THEN_errorIsReturned() {

	// WHEN
	budget, err := NewBudget(
		1,
		AccountIds{1},
		BudgetPeriodTypeMonth,
		CategoryBudgets{},
		MustMakeUpdatedByUserId(1),
	)

	// THEN
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), Budget{}, budget)
	assert.Equal(suite.T(), pkg.ErrBudgetValidation, errorCode(err, 0))
	assert.Equal(suite.T(), "Category Budgets can not be empty", err.Error())
	assert.Equal(suite.T(), "Category Budgets can not be empty", errorFields(err)["category_budgets"])
}

func (suite *BudgetTestSuite) Test_GIVEN_differentCurrencies_WHEN_CreatingBudget_THEN_errorIsReturned() {
	// GIVEN
	shoppingLimit := MustMoney(NewMoney("AED", 2000_00))
	billsLimit := MustMoney(NewMoney("USD", 1000_00))

	// WHEN
	budget, err := NewBudget(
		1,
		AccountIds{1},
		BudgetPeriodTypeMonth,
		CategoryBudgets{
			MustCategoryBudget(NewCategoryBudget(CategoryId(1), shoppingLimit)),
			MustCategoryBudget(NewCategoryBudget(CategoryId(2), billsLimit)),
		},
		MustMakeUpdatedByUserId(1),
	)

	// THEN
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), Budget{}, budget)
	assert.Equal(suite.T(), pkg.ErrBudgetValidation, errorCode(err, 0))
	assert.Equal(suite.T(), "Category Budgets must have the same currency", err.Error())
	assert.Equal(suite.T(), "Category Budgets must have the same currency", errorFields(err)["category_budgets"])
}

func (suite *BudgetTestSuite) Test_GIVEN_validParameters_WHEN_AccountIsCreated_THEN_noErrorsAreReturned() {

	// WHEN
	budget, err := NewBudget(
		1,
		AccountIds{1},
		BudgetPeriodTypeMonth,
		suite.testCategoryBudgets,
		MustMakeUpdatedByUserId(1),
	)

	// THEN
	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), budget)
	assert.Equal(suite.T(), BudgetId(1), budget.Id())
	assert.Equal(suite.T(), AccountIds{1}, budget.AccountIds())
	assert.Equal(suite.T(), "CategoryBudgets{1: AED 0.00 / AED 2000.00, 2: AED 0.00 / AED 1000.00}", budget.CategoryBudgets().String())
	assert.Equal(suite.T(), "UserId: 1", budget.CreatedBy().String())
	assert.Equal(suite.T(), Version(1), budget.Version())
	assert.True(suite.T(), time.Now().UTC().Sub(budget.CreatedAtUTC()) < time.Duration(1)*time.Second)
}

func (suite *BudgetTestSuite) Test_GIVEN_anAccount_WHEN_stringIsCalled_THEN_stringIsReadable() {

	// WHEN
	budget, _ := NewBudget(
		1,
		AccountIds{1},
		BudgetPeriodTypeMonth,
		suite.testCategoryBudgets,
		MustMakeUpdatedByUserId(1),
	)

	// THEN
	assert.Equal(suite.T(), "Budget{id: 1, accountIds: [1], periodType: Month, categoryBudgets: CategoryBudgets{1: AED 0.00 / AED 2000.00, 2: AED 0.00 / AED 1000.00}}", budget.String())
}
