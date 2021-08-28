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
)

type ApplicationTestSuite struct {
	suite.Suite
	testUser ledger.User
}

func TestApplicationTestSuite(t *testing.T) {
	suite.Run(t, new(ApplicationTestSuite))
}

// -- SETUP

func (suite *ApplicationTestSuite) SetupTest() {
	aUser, _ := ledger.NewUserWithEmailString(1, "jack.torrence@theoverlook.com")
	if err := UserDao.Save(aUser); err != nil {
		log.Fatalf("AccountDaoTestSuite: Test setup failed: %s", err)
	}

	suite.testUser = aUser
}

func (suite *ApplicationTestSuite) TearDownTest() {
	if err := ClearTables(); err != nil {
		log.Fatalf("Failed to tear down CategoriesHandlerTestSuite: %s", err)
	}
}

// -- SUITE

func (suite *ApplicationTestSuite) Test_GIVEN_aCategoryIsCreatedWithEmbeddedScript_WHEN_createCategoriesEndpointIsCalled_THEN_categoriesAreReturnedWithHTMLEscaped() {
	// GIVEN
	var createRequest bytes.Buffer
	createRequest.WriteString("{\"categories\":[{\"name\":\"<script>evil();</script>\"}]}")
	r, _ := http.NewRequest("POST", "/api/v1/categories", &createRequest)
	AddAuthorizationHeader(r, suite.testUser.Id())

	// WHEN
	w := httptest.NewRecorder()
	TestApp.Router().ServeHTTP(w, r)

	// THEN
	assert.Equal(suite.T(), 201, w.Code)

	// ---

	r, _ = http.NewRequest("GET", "/api/v1/categories", nil)
	AddAuthorizationHeader(r, suite.testUser.Id())

	w = httptest.NewRecorder()
	TestApp.Router().ServeHTTP(w, r)

	var getResponse svc.CategoriesResponse

	assert.Nil(suite.T(), json.Unmarshal(w.Body.Bytes(), &getResponse))
	assert.Equal(suite.T(), 200, w.Code)

	assert.Equal(suite.T(), "\u003cScript\u003eEvil();\u003c/Script\u003e", getResponse.Categories[0].Name)
}
