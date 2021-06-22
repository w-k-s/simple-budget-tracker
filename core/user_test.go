package core

import (
	"testing"

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
	account, err := NewUserWithEmailString(1, "bob")

	// THEN
	assert.Nil(suite.T(), account)
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), uint64(1004), uint64(err.(Error).Code()))
	assert.Equal(suite.T(), "mail: missing '@' or angle-addr", err.(Error).Error())
}

func (suite *UserTestSuite) Test_GIVEN_blankUserEmail_WHEN_userIsCreated_THEN_errorIsReturned() {
	
	// WHEN
	account, err := NewUserWithEmailString(1, "")

	// THEN
	assert.Nil(suite.T(), account)
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), uint64(1004), uint64(err.(Error).Code()))
	assert.Equal(suite.T(), "mail: no address", err.(Error).Error())
}