package test

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/w-k-s/simple-budget-tracker/pkg/ledger"
	svc "github.com/w-k-s/simple-budget-tracker/pkg/services"
	"schneider.vip/problem"
)

type AccountHandlerTestSuite struct {
	suite.Suite
}

func TestAccountHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(AccountHandlerTestSuite))
}

// -- SETUP

func (suite *AccountHandlerTestSuite) SetupTest() {
	aUser, _ := ledger.NewUserWithEmailString(testUserId, testUserEmail)
	if err := UserDao.Save(aUser); err != nil {
		log.Fatalf("AccountDaoTestSuite: Test setup failed: %s", err)
	}
}

func (suite *AccountHandlerTestSuite) TearDownTest() {
	if err := ClearTables(); err != nil {
		log.Fatalf("Failed to tear down AccountHandlerTestSuite: %s", err)
	}
}

// -- SUITE

func (suite *AccountHandlerTestSuite) Test_GIVEN_validCreateAccountsRequest_WHEN_createAccountsEndpointIsCalled_THEN_accountsAreCreatedAnd201IsReturned() {
	// GIVEN
	var createRequest bytes.Buffer
	createRequest.WriteString("{\"accounts\":[{\"name\":\"Current\", \"currency\":\"AED\"}]}")
	r, _ := http.NewRequest("POST", "/api/v1/accounts", &createRequest)
	AddAuthorizationHeader(r, testUserId)

	// WHEN
	w := httptest.NewRecorder()
	TestApp.Router().ServeHTTP(w, r)

	// THEN
	var createResponse svc.AccountsResponse

	assert.Nil(suite.T(), json.Unmarshal(w.Body.Bytes(), &createResponse))
	assert.Equal(suite.T(), 201, w.Code)
	assert.Positive(suite.T(), createResponse.Accounts[0].Id)
	assert.Equal(suite.T(), "Current", createResponse.Accounts[0].Name)
	assert.Equal(suite.T(), "AED", createResponse.Accounts[0].Currency)

	// --- 

	r, _ = http.NewRequest("GET", "/api/v1/accounts", nil)
	AddAuthorizationHeader(r, testUserId)

	w = httptest.NewRecorder()
	TestApp.Router().ServeHTTP(w, r)

	var getResponse svc.AccountsResponse

	assert.Nil(suite.T(), json.Unmarshal(w.Body.Bytes(), &getResponse))
	assert.Equal(suite.T(), 200, w.Code)
	assert.Positive(suite.T(), createResponse.Accounts[0].Id)
	assert.Equal(suite.T(), "Current", createResponse.Accounts[0].Name)
	assert.Equal(suite.T(), "AED", createResponse.Accounts[0].Currency)
}

func (suite *AccountHandlerTestSuite) Test_GIVEN_emptyRequest_WHEN_createAccountsEndpointIsCalled_THEN_noAccountsAreCreatedAnd201IsReturned() {
	// GIVEN
	var request bytes.Buffer
	request.WriteString("{\"accounts\":[]}")
	r, _ := http.NewRequest("POST", "/api/v1/accounts", &request)
	AddAuthorizationHeader(r, testUserId)

	// WHEN
	w := httptest.NewRecorder()
	TestApp.Router().ServeHTTP(w, r)

	// THEN
	var response svc.AccountsResponse

	assert.Nil(suite.T(), json.Unmarshal(w.Body.Bytes(), &response))
	assert.Equal(suite.T(), 201, w.Code)
	assert.Equal(suite.T(), 0, len(response.Accounts))
}

func (suite *AccountHandlerTestSuite) Test_GIVEN_accountRequestWithDuplicateNames_WHEN_createAccountsEndpointIsCalled_THEN_accountNameValidationFailsAnd400IsReturned() {
	// GIVEN
	var request bytes.Buffer
	request.WriteString("{\"accounts\":[{\"name\":\"Current\", \"currency\":\"AED\"},{\"name\":\"Current\", \"currency\":\"AED\"}]}")
	r, _ := http.NewRequest("POST", "/api/v1/accounts", &request)
	AddAuthorizationHeader(r, testUserId)

	// WHEN
	w := httptest.NewRecorder()
	TestApp.Router().ServeHTTP(w, r)

	// THEN
	p := problem.New()
	assert.Equal(suite.T(), 400, w.Code)
	assert.Nil(suite.T(), p.UnmarshalJSON(w.Body.Bytes()))
	assert.Equal(suite.T(), "{\"detail\":\"Acccount names must be unique. One of these is duplicated: Current, Current\",\"instance\":\"/api/v1/accounts\",\"status\":400,\"title\":\"ACCOUNT_NAME_DUPLICATED\",\"type\":\"/api/v1/problems/1009\"}", p.Error())
}

func (suite *AccountHandlerTestSuite) Test_GIVEN_accountRequestWithInvalidCurrency_WHEN_createAccountsEndpointIsCalled_THEN_currencyValidationFailsAnd400IsReturned() {
	// GIVEN
	var request bytes.Buffer
	request.WriteString("{\"accounts\":[{\"name\":\"Current\", \"currency\":\"XXX\"}]}")
	r, _ := http.NewRequest("POST", "/api/v1/accounts", &request)
	AddAuthorizationHeader(r, testUserId)

	// WHEN
	w := httptest.NewRecorder()
	TestApp.Router().ServeHTTP(w, r)

	// THEN
	p := problem.New()
	assert.Equal(suite.T(), 400, w.Code)
	assert.Nil(suite.T(), p.UnmarshalJSON(w.Body.Bytes()))
	assert.Equal(suite.T(), "{\"currency\":\"No such currency 'XXX'\",\"detail\":\"No such currency 'XXX'\",\"instance\":\"/api/v1/accounts\",\"status\":400,\"title\":\"ACCOUNT_VALIDATION_FAILED\",\"type\":\"/api/v1/problems/1007\"}", p.Error())
}

func (suite *AccountHandlerTestSuite) Test_GIVEN_unauthenticatedRequest_WHEN_createAccountsEndpointIsCalled_THEN_401IsReturned() {
	// GIVEN
	var request bytes.Buffer
	request.WriteString("{\"accounts\":[{\"name\":\"Current\", \"currency\":\"AED\"}]}")
	r, _ := http.NewRequest("POST", "/api/v1/accounts", &request)

	// WHEN
	w := httptest.NewRecorder()
	TestApp.Router().ServeHTTP(w, r)

	// THEN
	p := problem.New()
	assert.Equal(suite.T(), 401, w.Code)
	assert.Nil(suite.T(), p.UnmarshalJSON(w.Body.Bytes()))
	assert.Equal(suite.T(), "{\"detail\":\"User id is required\",\"instance\":\"/api/v1/accounts\",\"status\":401,\"title\":\"SERVICE_REQUIRED_USER_ID\",\"type\":\"/api/v1/problems/1021\"}", p.Error())
}

func (suite *AccountHandlerTestSuite) Test_GIVEN_unauthenticatedRequest_WHEN_getAccountsEndpointIsCalled_THEN_401IsReturned() {
	// GIVEN
	var createRequest bytes.Buffer
	createRequest.WriteString("{\"accounts\":[{\"name\":\"Current\", \"currency\":\"AED\"}]}")
	r, _ := http.NewRequest("POST", "/api/v1/accounts", &createRequest)
	AddAuthorizationHeader(r, testUserId)

	// WHEN
	w := httptest.NewRecorder()
	TestApp.Router().ServeHTTP(w, r)

	// THEN
	r, _ = http.NewRequest("GET", "/api/v1/accounts", nil)

	w = httptest.NewRecorder()
	TestApp.Router().ServeHTTP(w, r)

	p := problem.New()
	assert.Equal(suite.T(), 401, w.Code)
	assert.Nil(suite.T(), p.UnmarshalJSON(w.Body.Bytes()))
	assert.Equal(suite.T(), "{\"detail\":\"User id is required\",\"instance\":\"/api/v1/accounts\",\"status\":401,\"title\":\"SERVICE_REQUIRED_USER_ID\",\"type\":\"/api/v1/problems/1021\"}", p.Error())
}
