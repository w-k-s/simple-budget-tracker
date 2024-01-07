package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/lib/pq"
	"github.com/w-k-s/simple-budget-tracker/pkg"
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
	var (
		updatedBy ledger.UpdatedBy
		err       error
	)
	if updatedBy, err = ledger.ParseUpdatedBy(br.createdBy); err != nil {
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

type defaultBudgetDao struct {
	tx *sql.Tx
}

func (d *defaultBudgetDao) Save(
	ctx context.Context,
	userId ledger.UserId,
	b ledger.Budget,
) error {
	epoch := time.Time{}
	_, err := d.tx.ExecContext(
		ctx,
		`INSERT INTO budget.budget (
			id, 
			user_id,
			period, 
			created_by, 
			created_at, 
			last_modified_by, 
			last_modified_at, 
			version
		) VALUES (
			$1, 
			$2, 
			$3, 
			$4, 
			$5,
			$6, 
			$7,
			$8
		)`,
		b.Id(),
		userId,
		b.PeriodType(),
		b.CreatedBy().String(),
		b.CreatedAtUTC(),
		sql.NullString{
			String: b.ModifiedBy().String(),
			Valid:  b.ModifiedBy() != ledger.UpdatedBy{},
		},
		sql.NullTime{
			Time:  b.ModifiedAtUTC(),
			Valid: epoch != b.ModifiedAtUTC(),
		},
		b.Version(),
	)
	if err != nil {
		log.Printf("Failed to save budget %v. Reason: %s", b, err)
		return pkg.NewSystemError(pkg.ErrDatabaseState, "Failed to save budget", err)
	}

	stmt, err := d.tx.PrepareContext(
		ctx,
		pq.CopyInSchema(
			"budget_per_category",
			"budget_id",
			"category_id",
			"amount_minor_units",
		))
	if err != nil {
		return pkg.NewSystemError(pkg.ErrDatabaseState, "Failed to prepare bulk statement for budget per category", err)
	}

	for _, cb := range b.CategoryBudgets() {
		_, err = stmt.ExecContext(
			ctx,
			b.Id(),
			cb.CategoryId(),
			cb.MaxLimit().MustMinorUnits(),
		)
		if err != nil {
			log.Printf("Failed to save category budget %q for user id %d. Reason: %q", cb, userId, err)
		}
	}

	stmt, err = d.tx.PrepareContext(
		ctx,
		pq.CopyInSchema(
			"account_budgets",
			"account_id",
			"budget_id",
		))
	if err != nil {
		return pkg.NewSystemError(pkg.ErrDatabaseState, "Failed to prepare bulk statement for account budget", err)
	}

	for _, accountId := range b.AccountIds() {
		_, err = stmt.ExecContext(
			ctx,
			accountId,
			b.Id(),
		)
		if err != nil {
			log.Printf("Failed to save budget %q for account id %d. Reason: %q", b, accountId, err)
		}
	}

	return nil
}

func (d *defaultBudgetDao) GetBudgetById(
	ctx context.Context,
	id ledger.BudgetId,
	userId ledger.UserId,
) (ledger.Budget, error) {

	rows, err := d.tx.QueryContext(
		ctx,
		`SELECT 
			b.id,
			a.account_id,
			b.period, 
			bc.category_id,
			c.name,
			bc.currency,
			bc.amount_minor_units,
			b.created_by,
			b.created_at,
			b.last_modified_by,
			b.last_modified_at,
			b.version
		FROM 
			budget.budget b 
		LEFT JOIN budget.user u ON b.user_id = u.id 
		LEFT JOIN budget.account_budgets a ON a.budget_id = b.id
		LEFT JOIN budget.budget_per_category bc ON bc.budget_id = b.id
		LEFT JOIN budget.category c ON bc.category_id = c.id
		WHERE 
			u.id = $1
		AND
			b.id = $2`,
		userId,
		id,
	)
	if err != nil {
		return ledger.Budget{}, pkg.NewSystemError(pkg.ErrDatabaseState, "Query execution failed", err)
	}
	defer rows.Close()

	br := budgetRecord{}

	for rows.Next() {
		accountId := ledger.AccountId(0)
		categoryId := uint64(0)
		categoryName := ""
		currency := ""
		amountMinorUnits := int64(0)

		if err := rows.Scan(
			&br.id,
			&br.periodType,
			&categoryId,
			&categoryName,
			&currency,
			&amountMinorUnits,
			&accountId,
			&br.createdBy,
			&br.createdAt,
			&br.modifiedBy,
			&br.modifiedAt,
			&br.version,
		); err != nil {
			return ledger.Budget{}, pkg.NewSystemError(pkg.ErrCategoriesNotFound, fmt.Sprintf("Categories for user id %d not found", userId), err)
		}

		// TODO: remove duplicate accountIds
		br.accountIds = append(br.accountIds, accountId)
		br.categoryBudgets = append(
			br.categoryBudgets,
			ledger.MustCategoryBudget(ledger.NewCategoryBudget(
				ledger.CategoryId(categoryId),
				ledger.MustMoney(ledger.NewMoney(currency, amountMinorUnits)),
			)))
	}

	return ledger.NewBudgetFromRecord(br)
}
