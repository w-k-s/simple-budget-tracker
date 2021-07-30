package ledger

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/gobuffalo/validate"
	"github.com/gobuffalo/validate/validators"
)

type CategoryId uint64

type Category struct {
	AuditInfo
	id   CategoryId
	name string
}

type CategoryRecord interface {
	Id() CategoryId
	Name() string
	CreatedBy() UserId
	CreatedAtUTC() time.Time
	ModifiedBy() UserId
	ModifiedAtUTC() time.Time
	Version() Version
}

func NewCategory(id CategoryId, userId UserId, name string) (Category, error) {
	var (
		auditInfo AuditInfo
		err       error
	)
	if auditInfo, err = MakeAuditForCreation(userId); err != nil {
		return Category{}, err
	}
	return newCategory(id, name, auditInfo)
}

func NewCategoryFromRecord(cr CategoryRecord) (Category, error) {
	var (
		auditInfo AuditInfo
		err       error
	)
	if auditInfo, err = MakeAuditForModification(
		cr.CreatedBy(),
		cr.CreatedAtUTC(),
		cr.ModifiedBy(),
		cr.ModifiedAtUTC(),
		cr.Version(),
	); err != nil {
		return Category{}, err
	}
	return newCategory(cr.Id(), cr.Name(), auditInfo)
}

func newCategory(id CategoryId, name string, auditInfo AuditInfo) (Category, error) {
	errors := validate.Validate(
		&validators.IntIsGreaterThan{Name: "Id", Field: int(id), Compared: 0, Message: "Id must be greater than 0"},
		&validators.StringLengthInRange{Name: "Name", Field: name, Min: 1, Max: 25, Message: "Name must be 1 and 25 characters long"},
	)

	var err error
	if err = makeCoreValidationError(ErrCategoryValidation, errors); err != nil {
		return Category{}, err
	}

	return Category{
		AuditInfo: auditInfo,
		id:        id,
		name:      name,
	}, nil
}

func (c Category) Id() CategoryId {
	return c.id
}

func (c Category) Name() string {
	return c.name
}

func (c Category) String() string {
	return fmt.Sprintf("Category{id: %d, name: %s}", c.id, c.name)
}

type Categories []Category

func (cs Categories) Names() []string {
	names := make([]string, 0, len(cs))
	for _, category := range cs {
		names = append(names, category.Name())
	}
	sort.Strings(names)
	return names
}

func (cs Categories) String() string {
	sort.Sort(cs)
	strs := make([]string, 0, len(cs))
	for _, category := range cs {
		strs = append(strs, category.String())
	}
	return fmt.Sprintf("Categories{%s}", strings.Join(strs, ", "))
}

func (c Categories) Len() int           { return len(c) }
func (c Categories) Swap(i, j int)      { c[i], c[j] = c[j], c[i] }
func (c Categories) Less(i, j int) bool { return c[i].Name() < c[j].Name() }
