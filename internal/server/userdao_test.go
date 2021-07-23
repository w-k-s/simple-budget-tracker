package server

import (
	"log"
	"testing"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/w-k-s/simple-budget-tracker/pkg/ledger"
	dao "github.com/w-k-s/simple-budget-tracker/pkg/persistence"
)

type UserDaoTestSuite struct {
	suite.Suite
	userDao dao.UserDao
}

func TestUserDaoTestSuite(t *testing.T) {
	suite.Run(t, new(UserDaoTestSuite))
}

// -- SETUP

func (suite *UserDaoTestSuite) SetupTest() {
	suite.userDao = UserDao
}

// -- TEARDOWN

func (suite *UserDaoTestSuite) TearDownTest() {
	if err := ClearTables(); err != nil {
		log.Fatalf("Failed to tear down UserDaoTestSuite: %s", err)
	}
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
	aUser, _ := ledger.NewUserWithEmailString(1, "jack.torrence@theoverlook.com")

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
	userId := ledger.UserId(1)

	// WHEN
	theUser, err := suite.userDao.GetUserById(userId)

	// THEN
	assert.Equal(suite.T(), ledger.User{}, theUser)

	coreError := err.(ledger.Error)
	assert.EqualValues(suite.T(), coreError.Code(), ledger.ErrUserNotFound)
}

func (suite *UserDaoTestSuite) Test_Given_twoUsers_WHEN_theUsersHaveTheSameEmail_THEN_onlyOneUserIsSaved() {
	// GIVEN
	user1, _ := ledger.NewUserWithEmailString(1, "jack.torrence@theoverlook.com")
	user2, _ := ledger.NewUserWithEmailString(2, "jack.torrence@theoverlook.com")

	// WHEN
	err1 := suite.userDao.Save(user1)
	err2 := suite.userDao.Save(user2)

	// THEN
	assert.Nil(suite.T(), err1)
	assert.NotNil(suite.T(), err2)

	coreError := err2.(ledger.Error)
	assert.Equal(suite.T(), uint64(1005), uint64(coreError.Code()))
}
