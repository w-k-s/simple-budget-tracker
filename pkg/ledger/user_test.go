package ledger

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type UserTestSuite struct {
	suite.Suite
}

func TestUserTestSuite(t *testing.T) {
	suite.Run(t, new(UserTestSuite))
}

// -- SUITE

func (suite *UserTestSuite) Test_GIVEN_invalidUserEmail_WHEN_userIsCreated_THEN_errorIsReturned() {

	// WHEN
	userId := UserId(1)
	user, err := NewUserWithEmailString(userId, "bob")

	// THEN
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), User{}, user)
	assert.Equal(suite.T(), ErrUserEmailInvalid, err.(Error).Code())
	assert.Equal(suite.T(), "mail: missing '@' or angle-addr", err.(Error).Error())
}

func (suite *UserTestSuite) Test_GIVEN_blankUserEmail_WHEN_userIsCreated_THEN_errorIsReturned() {

	// WHEN
	userId := UserId(1)
	user, err := NewUserWithEmailString(userId, "")

	// THEN
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), User{}, user)
	assert.Equal(suite.T(), uint64(1004), uint64(err.(Error).Code()))
	assert.Equal(suite.T(), "mail: no address", err.(Error).Error())
}

func (suite *UserTestSuite) Test_GIVEN_validUserEmail_WHEN_userIsCreated_THEN_userIsCreatedSuccessfully() {

	// WHEN
	userId := UserId(1)
	user, err := NewUserWithEmailString(userId, "john@example.com")

	// THEN
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), userId, user.Id())
	assert.Equal(suite.T(), "john@example.com", user.Email().Address)
	assert.Equal(suite.T(), "UserId: 1", user.CreatedBy().String())
	assert.True(suite.T(), time.Now().In(time.UTC).Sub(user.createdAtUTC) < time.Duration(1)*time.Second)
}
