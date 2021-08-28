package test

// - test update category last used
// - given a transfer with category of different user, when record is saved, then error is returned
// - given a transfer sent with positive amount, when record is saved, sender account amount is negative
// - given a transfer sent with a negative amount, when record is saved, sender account amount is negative
// - given a transfer sent with positive amount, when record is saved, receiver account amount is positive
// - given a transfer sent with negative amount, when record is saved, receiver account amount is positive
// - test get account by id (should show correct total balance)

// - given an account receives transfer, when receiver account edits transfer, then error is returned
// - given an account receives transfer, when receiver account deletes transfer, then error is returned

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/w-k-s/simple-budget-tracker/pkg/ledger"
	svc "github.com/w-k-s/simple-budget-tracker/pkg/services"
)

type RecordsHandlerTestSuite struct {
	suite.Suite
	simulatedUser           ledger.User
	simulatedCurrentAccount ledger.Account
	simulatedSalaryCategory ledger.Category
}

func TestRecordsHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(RecordsHandlerTestSuite))
}

// -- SETUP

func (suite *RecordsHandlerTestSuite) SetupTest() {
	var (
		aUser ledger.User
		err   error
	)
	aUser, _ = ledger.NewUserWithEmailString(1, "jack.torrence@theoverlook.com")
	if err = UserDao.Save(aUser); err != nil {
		log.Fatalf("CategoriesHandlerTestSuite: Test setup failed: %s", err)
	}

	currentAccount, _ := ledger.NewAccount(ledger.AccountId(1630067787222), "Current", "AED", ledger.MustMakeUpdatedByUserId(aUser.Id()))
	salaryCategory, _ := ledger.NewCategory(ledger.CategoryId(1630067305041), "Salary", ledger.MustMakeUpdatedByUserId(aUser.Id()))

	tx, _ := AccountDao.BeginTx()
	_ = AccountDao.SaveTx(context.Background(), aUser.Id(), ledger.Accounts{currentAccount}, tx)
	_ = CategoryDao.SaveTx(context.Background(), aUser.Id(), ledger.Categories{salaryCategory}, tx)
	_ = tx.Commit()

	suite.simulatedUser = aUser
	suite.simulatedCurrentAccount = currentAccount
	suite.simulatedSalaryCategory = salaryCategory
}

func (suite *RecordsHandlerTestSuite) TearDownTest() {
	if err := ClearTables(); err != nil {
		log.Fatalf("Failed to tear down RecordsHandlerTestSuite: %s", err)
	}
}

// -- SUITE

func (suite *RecordsHandlerTestSuite) Test_GIVEN_validCreateRecordRequest_WHEN_createRecordsEndpointIsCalled_THEN_recordIsCreatedAnd201IsReturned() {
	// GIVEN
	userId := suite.simulatedUser.Id()
	accountId := suite.simulatedCurrentAccount.Id()

	var createRequest svc.CreateRecordRequest
	createRequest.Account.Id = uint64(accountId)
	createRequest.Note = "Salary"
	createRequest.Amount.Currency = "AED"
	createRequest.Amount.Value = 10000
	createRequest.Category.Id = uint64(suite.simulatedSalaryCategory.Id())
	createRequest.DateUTC = "2021-01-01T22:08:41+00:00"
	createRequest.Type = string(ledger.Income)

	data, _ := json.Marshal(createRequest)

	var buffer bytes.Buffer
	buffer.Write(data)
	r, _ := http.NewRequest("POST", fmt.Sprintf("/api/v1/accounts/%d/records", accountId), &buffer)
	AddAuthorizationHeader(r, userId)

	// WHEN
	w := httptest.NewRecorder()
	TestApp.Router().ServeHTTP(w, r)

	// THEN
	var createResponse svc.RecordResponse
	expected := `{
		"id": 1,
		"note": "Salary",
		"category": {
			"id": 1630067305041,
			"name": "Salary"
		},
		"amount": {
			"currency": "AED",
			"value": 10000
		},
		"date": "2021-01-01T22:08:41+0000",
		"type": "INCOME",
		"account": {
			"id": 1630067787222,
			"currentBalance": {
				"currency": "AED",
				"value": 10000
			}
		}
	}`
	assert.Nil(suite.T(), json.Unmarshal(w.Body.Bytes(), &createResponse))
	assert.Equal(suite.T(), 201, w.Code)
	assert.JSONEq(suite.T(), expected, w.Body.String())
	// ---

	r, _ = http.NewRequest("GET", fmt.Sprintf("/api/v1/accounts/%d/records?latest", accountId), nil)
	AddAuthorizationHeader(r, userId)

	w = httptest.NewRecorder()
	TestApp.Router().ServeHTTP(w, r)

	expected = `{
		"records": [{
			"id": 1,
			"note": "Salary",
			"category": {
				"id": 1630067305041,
				"name": "Salary"
			},
			"amount": {
				"currency": "AED",
				"value": 10000
			},
			"date": "2021-01-01T00:00:00+0000",
			"type": "INCOME"
		}],
		"summary": {
			"totalExpenses": {
				"currency": "AED",
				"value": 0
			},
			"totalIncome": {
				"currency": "AED",
				"value": 10000
			}
		},
		"search": {
			"from": "2021-01-01T00:00:00Z",
			"to": "2021-01-01T00:00:00Z"
		}
	}`

	assert.Equal(suite.T(), 200, w.Code)
	assert.JSONEq(suite.T(), expected, w.Body.String())
}
