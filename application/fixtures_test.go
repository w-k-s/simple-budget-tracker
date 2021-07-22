package application

import (
	"database/sql"
	"fmt"
	"log"
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

	testSavingsCategoryId   = core.CategoryId(3)
	testSavingsCategoryName = "Savings"
)

var (
	testRecordDate = time.Date(2021, time.July, 5, 18, 30, 0, 0, time.UTC)
)

func simulateRecords(db *sql.DB, numberOfUsers int, startMonth core.CalendarMonth, endMonth core.CalendarMonth) ([]core.AccountId, error) {
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	currentAccountIds := make([]core.AccountId, 0, numberOfUsers*10)
	for i := 0; i < numberOfUsers; i++ {
		// create user
		var user *core.User
		userId := core.UserId(time.Now().UnixNano())
		if user, err = core.NewUserWithEmailString(userId, fmt.Sprintf("testUser+%d@gmail.com", userId)); err != nil {
			return nil, err
		}
		if err = UserDao.SaveTx(user, tx); err != nil {
			return nil, err
		}

		// create account (2 accounts per user)
		var (
			currentAccount *core.Account
			savingsAccount *core.Account
		)

		if currentAccount, err = core.NewAccount(core.AccountId(time.Now().UnixNano()), "Current", "AED"); err != nil {
			return nil, err
		}
		if savingsAccount, err = core.NewAccount(core.AccountId(time.Now().UnixNano()), "Savings", "AED"); err != nil {
			return nil, err
		}

		if err = AccountDao.SaveTx(userId, currentAccount, tx); err != nil {
			return nil, err
		}

		if err = AccountDao.SaveTx(userId, savingsAccount, tx); err != nil {
			return nil, err
		}

		// create categories
		var (
			salaryCategory       *core.Category
			savingsCaegory       *core.Category
			billsCategory        *core.Category
			foodAndDrinkCategory *core.Category
		)
		if salaryCategory, err = core.NewCategory(core.CategoryId(time.Now().UnixNano()), "Salary"); err != nil {
			return nil, err
		}
		if savingsCaegory, err = core.NewCategory(core.CategoryId(time.Now().UnixNano()), "Savings"); err != nil {
			return nil, err
		}
		if billsCategory, err = core.NewCategory(core.CategoryId(time.Now().UnixNano()), "Bills"); err != nil {
			return nil, err
		}
		if foodAndDrinkCategory, err = core.NewCategory(core.CategoryId(time.Now().UnixNano()), "Food & Drink"); err != nil {
			return nil, err
		}

		categories := core.Categories{salaryCategory, savingsCaegory, billsCategory, foodAndDrinkCategory}
		if err = CategoryDao.SaveTx(userId, categories, tx); err != nil {
			return nil, err
		}

		// create records
		fromDate := startMonth.FirstDay()
		toDate := endMonth.LastDay()

		for date := fromDate; date.Before(toDate); date = date.AddDate(0, 0, 1) {
			if date.Day() == 1 {
				var (
					salaryRecord  *core.Record
					billsRecord   *core.Record
					savingsRecord *core.Record
				)
				// Add salary
				if salaryRecord, err = core.NewRecord(core.RecordId(time.Now().UnixNano()), "Salary", salaryCategory, quickMoney("AED", 50_000_00), date, core.Income, 0); err != nil {
					return nil, err
				}
				// Pay bills
				if billsRecord, err = core.NewRecord(core.RecordId(time.Now().UnixNano()), "Bills", billsCategory, quickMoney("AED", 1_000_00), date, core.Expense, 0); err != nil {
					return nil, err
				}
				// Save a little money for a rainy day
				if savingsRecord, err = core.NewRecord(core.RecordId(time.Now().UnixNano()), "Savings", savingsCaegory, quickMoney("AED", 10_000_00), date, core.Transfer, savingsAccount.Id()); err != nil {
					return nil, err
				}

				if err := RecordDao.SaveTx(currentAccount.Id(), salaryRecord, tx); err != nil {
					return nil, err
				}
				if err := RecordDao.SaveTx(currentAccount.Id(), billsRecord, tx); err != nil {
					return nil, err
				}
				if err := RecordDao.SaveTx(currentAccount.Id(), savingsRecord, tx); err != nil {
					return nil, err
				}
			}

			// buy lunch and dinner every day
			var (
				lunchRecord  *core.Record
				dinnerRecord *core.Record
			)
			if lunchRecord, err = core.NewRecord(core.RecordId(time.Now().UnixNano()), "Lunch", foodAndDrinkCategory, quickMoney("AED", 10_00), date, core.Expense, 0); err != nil {
				return nil, err
			}
			if dinnerRecord, err = core.NewRecord(core.RecordId(time.Now().UnixNano()), "Dinner", foodAndDrinkCategory, quickMoney("AED", 20_00), date, core.Expense, 0); err != nil {
				return nil, err
			}
			if err := RecordDao.SaveTx(currentAccount.Id(), lunchRecord, tx); err != nil {
				return nil, err
			}
			if err := RecordDao.SaveTx(currentAccount.Id(), dinnerRecord, tx); err != nil {
				return nil, err
			}

			// eat big lunch on birthday (e.g. June 2)
			if date.Day() == 2 && date.Month() == time.June {
				var specialOccasionRecord *core.Record
				if specialOccasionRecord, err = core.NewRecord(core.RecordId(time.Now().UnixNano()), "Birthday", foodAndDrinkCategory, quickMoney("AED", 100_00), date, core.Expense, 0); err != nil {
					return nil, err
				}
				if err := RecordDao.SaveTx(currentAccount.Id(), specialOccasionRecord, tx); err != nil {
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
