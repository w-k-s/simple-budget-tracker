package test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type HealthHandlerTestSuite struct {
	suite.Suite
}

func TestHealthHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(HealthHandlerTestSuite))
}

// -- SUITE

func (suite *HealthHandlerTestSuite) Test_GIVEN_databaseIsUp_WHEN_healthEndpointIsCalled_THEN_statusIsUpAndResponseCodeIs200() {
	// GIVEN
	r, _ := http.NewRequest("GET", "/health", nil)

	// WHEN
	w := httptest.NewRecorder()
	TestApp.Router().ServeHTTP(w, r)

	// THEN
	assert.Equal(suite.T(), 200, w.Code)
	assert.Equal(suite.T(), "{\"database\":\"UP\"}\n", w.Body.String())
}
