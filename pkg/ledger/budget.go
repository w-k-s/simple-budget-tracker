package ledger

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/gobuffalo/validate"
	"github.com/gobuffalo/validate/validators"
	"github.com/w-k-s/simple-budget-tracker/pkg"
)

type BudgetId uint64

type BudgetPeriodType string

const (
	BudgetPeriodTypeWeek  BudgetPeriodType = "Week"
	BudgetPeriodTypeMonth BudgetPeriodType = "Month"
)

type budgetPeriodTypeValidator struct {
	Name  string
	Field string
}

func (v *budgetPeriodTypeValidator) IsValid(errors *validate.Errors) {
	if len(v.Field) == 0 {
		errors.Add(strings.ToLower(v.Name), "periodType is required")
		return
	}
	validPeriodTypes := []string{
		string(BudgetPeriodTypeWeek),
		string(BudgetPeriodTypeMonth),
	}

	validator := &validators.StringInclusion{
		Name:    v.Name,
		Field:   v.Field,
		List:    validPeriodTypes,
		Message: fmt.Sprintf("periodType must be one of %q", validPeriodTypes),
	}
	validator.IsValid(errors)
}

type CategoryBudget struct {
	categoryId CategoryId
	// The maximum amount allowed to be spent for the associated category in a time period.
	maxLimit Money
	// The actual amount spent for the associated category in a time period.
	amountSpent Money
}

func NewCategoryBudget(
	categoryId CategoryId,
	maxLimit Money,
	amountSpent Money,
) (CategoryBudget, error) {

	errors := validate.Validate(
		&validators.IntIsGreaterThan{
			Name:     "Id",
			Field:    int(categoryId),
			Compared: 0,
			Message:  "CategoryId must be greater than 0",
		},
		&amountPositiveOrZeroValidator{
			Name:    "maxLimit",
			Field:   maxLimit,
			Message: "MaxLimit must be greater than or equal to 0",
		},
		&amountPositiveOrZeroValidator{
			Name:    "amountSpent",
			Field:   maxLimit,
			Message: "amountSpent must be greater than or equal to 0",
		},
	)

	err := pkg.ValidationErrorWithErrors(pkg.ErrBudgetValidation, "", errors)
	if err != nil {
		return CategoryBudget{}, err
	}

	return CategoryBudget{
		categoryId:  categoryId,
		maxLimit:    maxLimit,
		amountSpent: amountSpent,
	}, nil
}

func MustCategoryBudget(cb CategoryBudget, err error) CategoryBudget {
	if err != nil {
		log.Fatalf("Failed to create category budget. Reason: %s", err)
	}
	return cb
}

func (cb CategoryBudget) CategoryId() CategoryId {
	return cb.categoryId
}

func (cb CategoryBudget) MaxLimit() Money {
	return cb.maxLimit
}

func (cb CategoryBudget) AmountSpent() Money {
	return cb.amountSpent
}

func (cb CategoryBudget) Exceeded() bool {
	return cb.amountSpent.MustMinorUnits() > cb.maxLimit.MustMinorUnits()
}

func (cb CategoryBudget) String() string {
	return fmt.Sprintf("CategoryBudget{Category: %d, Max: %s, Spent: %s, Exceeded: %t}",
		cb.categoryId,
		cb.maxLimit,
		cb.amountSpent,
		cb.Exceeded(),
	)
}

type CategoryBudgets []CategoryBudget

func (a CategoryBudgets) Len() int           { return len(a) }
func (a CategoryBudgets) Less(i, j int) bool { return a[i].categoryId < a[j].categoryId }
func (a CategoryBudgets) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (cb CategoryBudgets) String() string {
	sb := strings.Builder{}
	sb.WriteString("CategoryBudgets{")
	for i, c := range cb {
		sb.WriteString(strconv.FormatUint(uint64(c.categoryId), 10))
		sb.WriteString(": ")
		sb.WriteString(c.amountSpent.String())
		sb.WriteString(" / ")
		sb.WriteString(c.maxLimit.String())
		if i != len(cb)-1 {
			sb.WriteString(", ")
		}
	}
	sb.WriteString("}")
	return sb.String()
}

type BudgetRecord interface {
	Id() BudgetId
	AccountIds() AccountIds
	PeriodType() BudgetPeriodType
	CategoryBudgets() CategoryBudgets
	CreatedBy() UpdatedBy
	CreatedAtUTC() time.Time
	ModifiedBy() UpdatedBy
	ModifiedAtUTC() time.Time
	Version() Version
}

type Budget struct {
	auditInfo
	// The id of the budget
	id BudgetId
	// The accounts that this budget applies to.
	// For example: Let's say a budget of no more than a $100 per month on shopping sets accountIds to Account A and B.
	// The budget will be exceeded when the total spent on shopping by the two accounts combined is greater than $100.
	accountIds AccountIds
	// Whether the budget limits the amount spent in a week or the amount spent in a month.
	periodType      BudgetPeriodType
	categoryBudgets CategoryBudgets
}

func (b Budget) Id() BudgetId {
	return b.id
}

func (b Budget) AccountIds() AccountIds {
	return b.accountIds
}

func (b Budget) PeriodType() BudgetPeriodType {
	return b.periodType
}

func (b Budget) CategoryBudgets() CategoryBudgets {
	return b.categoryBudgets
}

func (b Budget) CreatedBy() UpdatedBy {
	return b.createdBy
}

func (b Budget) CreatedAtUTC() time.Time {
	return b.createdAtUTC
}

func (b Budget) ModifiedBy() UpdatedBy {
	return b.modifiedBy
}

func (b Budget) ModifiedAtUTC() time.Time {
	return b.modifiedAtUTC
}

func (b Budget) Version() Version {
	return b.version
}

func (b Budget) String() string {
	return fmt.Sprintf("Budget{id: %d, accountIds: %v, periodType: %s, categoryBudgets: %s}",
		b.id,
		b.accountIds,
		b.periodType,
		b.categoryBudgets,
	)
}

func NewBudget(
	id BudgetId,
	accountIds AccountIds,
	periodType BudgetPeriodType,
	categoryBudgets CategoryBudgets,
	createdBy UpdatedBy,
) (Budget, error) {

	auditInfo, err := makeAuditForCreation(createdBy)
	if err != nil {
		return Budget{}, err
	}

	return newBudget(
		id,
		accountIds,
		periodType,
		categoryBudgets,
		auditInfo,
	)
}

func NewBudgetFromRecord(record BudgetRecord) (Budget, error) {

	auditInfo, err := makeAuditForModification(
		record.CreatedBy(),
		record.CreatedAtUTC(),
		record.ModifiedBy(),
		record.ModifiedAtUTC(),
		record.Version(),
	)
	if err != nil {
		return Budget{}, err
	}

	return newBudget(
		record.Id(),
		record.AccountIds(),
		record.PeriodType(),
		record.CategoryBudgets(),
		auditInfo,
	)
}

func newBudget(
	id BudgetId,
	accountIds AccountIds,
	periodType BudgetPeriodType,
	categoryBudgets CategoryBudgets,
	auditInfo auditInfo,
) (Budget, error) {

	errors := validate.Validate(
		&validators.IntIsGreaterThan{
			Name:     "Id",
			Field:    int(id),
			Compared: 0,
			Message:  "Id must be greater than 0",
		},
		&validators.IntIsGreaterThan{
			Name:     "accountIds",
			Field:    len(accountIds),
			Compared: 0,
			Message:  "AccountIds must not be empty",
		},
		&budgetPeriodTypeValidator{
			Name:  "periodType",
			Field: string(periodType),
		},
		&validators.IntIsGreaterThan{
			Name:     "categoryBudgets",
			Field:    len(categoryBudgets),
			Compared: 0,
			Message:  "Category Budgets can not be empty",
		},
	)

	err := pkg.ValidationErrorWithErrors(pkg.ErrBudgetValidation, "", errors)
	if err != nil {
		return Budget{}, err
	}

	return Budget{
		auditInfo:       auditInfo,
		id:              id,
		accountIds:      accountIds,
		periodType:      periodType,
		categoryBudgets: categoryBudgets,
	}, nil
}
