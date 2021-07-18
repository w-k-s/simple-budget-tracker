package application

import (
	"encoding/json"
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

func (suite *HealthHandlerTestSuite) Test_GIVEN_statusReport_WHEN_statusIsUp_THEN_jsonEncodedCorrectlyAndHttpStatusIsOk() {
	// GIVEN
	report := make(StatusReport)

	// WHEN
	report["database"] = up

	// THEN
	bytes, err := json.Marshal(report)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), up, report.overallStatus())
	assert.Equal(suite.T(), 200, report.overallStatus().HttpCode())

	statusJson := string(bytes)
	assert.Equal(suite.T(), "{\"database\":\"UP\"}", statusJson)
}

func (suite *HealthHandlerTestSuite) Test_GIVEN_statusReport_WHEN_statusIsDown_THEN_jsonEncodedCorrectlyAndHttpStatusIs500() {
	// GIVEN
	report := make(StatusReport)

	// WHEN
	report["database"] = down

	// THEN
	bytes, err := json.Marshal(report)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), down, report.overallStatus())
	assert.Equal(suite.T(), 500, report.overallStatus().HttpCode())

	statusJson := string(bytes)
	assert.Equal(suite.T(), "{\"database\":\"DOWN\"}", statusJson)
}

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
