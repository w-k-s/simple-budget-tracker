package server

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	"github.com/lib/pq"
	"github.com/w-k-s/simple-budget-tracker/pkg/ledger"
	dao "github.com/w-k-s/simple-budget-tracker/pkg/persistence"
)

type DefaultCategoryDao struct {
	db *sql.DB
}

func MustOpenCategoryDao(driverName, dataSourceName string) dao.CategoryDao {
	var db *sql.DB
	var err error
	if db, err = sql.Open(driverName, dataSourceName); err != nil {
		log.Fatalf("Failed to connect to data source: %q with driver driver: %q. Reason: %s", dataSourceName, driverName, err)
	}
	return &DefaultCategoryDao{db}
}

func (d DefaultCategoryDao) Close() error {
	return d.db.Close()
}

func (d *DefaultCategoryDao) NewCategoryId() (ledger.CategoryId, error) {
	var categoryId ledger.CategoryId
	err := d.db.QueryRow("SELECT nextval('budget.category_id')").Scan(&categoryId)
	if err != nil {
		log.Printf("Failed to assign category id. Reason; %s", err)
		return 0, ledger.NewError(ledger.ErrDatabaseState, "Failed to assign category id", err)
	}
	return categoryId, err
}

func (d *DefaultCategoryDao) SaveTx(userId ledger.UserId, c ledger.Categories, tx *sql.Tx) error {
	checkError := func(err error) error {
		if err != nil {
			log.Printf("Failed to save categories '%q' for user id %d. Reason: %q", c, userId, err)
			if _, ok := isDuplicateKeyError(err); ok {
				message := fmt.Sprintf("Category names must be unique. One of these is duplicated: %s", strings.Join(c.Names(), ", "))
				if c.Len() == 1 {
					message = fmt.Sprintf("Category named %q already exists", c.Names()[0])
				}
				return ledger.NewError(ledger.ErrCategoryNameDuplicated, message, err)
			}
			return ledger.NewError(ledger.ErrDatabaseState, fmt.Sprintf("Failed to save categories %q", c.Names()), err)
		}
		return nil
	}

	stmt, err := tx.Prepare(pq.CopyInSchema("budget", "category", "id", "name", "user_id"))
	if err = checkError(err); err != nil {
		return err
	}

	for _, category := range c {
		_, err = stmt.Exec(category.Id(), category.Name(), userId)
		if err != nil {
			log.Printf("Failed to save category %q for user id %d. Reason: %q", category.Name(), userId, err)
		}

	}

	_, err = stmt.Exec()
	if err = checkError(err); err != nil {
		return err
	}

	err = stmt.Close()
	if err = checkError(err); err != nil {
		return err
	}
	return nil
}

func (d *DefaultCategoryDao) GetCategoriesForUser(userId ledger.UserId) (ledger.Categories, error) {

	rows, err := d.db.Query("SELECT c.id, c.name FROM budget.category c LEFT JOIN budget.user u ON c.user_id = u.id WHERE u.id = $1", userId)
	if err != nil {
		return nil, ledger.NewError(ledger.ErrCategoriesNotFound, fmt.Sprintf("Categories for user id %d not found", userId), err)
	}
	defer rows.Close()

	entities := make([]*ledger.Category, 0)
	for rows.Next() {
		var id ledger.CategoryId
		var name string

		if err := rows.Scan(&id, &name); err != nil {
			log.Printf("Error processign categories for user %d. Reason: %s", userId, err)
			continue
		}

		var category *ledger.Category
		if category, err = ledger.NewCategory(id, name); err != nil {
			log.Printf("Error loading category with id: %d,  name: %q from database. Reason: %s", id, name, err)
			continue
		}

		entities = append(entities, category)
	}

	return entities, nil
}

func (d *DefaultCategoryDao) Save(userId ledger.UserId, c ledger.Categories) error {
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	err = d.SaveTx(userId, c, tx)
	if err == nil {
		err = tx.Commit()
	}
	return err
}
