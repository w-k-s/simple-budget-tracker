package core

import (
	"fmt"
	"sort"
	"strings"

	"github.com/gobuffalo/validate"
	"github.com/gobuffalo/validate/validators"
)

type CategoryId uint64

type Category struct {
	id   CategoryId
	name string
}

func NewCategory(id CategoryId, name string) (*Category, error) {
	category := &Category{
		id:   id,
		name: name,
	}

	errors := validate.Validate(
		&validators.IntIsGreaterThan{Name: "Id", Field: int(category.id), Compared: 0, Message: "Id must be greater than 0"},
		&validators.StringLengthInRange{Name: "Name", Field: category.name, Min: 1, Max: 25, Message: "Name must be 1 and 25 characters long"},
	)

	if errors.HasAny() {
		flatErrors := map[string]string{}
		for field, violations := range errors.Errors {
			flatErrors[field] = strings.Join(violations, ", ")
		}
		listErrors := []string{}
		for _, violations := range flatErrors {
			listErrors = append(listErrors, violations)
		}
		return nil, NewErrorWithFields(ErrCategoryValidation, strings.Join(listErrors, ", "), errors, flatErrors)
	}
	return category, nil
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

type Categories []*Category

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
