package application

import (
	"context"
	"os"
	"testing"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	tc "github.com/testcontainers/testcontainers-go"
	"github.com/w-k-s/simple-budget-tracker/core"
	"github.com/w-k-s/simple-budget-tracker/migrations"
)

type UserDaoTestSuite struct {
	suite.Suite
	containerCtx context.Context
	postgresC    tc.Container
	userDao      core.UserDao
}

func TestUserDaoTestSuite(t *testing.T) {
	suite.Run(t, new(UserDaoTestSuite))
}

// -- SETUP

func (suite *UserDaoTestSuite) SetupTest() {
	containerCtx, postgresC, dataSourceName, err := requestPostgresTestContainer()
	if err != nil {
		panic(err)
	}

	suite.containerCtx = *containerCtx
	suite.postgresC = postgresC
	migrations.MustRunMigrations(TestContainerDriverName, dataSourceName, os.Getenv("TEST_MIGRATIONS_DIRECTORY"))
	suite.userDao = MustOpenUserDao(TestContainerDriverName, dataSourceName)
}

// -- TEARDOWN

func (suite *UserDaoTestSuite) TearDownTest() {
	if container := suite.postgresC; container != nil {
		_ = container.Terminate(suite.containerCtx)
	}
	suite.userDao.Close()
}

// -- SUITE

func (suite *UserDaoTestSuite) Test_WHEN_NewUserIdIsCalled_THEN_userIdIsReturnedFromDatabaseSequence() {
	// WHEN
	userId, err := suite.userDao.NewUserId()

	// THEN
	assert.Nil(suite.T(), err)
	assert.Positive(suite.T(), userId)
}

func (suite *UserDaoTestSuite) Test_Given_aUser_WHEN_theUserIsSaved_THEN_userCanBeRetrieved() {
	// GIVEN
	aUser, _ := core.NewUserWithEmailString(1, "jack.torrence@theoverlook.com")

	// WHEN
	_ = suite.userDao.Save(aUser)
	theUser, err := suite.userDao.GetUserById(1)

	// THEN
	assert.Nil(suite.T(), err)
	assert.EqualValues(suite.T(), aUser.Email().Address, theUser.Email().Address)
	assert.EqualValues(suite.T(), aUser.Id(), theUser.Id())
}

func (suite *UserDaoTestSuite) Test_Given_aUserId_WHEN_noUserWithThatIdExists_THEN_appropriateErrorIsReturned() {
	// GIVEN
	userId := core.UserId(1)

	// WHEN
	theUser, err := suite.userDao.GetUserById(userId)

	// THEN
	assert.Nil(suite.T(), theUser)

	coreError := err.(core.Error)
	assert.EqualValues(suite.T(), coreError.Code(), core.ErrUserNotFound)
}

func (suite *UserDaoTestSuite) Test_Given_twoUsers_WHEN_theUsersHaveTheSameEmail_THEN_onlyOneUserIsSaved() {
	// GIVEN
	user1, _ := core.NewUserWithEmailString(1, "jack.torrence@theoverlook.com")
	user2, _ := core.NewUserWithEmailString(2, "jack.torrence@theoverlook.com")

	// WHEN
	err1 := suite.userDao.Save(user1)
	err2 := suite.userDao.Save(user2)

	// THEN
	assert.Nil(suite.T(), err1)
	assert.NotNil(suite.T(), err2)

	coreError := err2.(core.Error)
	assert.Equal(suite.T(), uint64(1005), uint64(coreError.Code()))
}
