package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	svc "github.com/w-k-s/simple-budget-tracker/pkg/services"
	"schneider.vip/problem"
)

type UserHandlerTestSuite struct {
	suite.Suite
}

func TestUserHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(UserHandlerTestSuite))
}

// -- SUITE

func (suite *UserHandlerTestSuite) Test_GIVEN_validCreateUserRequest_WHEN_createUserEndpointIsCalled_THEN_userIsCreatedAnd201IsReturned() {
	// GIVEN
	var request bytes.Buffer
	request.WriteString("{\"email\":\"test@burger.com\"}")
	r, _ := http.NewRequest("POST", "/api/v1/user", &request)

	// WHEN
	w := httptest.NewRecorder()
	TestApp.Router().ServeHTTP(w, r)

	// THEN
	var response svc.CreateUserResponse

	assert.Nil(suite.T(), json.Unmarshal(w.Body.Bytes(), &response))
	assert.Equal(suite.T(), 201, w.Code)
	assert.Positive(suite.T(), response.Id)
	assert.Equal(suite.T(), "test@burger.com", response.Email)

	user, err := UserDao.GetUserById(response.Id)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), "test@burger.com", user.Email().Address)

}

func (suite *HealthHandlerTestSuite) Test_GIVEN_blankRequest_WHEN_createUserEndpointIsCalled_THEN_emailValidationFailsAnd400IsReturned() {
	// GIVEN
	var request bytes.Buffer
	request.WriteString("{}")
	r, _ := http.NewRequest("POST", "/api/v1/user", &request)

	// WHEN
	w := httptest.NewRecorder()
	TestApp.Router().ServeHTTP(w, r)

	// THEN
	p := problem.New()
	assert.Equal(suite.T(), 400, w.Code)
	assert.Nil(suite.T(), p.UnmarshalJSON(w.Body.Bytes()))
	assert.Equal(suite.T(), "{\"detail\":\"mail: no address\",\"instance\":\"/api/v1/user\",\"status\":400,\"title\":\"USER_EMAIL_INVALID\",\"type\":\"/api/v1/problems/1004\"}", p.Error())
}

func (suite *HealthHandlerTestSuite) Test_GIVEN_invalidEmail_WHEN_createUserEndpointIsCalled_THEN_emailValidationFailsAnd400IsReturned() {
	// GIVEN
	var request bytes.Buffer
	request.WriteString("{\"email\":\"bob\"}")
	r, _ := http.NewRequest("POST", "/api/v1/user", &request)

	// WHEN
	w := httptest.NewRecorder()
	TestApp.Router().ServeHTTP(w, r)

	// THEN
	p := problem.New()
	assert.Equal(suite.T(), 400, w.Code)
	assert.Nil(suite.T(), p.UnmarshalJSON(w.Body.Bytes()))
	assert.Equal(suite.T(), "{\"detail\":\"mail: missing '@' or angle-addr\",\"instance\":\"/api/v1/user\",\"status\":400,\"title\":\"USER_EMAIL_INVALID\",\"type\":\"/api/v1/problems/1004\"}", p.Error())
}
