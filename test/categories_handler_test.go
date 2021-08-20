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

type CategoriesHandlerTestSuite struct {
	suite.Suite
}

func TestCategoriesHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(CategoriesHandlerTestSuite))
}

// -- SETUP

func (suite *CategoriesHandlerTestSuite) SetupTest() {
	aUser, _ := ledger.NewUserWithEmailString(testUserId, testUserEmail)
	if err := UserDao.Save(aUser); err != nil {
		log.Fatalf("AccountDaoTestSuite: Test setup failed: %s", err)
	}
}

func (suite *CategoriesHandlerTestSuite) TearDownTest() {
	if err := ClearTables(); err != nil {
		log.Fatalf("Failed to tear down CategoriesHandlerTestSuite: %s", err)
	}
}

// -- SUITE

func (suite *CategoriesHandlerTestSuite) Test_GIVEN_validCreateCategoriesRequest_WHEN_createCategoriesEndpointIsCalled_THEN_categoriesAreCreatedAnd201IsReturned() {
	// GIVEN
	var createRequest bytes.Buffer
	createRequest.WriteString("{\"categories\":[{\"name\":\"Bills\"},{\"name\":\"Groceries\"}]}")
	r, _ := http.NewRequest("POST", "/api/v1/categories", &createRequest)
	AddAuthorizationHeader(r, testUserId)

	// WHEN
	w := httptest.NewRecorder()
	TestApp.Router().ServeHTTP(w, r)

	// THEN
	var createResponse svc.CategoriesResponse

	assert.Nil(suite.T(), json.Unmarshal(w.Body.Bytes(), &createResponse))
	assert.Equal(suite.T(), 201, w.Code)
	assert.Equal(suite.T(), 2, len(createResponse.Categories))

	assert.Positive(suite.T(), createResponse.Categories[0].Id)
	assert.Equal(suite.T(), "Bills", createResponse.Categories[0].Name)

	assert.Positive(suite.T(), createResponse.Categories[1].Id)
	assert.Equal(suite.T(), "Groceries", createResponse.Categories[1].Name)

	// ---

	r, _ = http.NewRequest("GET", "/api/v1/categories", nil)
	AddAuthorizationHeader(r, testUserId)

	w = httptest.NewRecorder()
	TestApp.Router().ServeHTTP(w, r)

	var getResponse svc.CategoriesResponse

	assert.Nil(suite.T(), json.Unmarshal(w.Body.Bytes(), &getResponse))
	assert.Equal(suite.T(), 200, w.Code)
	assert.Equal(suite.T(), 2, len(getResponse.Categories))

	assert.Positive(suite.T(), getResponse.Categories[0].Id)
	assert.Equal(suite.T(), "Bills", getResponse.Categories[0].Name)

	assert.Positive(suite.T(), getResponse.Categories[1].Id)
	assert.Equal(suite.T(), "Groceries", getResponse.Categories[1].Name)
}

func (suite *CategoriesHandlerTestSuite) Test_GIVEN_emptyRequest_WHEN_createCategoriesEndpointIsCalled_THEN_noCategoriesAreCreatedAnd201IsReturned() {
	// GIVEN
	var request bytes.Buffer
	request.WriteString("{\"categories\":[]}")
	r, _ := http.NewRequest("POST", "/api/v1/categories", &request)
	AddAuthorizationHeader(r, testUserId)

	// WHEN
	w := httptest.NewRecorder()
	TestApp.Router().ServeHTTP(w, r)

	// THEN
	var response svc.CategoriesResponse

	assert.Nil(suite.T(), json.Unmarshal(w.Body.Bytes(), &response))
	assert.Equal(suite.T(), 201, w.Code)
	assert.Equal(suite.T(), 0, len(response.Categories))
}

func (suite *CategoriesHandlerTestSuite) Test_GIVEN_categoriesRequestWithDuplicateNames_WHEN_createCategoriesEndpointIsCalled_THEN_categoriesNameValidationFailsAnd400IsReturned() {
	// GIVEN
	var request bytes.Buffer
	request.WriteString("{\"categories\":[{\"name\":\"Bills\"},{\"name\":\"Bills\"}]}")
	r, _ := http.NewRequest("POST", "/api/v1/categories", &request)
	AddAuthorizationHeader(r, testUserId)

	// WHEN
	w := httptest.NewRecorder()
	TestApp.Router().ServeHTTP(w, r)

	// THEN
	p := problem.New()
	assert.Equal(suite.T(), 400, w.Code)
	assert.Nil(suite.T(), p.UnmarshalJSON(w.Body.Bytes()))
	assert.Equal(suite.T(), "{\"detail\":\"Category names must be unique. One of these is duplicated: Bills, Bills\",\"instance\":\"/api/v1/categories\",\"status\":400,\"title\":\"CATEGORY_NAME_DUPLICATED\",\"type\":\"/api/v1/problems/1012\"}", p.Error())
}

func (suite *CategoriesHandlerTestSuite) Test_GIVEN_categoriesRequestWithBlankName_WHEN_createCategoriesEndpointIsCalled_THEN_nameValidationFailsAnd400IsReturned() {
	// GIVEN
	var request bytes.Buffer
	request.WriteString("{\"categories\":[{\"name\":\"\"}]}")
	r, _ := http.NewRequest("POST", "/api/v1/categories", &request)
	AddAuthorizationHeader(r, testUserId)

	// WHEN
	w := httptest.NewRecorder()
	TestApp.Router().ServeHTTP(w, r)

	// THEN
	p := problem.New()
	assert.Equal(suite.T(), 400, w.Code)
	assert.Nil(suite.T(), p.UnmarshalJSON(w.Body.Bytes()))
	assert.Equal(suite.T(), "{\"detail\":\"Name must be 1 and 25 characters long\",\"instance\":\"/api/v1/categories\",\"name\":\"Name must be 1 and 25 characters long\",\"status\":400,\"title\":\"CATEGORY_VALIDATION_FAILED\",\"type\":\"/api/v1/problems/1011\"}", p.Error())
}

func (suite *CategoriesHandlerTestSuite) Test_GIVEN_unauthenticatedRequest_WHEN_createCategoriesEndpointIsCalled_THEN_401IsReturned() {
	// GIVEN
	var request bytes.Buffer
	request.WriteString("{\"categories\":[{\"name\":\"Bills\"}]}")
	r, _ := http.NewRequest("POST", "/api/v1/categories", &request)

	// WHEN
	w := httptest.NewRecorder()
	TestApp.Router().ServeHTTP(w, r)

	// THEN
	p := problem.New()
	assert.Equal(suite.T(), 401, w.Code)
	assert.Nil(suite.T(), p.UnmarshalJSON(w.Body.Bytes()))
	assert.Equal(suite.T(), "{\"detail\":\"User id is required\",\"instance\":\"/api/v1/categories\",\"status\":401,\"title\":\"SERVICE_REQUIRED_USER_ID\",\"type\":\"/api/v1/problems/1021\"}", p.Error())
}

func (suite *CategoriesHandlerTestSuite) Test_GIVEN_unauthenticatedRequest_WHEN_getAccountsEndpointIsCalled_THEN_401IsReturned() {
	// GIVEN
	var createRequest bytes.Buffer
	createRequest.WriteString("{\"categories\":[{\"name\":\"Bills\"}]}")
	r, _ := http.NewRequest("POST", "/api/v1/categories", &createRequest)
	AddAuthorizationHeader(r, testUserId)

	// WHEN
	w := httptest.NewRecorder()
	TestApp.Router().ServeHTTP(w, r)

	// THEN
	r, _ = http.NewRequest("GET", "/api/v1/categories", nil)

	w = httptest.NewRecorder()
	TestApp.Router().ServeHTTP(w, r)

	p := problem.New()
	assert.Equal(suite.T(), 401, w.Code)
	assert.Nil(suite.T(), p.UnmarshalJSON(w.Body.Bytes()))
	assert.Equal(suite.T(), "{\"detail\":\"User id is required\",\"instance\":\"/api/v1/categories\",\"status\":401,\"title\":\"SERVICE_REQUIRED_USER_ID\",\"type\":\"/api/v1/problems/1021\"}", p.Error())
}
