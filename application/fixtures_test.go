package application

import (
	"time"

	"github.com/w-k-s/simple-budget-tracker/core"
)

const (
	// User
	testUserId    = core.UserId(1)
	testUserEmail = "jack.torrence@theoverlook.com"

	// Account
	testCurrentAccountId       = core.AccountId(1)
	testCurrentAccountName     = "Current"
	testCurrentAccountCurrency = "AED"

	testSavingsAccountId       = core.AccountId(2)
	testSavingsAccountName     = "Savings"
	testSavingsAccountCurrency = "AED"

	testSalaryCategoryId   = core.CategoryId(1)
	testSalaryCategoryName = "Salary"

	testBillsCategoryId   = core.CategoryId(2)
	testBillsCategoryName = "Bills"
)

var (
	testRecordDate = time.Date(2021, time.July, 5, 18, 30, 0, 0, time.UTC)
)
