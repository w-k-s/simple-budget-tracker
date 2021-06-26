package application

import (
	"context"
	"log"
	"os"
	"sort"
	"testing"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	tc "github.com/testcontainers/testcontainers-go"
	"github.com/w-k-s/simple-budget-tracker/core"
	"github.com/w-k-s/simple-budget-tracker/migrations"
)

type CategoryDaoTestSuite struct {
	suite.Suite
	containerCtx context.Context
	postgresC    tc.Container
	userDao      core.UserDao
	categoryDao  core.CategoryDao
}

func TestCategoryDaoTestSuite(t *testing.T) {
	suite.Run(t, new(CategoryDaoTestSuite))
}

// -- SETUP

func (suite *CategoryDaoTestSuite) SetupTest() {
	containerCtx, postgresC, dataSourceName, err := requestPostgresTestContainer()
	if err != nil {
		panic(err)
	}

	suite.containerCtx = *containerCtx
	suite.postgresC = postgresC
	migrations.MustRunMigrations(TestContainerDriverName, dataSourceName, os.Getenv("TEST_MIGRATIONS_DIRECTORY"))

	suite.userDao = MustOpenUserDao(TestContainerDriverName, dataSourceName)
	suite.categoryDao = MustOpenCategoryDao(TestContainerDriverName, dataSourceName)

	aUser, _ := core.NewUserWithEmailString(testUserId, testUserEmail)
	if err = suite.userDao.Save(aUser); err != nil {
		log.Fatalf("CategoryDaoTestSuite: Test setup failed: %s", err)
	}
}

// -- TEARDOWN

func (suite *CategoryDaoTestSuite) TearDownTest() {
	if container := suite.postgresC; container != nil {
		_ = container.Terminate(suite.containerCtx)
	}
	suite.categoryDao.Close()
	suite.userDao.Close()
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
	salaryCategory, _ := core.NewCategory(1, "Salary")
	rentCategory, _ := core.NewCategory(2, "Rent")

	// WHEN
	_ = suite.categoryDao.Save(testUserId, core.Categories{salaryCategory, rentCategory})
	theCategories, err := suite.categoryDao.GetCategoriesForUser(testUserId)

	// THEN
	assert.Nil(suite.T(), err)

	sort.Sort(theCategories)
	assert.EqualValues(suite.T(), core.CategoryId(2), theCategories[0].Id())
	assert.EqualValues(suite.T(), "Rent", theCategories[0].Name())

	assert.EqualValues(suite.T(), core.CategoryId(1), theCategories[1].Id())
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
	aUser, _ := core.NewUserWithEmailString(999, "evil@bob.com")
	if err := suite.userDao.Save(aUser); err != nil {
		log.Fatalf("CategoryDaoTestSuite: Failed to test duplicate category name scenario: %s", err)
	}

	category1ForUser1, _ := core.NewCategory(1, "Rent")
	category2ForUser1, _ := core.NewCategory(2, "Savings")

	category1ForUser2, _ := core.NewCategory(3, "Shopping")
	category2ForUser2, _ := core.NewCategory(4, "Shopping")

	// WHEN
	err1 := suite.categoryDao.Save(testUserId, core.Categories{category1ForUser1, category2ForUser1})
	err2 := suite.categoryDao.Save(aUser.Id(), core.Categories{category1ForUser2, category2ForUser2})

	// THEN
	assert.Nil(suite.T(), err1)
	assert.NotNil(suite.T(), err2)

	coreError := err2.(core.Error)
	assert.Equal(suite.T(), core.ErrCategoryNameDuplicated, coreError.Code())
	assert.Equal(suite.T(), "Category name must be unique", coreError.Error())
}
