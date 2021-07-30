package ledger

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ErrorTestSuite struct {
	suite.Suite
}

func TestErrorTestSuite(t *testing.T) {
	suite.Run(t, new(ErrorTestSuite))
}

// -- SUITE

func (suite *ErrorTestSuite) Test_GIVEN_errorCode_WHEN_mappedToNumber_THEN_mappingIsCorrect() {
	assert.Equal(suite.T(), uint64(1000), uint64(ErrUnknown))
	assert.Equal(suite.T(), uint64(1001), uint64(ErrDatabaseConnectivity))
	assert.Equal(suite.T(), uint64(1002), uint64(ErrDatabaseState))
	assert.Equal(suite.T(), uint64(1003), uint64(ErrUserIdDuplicated))
	assert.Equal(suite.T(), uint64(1004), uint64(ErrUserEmailInvalid))
	assert.Equal(suite.T(), uint64(1005), uint64(ErrUserEmailDuplicated))
	assert.Equal(suite.T(), uint64(1006), uint64(ErrUserNotFound))
	assert.Equal(suite.T(), uint64(1007), uint64(ErrAccountValidation))
	assert.Equal(suite.T(), uint64(1008), uint64(ErrAccountNotFound))
	assert.Equal(suite.T(), uint64(1009), uint64(ErrAccountNameDuplicated))
	assert.Equal(suite.T(), uint64(1010), uint64(ErrCurrencyInvalidCode))
	assert.Equal(suite.T(), uint64(1011), uint64(ErrCategoryValidation))
	assert.Equal(suite.T(), uint64(1012), uint64(ErrCategoryNameDuplicated))
	assert.Equal(suite.T(), uint64(1013), uint64(ErrCategoriesNotFound))
	assert.Equal(suite.T(), uint64(1014), uint64(ErrRecordValidation))
	assert.Equal(suite.T(), uint64(1015), uint64(ErrAmountOverflow))
	assert.Equal(suite.T(), uint64(1016), uint64(ErrAmountMismatchingCurrencies))
	assert.Equal(suite.T(), uint64(1017), uint64(ErrAmountTotalOfEmptySet))
	assert.Equal(suite.T(), uint64(1018), uint64(ErrAuditValidation))
	assert.Equal(suite.T(), uint64(1019), uint64(ErrRequestUnmarshallingFailed))
}

func (suite *ErrorTestSuite) Test_GIVEN_errorCode_WHEN_mappedToHttpStatus_THEN_mappingIsCorrect() {
	assert.Equal(suite.T(), http.StatusInternalServerError, ErrUnknown.Status())
	assert.Equal(suite.T(), http.StatusInternalServerError, ErrDatabaseConnectivity.Status())
	assert.Equal(suite.T(), http.StatusInternalServerError, ErrDatabaseState.Status())
	assert.Equal(suite.T(), http.StatusBadRequest, ErrUserIdDuplicated.Status())
	assert.Equal(suite.T(), http.StatusBadRequest, ErrUserEmailInvalid.Status())
	assert.Equal(suite.T(), http.StatusBadRequest, ErrUserEmailDuplicated.Status())
	assert.Equal(suite.T(), http.StatusNotFound, ErrUserNotFound.Status())
	assert.Equal(suite.T(), http.StatusBadRequest, ErrAccountValidation.Status())
	assert.Equal(suite.T(), http.StatusNotFound, ErrAccountNotFound.Status())
	assert.Equal(suite.T(), http.StatusBadRequest, ErrAccountNameDuplicated.Status())
	assert.Equal(suite.T(), http.StatusBadRequest, ErrCurrencyInvalidCode.Status())
	assert.Equal(suite.T(), http.StatusBadRequest, ErrCategoryValidation.Status())
	assert.Equal(suite.T(), http.StatusBadRequest, ErrCategoryNameDuplicated.Status())
	assert.Equal(suite.T(), http.StatusNotFound, ErrCategoriesNotFound.Status())
	assert.Equal(suite.T(), http.StatusBadRequest, ErrRecordValidation.Status())
	assert.Equal(suite.T(), http.StatusInternalServerError, ErrAmountOverflow.Status())
	assert.Equal(suite.T(), http.StatusBadRequest, ErrAmountMismatchingCurrencies.Status())
	assert.Equal(suite.T(), http.StatusInternalServerError, ErrAmountTotalOfEmptySet.Status())
	assert.Equal(suite.T(), http.StatusBadRequest, ErrAuditValidation.Status())
	assert.Equal(suite.T(), http.StatusBadRequest, ErrRequestUnmarshallingFailed.Status())
}
