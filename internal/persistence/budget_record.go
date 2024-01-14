package persistence

import (
	"database/sql"
	"log"
	"time"

	"github.com/w-k-s/simple-budget-tracker/pkg/ledger"
)

type budgetRecord struct {
	id              ledger.BudgetId
	accountIds      ledger.AccountIds
	periodType      ledger.BudgetPeriodType
	categoryBudgets ledger.CategoryBudgets
	createdBy       string
	createdAt       time.Time
	modifiedBy      sql.NullString
	modifiedAt      sql.NullTime
	version         ledger.Version
}

func (br budgetRecord) Id() ledger.BudgetId {
	return br.id
}

func (br budgetRecord) AccountIds() ledger.AccountIds {
	return br.accountIds
}

func (br budgetRecord) PeriodType() ledger.BudgetPeriodType {
	return br.periodType
}

func (br budgetRecord) CategoryBudgets() ledger.CategoryBudgets {
	return br.categoryBudgets
}

func (br budgetRecord) CreatedBy() ledger.UpdatedBy {
	updatedBy, err := ledger.ParseUpdatedBy(br.createdBy)
	if err != nil {
		log.Fatalf("Invalid createdBy persisted for record %d: %s", br.id, br.createdBy)
	}
	return updatedBy
}

func (br budgetRecord) CreatedAtUTC() time.Time {
	return br.createdAt
}

func (br budgetRecord) ModifiedBy() ledger.UpdatedBy {
	if !br.modifiedBy.Valid {
		return ledger.UpdatedBy{}
	}
	var (
		updatedBy ledger.UpdatedBy
		err       error
	)
	if updatedBy, err = ledger.ParseUpdatedBy(br.modifiedBy.String); err != nil {
		log.Fatalf("Invalid modifiedBy persisted for record %d: %s", br.id, br.ModifiedBy())
	}
	return updatedBy
}

func (br budgetRecord) ModifiedAtUTC() time.Time {
	if br.modifiedAt.Valid {
		return br.modifiedAt.Time
	}
	return time.Time{}
}

func (br budgetRecord) Version() ledger.Version {
	return br.version
}
