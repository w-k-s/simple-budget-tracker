package ledger

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/gobuffalo/validate"
	"github.com/gobuffalo/validate/validators"
	"github.com/w-k-s/simple-budget-tracker/pkg"
)

type CategoryId uint64

type Category struct {
	auditInfo
	id   CategoryId
	name string
}

type CategoryRecord interface {
	Id() CategoryId
	Name() string
	CreatedBy() UpdatedBy
	CreatedAtUTC() time.Time
	ModifiedBy() UpdatedBy
	ModifiedAtUTC() time.Time
	Version() Version
}

func NewCategory(id CategoryId, name string, updatedBy UpdatedBy) (Category, error) {
	var (
		auditInfo auditInfo
		err       error
	)
	if auditInfo, err = makeAuditForCreation(updatedBy); err != nil {
		return Category{}, err
	}
	return newCategory(id, name, auditInfo)
}

func NewCategoryFromRecord(cr CategoryRecord) (Category, error) {
	var (
		auditInfo auditInfo
		err       error
	)
	if auditInfo, err = makeAuditForModification(
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

func newCategory(id CategoryId, name string, auditInfo auditInfo) (Category, error) {
	errors := validate.Validate(
		&validators.IntIsGreaterThan{Name: "Id", Field: int(id), Compared: 0, Message: "Id must be greater than 0"},
		&validators.StringLengthInRange{Name: "Name", Field: name, Min: 1, Max: 25, Message: "Name must be 1 and 25 characters long"},
	)

	var err error
	if err = pkg.ValidationErrorWithErrors(pkg.ErrCategoryValidation, "", errors); err != nil {
		return Category{}, err
	}

	return Category{
		auditInfo: auditInfo,
		id:        id,
		name:      strings.Title(strings.ToLower(name)),
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

func (cs Categories) MayById() map[CategoryId]Category{
	m := map[CategoryId]Category{}
	for _,c := range cs{
		m[c.id] = c
	}
	return m
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
