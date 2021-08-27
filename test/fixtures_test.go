package test

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/w-k-s/simple-budget-tracker/pkg/ledger"
)

const (
	// User
	testUserId    = ledger.UserId(1)
	testUserEmail = "jack.torrence@theoverlook.com"

	// Account
	testCurrentAccountId       = ledger.AccountId(1)
	testCurrentAccountName     = "Current"
	testCurrentAccountCurrency = "AED"

	testSavingsAccountId       = ledger.AccountId(2)
	testSavingsAccountName     = "Savings"
	testSavingsAccountCurrency = "AED"

	testSalaryCategoryId   = ledger.CategoryId(1)
	testSalaryCategoryName = "Salary"

	testBillsCategoryId   = ledger.CategoryId(2)
	testBillsCategoryName = "Bills"

	testSavingsCategoryId   = ledger.CategoryId(3)
	testSavingsCategoryName = "Savings"
)

var (
	testRecordDate = time.Date(2021, time.July, 5, 18, 30, 0, 0, time.UTC)
)

type UserAndAccounts map[ledger.UserId][]ledger.AccountId

func (u UserAndAccounts) First() ledger.UserId {
	for userId := range u {
		return userId
	}
	return ledger.UserId(0)
}

func simulateRecords(db *sql.DB, numberOfUsers int, startMonth ledger.CalendarMonth, endMonth ledger.CalendarMonth) (UserAndAccounts, error) {
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	result := map[ledger.UserId][]ledger.AccountId{}

	for i := 0; i < numberOfUsers; i++ {
		// create user
		var user ledger.User
		userId := ledger.UserId(time.Now().UnixNano())
		if user, err = ledger.NewUserWithEmailString(userId, fmt.Sprintf("testUser+%d@gmail.com", userId)); err != nil {
			return nil, err
		}
		if err = UserDao.SaveTx(user, tx); err != nil {
			return nil, err
		}

		// create account (2 accounts per user)
		var (
			currentAccount ledger.Account
			savingsAccount ledger.Account
		)

		if currentAccount, err = ledger.NewAccount(ledger.AccountId(time.Now().UnixNano()), "Current", "AED", ledger.MustMakeUpdatedByUserId(userId)); err != nil {
			return nil, err
		}
		if savingsAccount, err = ledger.NewAccount(ledger.AccountId(time.Now().UnixNano()), "Savings", "AED", ledger.MustMakeUpdatedByUserId(userId)); err != nil {
			return nil, err
		}

		if err = AccountDao.SaveTx(context.Background(), userId, ledger.Accounts{currentAccount, savingsAccount}, tx); err != nil {
			return nil, err
		}

		// create categories
		var (
			salaryCategory       ledger.Category
			savingsCaegory       ledger.Category
			billsCategory        ledger.Category
			foodAndDrinkCategory ledger.Category
		)
		if salaryCategory, err = ledger.NewCategory(ledger.CategoryId(time.Now().UnixNano()), "Salary", ledger.MustMakeUpdatedByUserId(userId)); err != nil {
			return nil, err
		}
		if savingsCaegory, err = ledger.NewCategory(ledger.CategoryId(time.Now().UnixNano()), "Savings", ledger.MustMakeUpdatedByUserId(userId)); err != nil {
			return nil, err
		}
		if billsCategory, err = ledger.NewCategory(ledger.CategoryId(time.Now().UnixNano()), "Bills", ledger.MustMakeUpdatedByUserId(userId)); err != nil {
			return nil, err
		}
		if foodAndDrinkCategory, err = ledger.NewCategory(ledger.CategoryId(time.Now().UnixNano()), "Food & Drink", ledger.MustMakeUpdatedByUserId(userId)); err != nil {
			return nil, err
		}

		categories := ledger.Categories{salaryCategory, savingsCaegory, billsCategory, foodAndDrinkCategory}
		if err = CategoryDao.SaveTx(context.Background(), userId, categories, tx); err != nil {
			return nil, err
		}

		// create records
		fromDate := startMonth.FirstDay()
		toDate := endMonth.LastDay()

		for date := fromDate; date.Before(toDate); date = date.AddDate(0, 0, 1) {
			if date.Day() == 1 {
				var (
					salaryRecord           ledger.Record
					billsRecord            ledger.Record
					sendToSavingsRecord    ledger.Record
					receiveInSavingsRecord ledger.Record
				)
				// Add salary
				if salaryRecord, err = ledger.NewRecord(ledger.RecordId(time.Now().UnixNano()), "Salary", salaryCategory, quickMoney("AED", 50_000_00), date, ledger.Income, ledger.NoSourceAccount, ledger.NoBeneficiaryAccount, ledger.NoTransferReference, ledger.MustMakeUpdatedByUserId(userId)); err != nil {
					return nil, err
				}
				// Pay bills
				if billsRecord, err = ledger.NewRecord(ledger.RecordId(time.Now().UnixNano()), "Bills", billsCategory, quickMoney("AED", 1_000_00), date, ledger.Expense, ledger.NoSourceAccount, ledger.NoBeneficiaryAccount, ledger.NoTransferReference, ledger.MustMakeUpdatedByUserId(userId)); err != nil {
					return nil, err
				}
				// Save a little money for a rainy day
				ref := ledger.MakeTransferReference()
				if sendToSavingsRecord, err = ledger.NewRecord(ledger.RecordId(time.Now().UnixNano()), "Savings", savingsCaegory, quickMoney("AED", -10_000_00), date, ledger.Transfer, currentAccount.Id(), savingsAccount.Id(), ref, ledger.MustMakeUpdatedByUserId(userId)); err != nil {
					return nil, err
				}
				if receiveInSavingsRecord, err = ledger.NewRecord(ledger.RecordId(time.Now().UnixNano()), "Savings", savingsCaegory, quickMoney("AED", 10_000_00), date, ledger.Transfer, currentAccount.Id(), savingsAccount.Id(), ref, ledger.MustMakeUpdatedByUserId(userId)); err != nil {
					return nil, err
				}

				if err := RecordDao.SaveTx(context.Background(), currentAccount.Id(), salaryRecord, tx); err != nil {
					return nil, err
				}
				if err := RecordDao.SaveTx(context.Background(), currentAccount.Id(), billsRecord, tx); err != nil {
					return nil, err
				}
				if err := RecordDao.SaveTx(context.Background(), currentAccount.Id(), sendToSavingsRecord, tx); err != nil {
					return nil, err
				}
				if err := RecordDao.SaveTx(context.Background(), currentAccount.Id(), receiveInSavingsRecord, tx); err != nil {
					return nil, err
				}
			}

			// buy lunch and dinner every day
			var (
				lunchRecord  ledger.Record
				dinnerRecord ledger.Record
			)
			if lunchRecord, err = ledger.NewRecord(ledger.RecordId(time.Now().UnixNano()), "Lunch", foodAndDrinkCategory, quickMoney("AED", 10_00), date, ledger.Expense, ledger.NoSourceAccount, ledger.NoBeneficiaryAccount, ledger.NoTransferReference, ledger.MustMakeUpdatedByUserId(userId)); err != nil {
				return nil, err
			}
			if dinnerRecord, err = ledger.NewRecord(ledger.RecordId(time.Now().UnixNano()), "Dinner", foodAndDrinkCategory, quickMoney("AED", 20_00), date, ledger.Expense, ledger.NoSourceAccount, ledger.NoBeneficiaryAccount, ledger.NoTransferReference, ledger.MustMakeUpdatedByUserId(userId)); err != nil {
				return nil, err
			}
			if err := RecordDao.SaveTx(context.Background(), currentAccount.Id(), lunchRecord, tx); err != nil {
				return nil, err
			}
			if err := RecordDao.SaveTx(context.Background(), currentAccount.Id(), dinnerRecord, tx); err != nil {
				return nil, err
			}

			// eat big lunch on birthday (e.g. June 2)
			if date.Day() == 2 && date.Month() == time.June {
				var specialOccasionRecord ledger.Record
				if specialOccasionRecord, err = ledger.NewRecord(ledger.RecordId(time.Now().UnixNano()), "Birthday", foodAndDrinkCategory, quickMoney("AED", 100_00), date, ledger.Expense, ledger.NoSourceAccount, ledger.NoBeneficiaryAccount, ledger.NoTransferReference, ledger.MustMakeUpdatedByUserId(userId)); err != nil {
					return nil, err
				}
				if err := RecordDao.SaveTx(context.Background(), currentAccount.Id(), specialOccasionRecord, tx); err != nil {
					return nil, err
				}
			}
		}

		if err = tx.Commit(); err != nil {
			return nil, err
		}

		var count int
		_ = db.QueryRow("SELECT COUNT(*) FROM budget.record").Scan(&count)
		log.Printf("%d records inserted for simulation", count)

		result[userId] = []ledger.AccountId{currentAccount.Id(), savingsAccount.Id()}
	}
	return result, nil
}
