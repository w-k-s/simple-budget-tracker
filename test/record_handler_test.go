package test

// - test update category last used
// - test get account by id (should show correct total balance)

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
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
	simulatedSavingAccount  ledger.Account

	otherUser           ledger.User
	otherCurrentAccount ledger.Account
	otherSalaryCategory ledger.Category
}

func TestRecordsHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(RecordsHandlerTestSuite))
}

// -- SETUP

func (suite *RecordsHandlerTestSuite) SetupTest() {

	aUser, _ := ledger.NewUserWithEmailString(1, "jack.torrence@theoverlook.com")
	otherUser, _ := ledger.NewUserWithEmailString(2, "toby.torrence@theoverlook.com")

	currentAccount, _ := ledger.NewAccount(
		ledger.AccountId(1630067787222),
		"Current",
		ledger.AccountTypeCurrent,
		"AED",
		ledger.MustMakeUpdatedByUserId(aUser.Id()),
	)
	savingAccount, _ := ledger.NewAccount(
		ledger.AccountId(1630067787223),
		"Saving",
		ledger.AccountTypeSaving,
		"AED",
		ledger.MustMakeUpdatedByUserId(aUser.Id()),
	)
	salaryCategory, _ := ledger.NewCategory(
		ledger.CategoryId(1630067305041),
		"Salary",
		ledger.MustMakeUpdatedByUserId(aUser.Id()),
	)

	otherCurrentAccount, _ := ledger.NewAccount(
		ledger.AccountId(1630067787224),
		"Current",
		ledger.AccountTypeCurrent,
		"AED",
		ledger.MustMakeUpdatedByUserId(otherUser.Id()),
	)
	otherSalaryCategory, _ := ledger.NewCategory(
		ledger.CategoryId(1630067305042),
		"Salary",
		ledger.MustMakeUpdatedByUserId(otherUser.Id()),
	)

	if err := UserDao.Save(aUser); err != nil {
		log.Fatalf("CategoriesHandlerTestSuite: Test setup failed: %s", err)
	}
	if err := UserDao.Save(otherUser); err != nil {
		log.Fatalf("CategoriesHandlerTestSuite: Test setup failed: %s", err)
	}

	tx, _ := AccountDao.BeginTx()
	_ = AccountDao.SaveTx(context.Background(), aUser.Id(), ledger.Accounts{currentAccount, savingAccount}, tx)
	_ = CategoryDao.SaveTx(context.Background(), aUser.Id(), ledger.Categories{salaryCategory}, tx)

	_ = AccountDao.SaveTx(context.Background(), otherUser.Id(), ledger.Accounts{otherCurrentAccount}, tx)
	_ = CategoryDao.SaveTx(context.Background(), otherUser.Id(), ledger.Categories{otherSalaryCategory}, tx)

	_ = tx.Commit()

	suite.simulatedUser = aUser
	suite.simulatedCurrentAccount = currentAccount
	suite.simulatedSavingAccount = savingAccount
	suite.simulatedSalaryCategory = salaryCategory
	suite.otherCurrentAccount = otherCurrentAccount
	suite.otherSalaryCategory = otherSalaryCategory
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
			},
			"totalSavings": {
				"currency": "AED",
				"value": 0
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

func (suite *RecordsHandlerTestSuite) Test_GIVEN_aRecordRequestWithLargeIncome_WHEN_createRecordsEndpointIsCalled_THEN_400IsReturned() {
	// GIVEN
	userId := suite.simulatedUser.Id()
	accountId := suite.simulatedCurrentAccount.Id()

	const body = `{
		"note": "September Salary",
		"category": {
			"id": 1630067305041
		},
		"amount": {
			"currency": "AED",
			"value": 18446744073709551616
		},
		"date": "2021-09-09T00:00:00+00:00",
		"type": "INCOME"
	}`

	var buffer bytes.Buffer
	buffer.WriteString(body)
	r, _ := http.NewRequest("POST", fmt.Sprintf("/api/v1/accounts/%d/records", accountId), &buffer)
	AddAuthorizationHeader(r, userId)

	// WHEN
	w := httptest.NewRecorder()
	TestApp.Router().ServeHTTP(w, r)

	// THEN
	expected := `{
		"detail": "Failed to parse request",
		"instance": "/api/v1/accounts/1630067787222/records",
		"status": 400,
		"title": "REQUEST_UNMARSHALLING_FAILED",
		"type": "2021-01-01T22:08:41+0000",
		"type": "/api/v1/problems/1021"
	}`
	assert.Equal(suite.T(), 400, w.Code)
	assert.JSONEq(suite.T(), expected, w.Body.String())
}

func (suite *RecordsHandlerTestSuite) Test_GIVEN_aRecordRequestWithMaxIncome_WHEN_createRecordsEndpointIsCalled_THEN_maxIncomeCanBeRetrieved() {
	// GIVEN
	userId := suite.simulatedUser.Id()
	accountId := suite.simulatedCurrentAccount.Id()

	const body = `{
		"note": "September Salary",
		"category": {
			"id": 1630067305041
		},
		"amount": {
			"currency": "AED",
			"value": 9223372036854775807
		},
		"date": "2021-09-09T00:00:00+00:00",
		"type": "INCOME"
	}`

	var buffer bytes.Buffer
	buffer.WriteString(body)
	r, _ := http.NewRequest("POST", fmt.Sprintf("/api/v1/accounts/%d/records", accountId), &buffer)
	AddAuthorizationHeader(r, userId)

	// WHEN
	w := httptest.NewRecorder()
	TestApp.Router().ServeHTTP(w, r)

	// THEN
	assert.Equal(suite.T(), 201, w.Code)

	// ---

	r, _ = http.NewRequest("GET", fmt.Sprintf("/api/v1/accounts/%d/records?latest", accountId), nil)
	AddAuthorizationHeader(r, userId)

	w = httptest.NewRecorder()
	TestApp.Router().ServeHTTP(w, r)

	expected := `{
		"records": [{
			"id": 1,
			"note": "September Salary",
			"category": {
				"id": 1630067305041,
				"name": "Salary"
			},
			"amount": {
				"currency": "AED",
				"value": 9223372036854775807
			},
			"date": "2021-09-09T00:00:00+0000",
			"type": "INCOME"
		}],
		"summary": {
			"totalExpenses": {
				"currency": "AED",
				"value": 0
			},
			"totalIncome": {
				"currency": "AED",
				"value": 9223372036854775807
			},
			"totalSavings": {
				"currency": "AED",
				"value": 0
			}
		},
		"search": {
			"from": "2021-09-09T00:00:00Z",
			"to": "2021-09-09T00:00:00Z"
		}
	}`

	assert.Equal(suite.T(), 200, w.Code)
	assert.JSONEq(suite.T(), expected, w.Body.String())
}

func (suite *RecordsHandlerTestSuite) Test_GIVEN_aRecordRequestWithLargeExpense_WHEN_createRecordsEndpointIsCalled_THEN_400IsReturned() {
	// GIVEN
	userId := suite.simulatedUser.Id()
	accountId := suite.simulatedCurrentAccount.Id()

	const body = `{
		"note": "September Salary",
		"category": {
			"id": 1630067305041
		},
		"amount": {
			"currency": "AED",
			"value": 18446744073709551616
		},
		"date": "2021-09-09T00:00:00+00:00",
		"type": "EXPENSE"
	}`

	var buffer bytes.Buffer
	buffer.WriteString(body)
	r, _ := http.NewRequest("POST", fmt.Sprintf("/api/v1/accounts/%d/records", accountId), &buffer)
	AddAuthorizationHeader(r, userId)

	// WHEN
	w := httptest.NewRecorder()
	TestApp.Router().ServeHTTP(w, r)

	// THEN
	expected := `{
		"detail": "Failed to parse request",
		"instance": "/api/v1/accounts/1630067787222/records",
		"status": 400,
		"title": "REQUEST_UNMARSHALLING_FAILED",
		"type": "2021-01-01T22:08:41+0000",
		"type": "/api/v1/problems/1021"
	}`
	assert.Equal(suite.T(), 400, w.Code)
	assert.JSONEq(suite.T(), expected, w.Body.String())
}

func (suite *RecordsHandlerTestSuite) Test_GIVEN_aRecordRequestWithMaxExpense_WHEN_createRecordsEndpointIsCalled_THEN_maxExpenseCanBeRetrieved() {
	// GIVEN
	userId := suite.simulatedUser.Id()
	accountId := suite.simulatedCurrentAccount.Id()
	log.Printf("%v\n", math.MaxInt64)
	const body = `{
		"note": "September Salary",
		"category": {
			"id": 1630067305041
		},
		"amount": {
			"currency": "AED",
			"value": 9223372036854775807
		},
		"date": "2021-09-09T00:00:00+00:00",
		"type": "EXPENSE"
	}`

	var buffer bytes.Buffer
	buffer.WriteString(body)
	r, _ := http.NewRequest("POST", fmt.Sprintf("/api/v1/accounts/%d/records", accountId), &buffer)
	AddAuthorizationHeader(r, userId)

	// WHEN
	w := httptest.NewRecorder()
	TestApp.Router().ServeHTTP(w, r)

	// THEN
	assert.Equal(suite.T(), 201, w.Code)

	// ---

	r, _ = http.NewRequest("GET", fmt.Sprintf("/api/v1/accounts/%d/records?latest", accountId), nil)
	AddAuthorizationHeader(r, userId)

	w = httptest.NewRecorder()
	TestApp.Router().ServeHTTP(w, r)

	expected := `{
		"records": [{
			"id": 1,
			"note": "September Salary",
			"category": {
				"id": 1630067305041,
				"name": "Salary"
			},
			"amount": {
				"currency": "AED",
				"value": -9223372036854775807
			},
			"date": "2021-09-09T00:00:00+0000",
			"type": "EXPENSE"
		}],
		"summary": {
			"totalExpenses": {
				"currency": "AED",
				"value": 9223372036854775807
			},
			"totalIncome": {
				"currency": "AED",
				"value": 0
			},
			"totalSavings": {
				"currency": "AED",
				"value": 0
			}
		},
		"search": {
			"from": "2021-09-09T00:00:00Z",
			"to": "2021-09-09T00:00:00Z"
		}
	}`

	assert.Equal(suite.T(), 200, w.Code)
	assert.JSONEq(suite.T(), expected, w.Body.String())
}

func (suite *RecordsHandlerTestSuite) Test_GIVEN_aRecordWithCategoryFromDifferentUser_WHEN_WHEN_createRecordsEndpointIsCalled_THEN_400IsReturned() {
	// GIVEN
	userId := suite.simulatedUser.Id()
	accountId := suite.simulatedCurrentAccount.Id()

	var createRequest svc.CreateRecordRequest
	createRequest.Note = "June 2023"
	createRequest.Amount.Currency = "AED"
	createRequest.Amount.Value = 100_00
	createRequest.Category.Id = uint64(suite.otherSalaryCategory.Id())
	createRequest.DateUTC = "2023-01-01T22:08:41+00:00"
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
	expected := `{
		"detail": "Category with id 1630067305042 not found",
		"instance": "/api/v1/accounts/1630067787222/records",
		"status": 404,
		"title": "CATEGORIES_NOT_FOUND",
		"type": "2021-01-01T22:08:41+0000",
		"type": "/api/v1/problems/1013"
	}`
	assert.Equal(suite.T(), 404, w.Code)
	assert.JSONEq(suite.T(), expected, w.Body.String())
}

func (suite *RecordsHandlerTestSuite) Test_GIVEN_aTransferRecordWithPositiveAmount_WHEN_createRecordsEndpointIsCalled_THEN_senderAmountisNegativeAndReceiverAmountIsPositive() {
	// GIVEN
	userId := suite.simulatedUser.Id()
	currentAccountId := suite.simulatedCurrentAccount.Id()

	var createRequest svc.CreateRecordRequest
	createRequest.Note = "September Salary"
	createRequest.Amount.Currency = "AED"
	createRequest.Amount.Value = 100_00
	createRequest.Category.Id = uint64(suite.simulatedSalaryCategory.Id())
	createRequest.DateUTC = "2023-01-01T22:08:41+00:00"
	createRequest.Type = string(ledger.Transfer)
	createRequest.Transfer.Beneficiary.Id = uint64(suite.simulatedSavingAccount.Id())

	data, _ := json.Marshal(createRequest)

	var buffer bytes.Buffer
	buffer.Write(data)

	r, _ := http.NewRequest("POST", fmt.Sprintf("/api/v1/accounts/%d/records", currentAccountId), &buffer)
	AddAuthorizationHeader(r, userId)

	// WHEN
	w := httptest.NewRecorder()
	TestApp.Router().ServeHTTP(w, r)

	// THEN
	assert.Equal(suite.T(), 201, w.Code)

	// --- Sender Account
	r, _ = http.NewRequest("GET", fmt.Sprintf("/api/v1/accounts/%d/records?latest", currentAccountId), nil)
	AddAuthorizationHeader(r, userId)

	w = httptest.NewRecorder()
	TestApp.Router().ServeHTTP(w, r)

	expected := `{
		"records": [{
			"id": 1,
			"note": "September Salary",
			"category": {
				"id": 1630067305041,
				"name": "Salary"
			},
			"amount": {
				"currency": "AED",
				"value": -10000
			},
			"date": "2023-01-01T00:00:00+0000",
			"type": "TRANSFER",
            "transfer": {
                "beneficiary": {
                    "id": 1630067787223
                }
            }
		}],
		"summary": {
			"totalExpenses": {
				"currency": "AED",
				"value": 0
			},
			"totalIncome": {
				"currency": "AED",
				"value": 0
			},
			"totalSavings": {
				"currency": "AED",
				"value": 10000
			}
		},
		"search": {
			"from": "2023-01-01T00:00:00Z",
			"to": "2023-01-01T00:00:00Z"
		}
	}`

	assert.Equal(suite.T(), 200, w.Code)
	assert.JSONEq(suite.T(), expected, w.Body.String())

	// --- Receiver Account
	r, _ = http.NewRequest("GET", fmt.Sprintf("/api/v1/accounts/%d/records?latest", suite.simulatedSavingAccount.Id()), nil)
	AddAuthorizationHeader(r, userId)

	w = httptest.NewRecorder()
	TestApp.Router().ServeHTTP(w, r)

	expected = `{
		"records": [{
			"id": 2,
			"note": "September Salary",
			"category": {
				"id": 1630067305041,
				"name": "Salary"
			},
			"amount": {
				"currency": "AED",
				"value": 10000
			},
			"date": "2023-01-01T00:00:00+0000",
			"type": "TRANSFER",
            "transfer": {
                "beneficiary": {
                    "id": 1630067787223
                }
            }
		}],
		"summary": {
			"totalExpenses": {
				"currency": "AED",
				"value": 0
			},
			"totalIncome": {
				"currency": "AED",
				"value": 0
			},
			"totalSavings": {
				"currency": "AED",
				"value": 10000
			}
		},
		"search": {
			"from": "2023-01-01T00:00:00Z",
			"to": "2023-01-01T00:00:00Z"
		}
	}`

	assert.Equal(suite.T(), 200, w.Code)
	assert.JSONEq(suite.T(), expected, w.Body.String())
}

func (suite *RecordsHandlerTestSuite) Test_GIVEN_aTransferRecordWithNegativeAmount_WHEN_createRecordsEndpointIsCalled_THEN_senderAmountisNegativeAndReceiverAmountIsPositive() {
	// GIVEN
	userId := suite.simulatedUser.Id()
	currentAccountId := suite.simulatedCurrentAccount.Id()

	var createRequest svc.CreateRecordRequest
	createRequest.Note = "September Salary"
	createRequest.Amount.Currency = "AED"
	createRequest.Amount.Value = -100_00
	createRequest.Category.Id = uint64(suite.simulatedSalaryCategory.Id())
	createRequest.DateUTC = "2023-01-01T22:08:41+00:00"
	createRequest.Type = string(ledger.Transfer)
	createRequest.Transfer.Beneficiary.Id = uint64(suite.simulatedSavingAccount.Id())

	data, _ := json.Marshal(createRequest)

	var buffer bytes.Buffer
	buffer.Write(data)

	r, _ := http.NewRequest("POST", fmt.Sprintf("/api/v1/accounts/%d/records", currentAccountId), &buffer)
	AddAuthorizationHeader(r, userId)

	// WHEN
	w := httptest.NewRecorder()
	TestApp.Router().ServeHTTP(w, r)

	// THEN
	assert.Equal(suite.T(), 201, w.Code)

	// --- Sender Account
	r, _ = http.NewRequest("GET", fmt.Sprintf("/api/v1/accounts/%d/records?latest", currentAccountId), nil)
	AddAuthorizationHeader(r, userId)

	w = httptest.NewRecorder()
	TestApp.Router().ServeHTTP(w, r)

	expected := `{
		"records": [{
			"id": 1,
			"note": "September Salary",
			"category": {
				"id": 1630067305041,
				"name": "Salary"
			},
			"amount": {
				"currency": "AED",
				"value": -10000
			},
			"date": "2023-01-01T00:00:00+0000",
			"type": "TRANSFER",
            "transfer": {
                "beneficiary": {
                    "id": 1630067787223
                }
            }
		}],
		"summary": {
			"totalExpenses": {
				"currency": "AED",
				"value": 0
			},
			"totalIncome": {
				"currency": "AED",
				"value": 0
			},
			"totalSavings": {
				"currency": "AED",
				"value": 10000
			}
		},
		"search": {
			"from": "2023-01-01T00:00:00Z",
			"to": "2023-01-01T00:00:00Z"
		}
	}`

	assert.Equal(suite.T(), 200, w.Code)
	assert.JSONEq(suite.T(), expected, w.Body.String())

	// --- Receiver Account
	r, _ = http.NewRequest("GET", fmt.Sprintf("/api/v1/accounts/%d/records?latest", suite.simulatedSavingAccount.Id()), nil)
	AddAuthorizationHeader(r, userId)

	w = httptest.NewRecorder()
	TestApp.Router().ServeHTTP(w, r)

	expected = `{
		"records": [{
			"id": 2,
			"note": "September Salary",
			"category": {
				"id": 1630067305041,
				"name": "Salary"
			},
			"amount": {
				"currency": "AED",
				"value": 10000
			},
			"date": "2023-01-01T00:00:00+0000",
			"type": "TRANSFER",
            "transfer": {
                "beneficiary": {
                    "id": 1630067787223
                }
            }
		}],
		"summary": {
			"totalExpenses": {
				"currency": "AED",
				"value": 0
			},
			"totalIncome": {
				"currency": "AED",
				"value": 0
			},
			"totalSavings": {
				"currency": "AED",
				"value": 10000
			}
		},
		"search": {
			"from": "2023-01-01T00:00:00Z",
			"to": "2023-01-01T00:00:00Z"
		}
	}`

	assert.Equal(suite.T(), 200, w.Code)
	assert.JSONEq(suite.T(), expected, w.Body.String())
}
