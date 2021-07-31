package test

import (
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

func simulateRecords(db *sql.DB, numberOfUsers int, startMonth ledger.CalendarMonth, endMonth ledger.CalendarMonth) ([]ledger.AccountId, error) {
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	currentAccountIds := make([]ledger.AccountId, 0, numberOfUsers*10)
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

		if err = AccountDao.SaveTx(userId, &currentAccount, tx); err != nil {
			return nil, err
		}

		if err = AccountDao.SaveTx(userId, &savingsAccount, tx); err != nil {
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
		if err = CategoryDao.SaveTx(userId, categories, tx); err != nil {
			return nil, err
		}

		// create records
		fromDate := startMonth.FirstDay()
		toDate := endMonth.LastDay()

		for date := fromDate; date.Before(toDate); date = date.AddDate(0, 0, 1) {
			if date.Day() == 1 {
				var (
					salaryRecord  ledger.Record
					billsRecord   ledger.Record
					savingsRecord ledger.Record
				)
				// Add salary
				if salaryRecord, err = ledger.NewRecord(ledger.RecordId(time.Now().UnixNano()), "Salary", salaryCategory, quickMoney("AED", 50_000_00), date, ledger.Income, 0, ledger.MustMakeUpdatedByUserId(userId)); err != nil {
					return nil, err
				}
				// Pay bills
				if billsRecord, err = ledger.NewRecord(ledger.RecordId(time.Now().UnixNano()), "Bills", billsCategory, quickMoney("AED", 1_000_00), date, ledger.Expense, 0, ledger.MustMakeUpdatedByUserId(userId)); err != nil {
					return nil, err
				}
				// Save a little money for a rainy day
				if savingsRecord, err = ledger.NewRecord(ledger.RecordId(time.Now().UnixNano()), "Savings", savingsCaegory, quickMoney("AED", 10_000_00), date, ledger.Transfer, savingsAccount.Id(), ledger.MustMakeUpdatedByUserId(userId)); err != nil {
					return nil, err
				}

				if err := RecordDao.SaveTx(currentAccount.Id(), &salaryRecord, tx); err != nil {
					return nil, err
				}
				if err := RecordDao.SaveTx(currentAccount.Id(), &billsRecord, tx); err != nil {
					return nil, err
				}
				if err := RecordDao.SaveTx(currentAccount.Id(), &savingsRecord, tx); err != nil {
					return nil, err
				}
			}

			// buy lunch and dinner every day
			var (
				lunchRecord  ledger.Record
				dinnerRecord ledger.Record
			)
			if lunchRecord, err = ledger.NewRecord(ledger.RecordId(time.Now().UnixNano()), "Lunch", foodAndDrinkCategory, quickMoney("AED", 10_00), date, ledger.Expense, 0, ledger.MustMakeUpdatedByUserId(userId)); err != nil {
				return nil, err
			}
			if dinnerRecord, err = ledger.NewRecord(ledger.RecordId(time.Now().UnixNano()), "Dinner", foodAndDrinkCategory, quickMoney("AED", 20_00), date, ledger.Expense, 0, ledger.MustMakeUpdatedByUserId(userId)); err != nil {
				return nil, err
			}
			if err := RecordDao.SaveTx(currentAccount.Id(), &lunchRecord, tx); err != nil {
				return nil, err
			}
			if err := RecordDao.SaveTx(currentAccount.Id(), &dinnerRecord, tx); err != nil {
				return nil, err
			}

			// eat big lunch on birthday (e.g. June 2)
			if date.Day() == 2 && date.Month() == time.June {
				var specialOccasionRecord ledger.Record
				if specialOccasionRecord, err = ledger.NewRecord(ledger.RecordId(time.Now().UnixNano()), "Birthday", foodAndDrinkCategory, quickMoney("AED", 100_00), date, ledger.Expense, 0, ledger.MustMakeUpdatedByUserId(userId)); err != nil {
					return nil, err
				}
				if err := RecordDao.SaveTx(currentAccount.Id(), &specialOccasionRecord, tx); err != nil {
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

		currentAccountIds = append(currentAccountIds, currentAccount.Id())
	}
	return currentAccountIds, nil
}
