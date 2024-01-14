package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/lib/pq"
	"github.com/w-k-s/simple-budget-tracker/pkg/ledger"
	dao "github.com/w-k-s/simple-budget-tracker/pkg/persistence"
)

type DefaultBudgetDao struct {
	*RootDao
}

func MustOpenBudgetDao(db *sql.DB) dao.BudgetDao {
	return &DefaultBudgetDao{&RootDao{db}}
}

func (d *DefaultBudgetDao) Save(
	ctx context.Context,
	userId ledger.UserId,
	b ledger.Budget,
	tx *sql.Tx,
) error {
	epoch := time.Time{}
	_, err := tx.ExecContext(
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
		return fmt.Errorf("Faiiled to save budget. Reason: %w", err)
	}

	log.Printf("insrrted budget")
	stmt, err := tx.PrepareContext(
		ctx,
		pq.CopyInSchema(
			"budget",
			"budget_per_category",
			"budget_id",
			"category_id",
			"currency",
			"amount_minor_units",
		))
	if err != nil {
		return fmt.Errorf("Failed to prepare bulk statement for budget per category. Reason: %w", err)
	}
	defer stmt.Close()
	log.Printf("created cb stmts")

	for _, cb := range b.CategoryBudgets() {
		_, err = stmt.ExecContext(
			ctx,
			b.Id(),
			cb.CategoryId(),
			cb.MaxLimit().Currency().CurrencyCode(),
			cb.MaxLimit().MustMinorUnits(),
		)
		if err != nil {
			return fmt.Errorf("Failed to save category budget %q for user id %d. Reason: %w", cb, userId, err)
		}
	}
	if _, err := stmt.ExecContext(ctx); err != nil {
		return fmt.Errorf("Failed to flush category budgets for user id %d. Reason: %w", userId, err)
	}

	stmt, err = tx.PrepareContext(
		ctx,
		pq.CopyInSchema(
			"budget",
			"account_budgets",
			"account_id",
			"budget_id",
		),
	)
	if err != nil {
		return fmt.Errorf("Failed to prepare bulk statement for account budget: %w", err)
	}
	log.Printf("created ab stmts")

	for _, accountId := range b.AccountIds() {
		_, err = stmt.ExecContext(
			ctx,
			accountId,
			b.Id(),
		)
		if err != nil {
			return fmt.Errorf("Failed to save budget %q for account id %d. Reason: %w", b, accountId, err)
		}
	}
	_, err = stmt.ExecContext(ctx)
	log.Printf("exec ab stmts")

	return nil
}

func (d *DefaultBudgetDao) GetBudgetById(
	ctx context.Context,
	id ledger.BudgetId,
	userId ledger.UserId,
	tx *sql.Tx,
) (ledger.Budget, error) {

	rows, err := tx.QueryContext(
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
		return ledger.Budget{}, fmt.Errorf("Query execution failed", err)
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
			&accountId,
			&br.periodType,
			&categoryId,
			&categoryName,
			&currency,
			&amountMinorUnits,
			&br.createdBy,
			&br.createdAt,
			&br.modifiedBy,
			&br.modifiedAt,
			&br.version,
		); err != nil {
			return ledger.Budget{}, fmt.Errorf("Failed to scan row. Reason: %w", err)
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
