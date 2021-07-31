package test

import (
	"log"
	"sort"
	"testing"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/w-k-s/simple-budget-tracker/pkg/ledger"
	dao "github.com/w-k-s/simple-budget-tracker/pkg/persistence"
)

type CategoryDaoTestSuite struct {
	suite.Suite
	userDao     dao.UserDao
	categoryDao dao.CategoryDao
}

func TestCategoryDaoTestSuite(t *testing.T) {
	suite.Run(t, new(CategoryDaoTestSuite))
}

// -- SETUP

func (suite *CategoryDaoTestSuite) SetupTest() {

	suite.userDao = UserDao
	suite.categoryDao = CategoryDao

	aUser, _ := ledger.NewUserWithEmailString(testUserId, testUserEmail)
	if err := suite.userDao.Save(aUser); err != nil {
		log.Fatalf("CategoryDaoTestSuite: Test setup failed: %s", err)
	}
}

// -- TEARDOWN

func (suite *CategoryDaoTestSuite) TearDownTest() {
	if err := ClearTables(); err != nil {
		log.Fatalf("Failed to tear down CategoryDaoTestSuite: %s", err)
	}
}

// -- SUITE

func (suite *CategoryDaoTestSuite) Test_WHEN_NewCategoryIdIsCalled_THEN_categoryIdIsReturnedFromDatabaseSequence() {
	// WHEN
	categoryId, err := suite.categoryDao.NewCategoryId()

	// THEN
	assert.Nil(suite.T(), err)
	assert.Positive(suite.T(), categoryId)
}

func (suite *CategoryDaoTestSuite) Test_Given_categories_WHEN_theCategoriesAreSaved_THEN_categoriesCanBeRetrievedByUserId() {
	// GIVEN
	salaryCategory, _ := ledger.NewCategory(1, "Salary", ledger.MustMakeUpdatedByUserId(testUserId))
	rentCategory, _ := ledger.NewCategory(2, "Rent", ledger.MustMakeUpdatedByUserId(testUserId))

	// WHEN
	_ = suite.categoryDao.Save(testUserId, ledger.Categories{salaryCategory, rentCategory})
	theCategories, err := suite.categoryDao.GetCategoriesForUser(testUserId)

	// THEN
	assert.Nil(suite.T(), err)

	sort.Sort(theCategories)
	assert.EqualValues(suite.T(), ledger.CategoryId(2), theCategories[0].Id())
	assert.EqualValues(suite.T(), "Rent", theCategories[0].Name())

	assert.EqualValues(suite.T(), ledger.CategoryId(1), theCategories[1].Id())
	assert.EqualValues(suite.T(), "Salary", theCategories[1].Name())
}

func (suite *CategoryDaoTestSuite) Test_Given_aUserId_WHEN_noCategoriesExistForThatUser_THEN_emptyCategoryListIsReturned() {
	// WHEN
	theCategories, err := suite.categoryDao.GetCategoriesForUser(testUserId)

	// THEN
	assert.Nil(suite.T(), err)
	assert.Empty(suite.T(), theCategories)
}

func (suite *CategoryDaoTestSuite) Test_Given_twoCategoriesPerUsers_WHEN_oneUserDuplicatesCategoryNames_THEN_duplicatedCategoryNameCausesError() {
	// GIVEN
	aUser, _ := ledger.NewUserWithEmailString(999, "evil@bob.com")
	if err := suite.userDao.Save(aUser); err != nil {
		log.Fatalf("CategoryDaoTestSuite: Failed to test duplicate category name scenario: %s", err)
	}

	category1ForUser1, _ := ledger.NewCategory(1, "Rent", ledger.MustMakeUpdatedByUserId(testUserId))
	category2ForUser1, _ := ledger.NewCategory(2, "Savings", ledger.MustMakeUpdatedByUserId(testUserId))

	category1ForUser2, _ := ledger.NewCategory(3, "Shopping", ledger.MustMakeUpdatedByUserId(testUserId))
	category2ForUser2, _ := ledger.NewCategory(4, "Shopping", ledger.MustMakeUpdatedByUserId(testUserId))

	// WHEN
	err1 := suite.categoryDao.Save(testUserId, ledger.Categories{category1ForUser1, category2ForUser1})
	err2 := suite.categoryDao.Save(aUser.Id(), ledger.Categories{category1ForUser2, category2ForUser2})

	// THEN
	assert.Nil(suite.T(), err1)
	assert.NotNil(suite.T(), err2)

	coreError := err2.(ledger.Error)
	assert.Equal(suite.T(), ledger.ErrCategoryNameDuplicated, coreError.Code())
	assert.Equal(suite.T(), "Category names must be unique. One of these is duplicated: Shopping, Shopping", coreError.Error())
}
