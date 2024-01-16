package persistence

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/w-k-s/simple-budget-tracker/pkg"
	"github.com/w-k-s/simple-budget-tracker/pkg/ledger"
	dao "github.com/w-k-s/simple-budget-tracker/pkg/persistence"
)

type DefaultUserDao struct {
	*RootDao
}

func MustOpenUserDao(db *sql.DB) dao.UserDao {
	return &DefaultUserDao{
		&RootDao{db},
	}
}

func (d *DefaultUserDao) NewUserId() (ledger.UserId, error) {
	var userId ledger.UserId
	err := d.db.QueryRow("SELECT nextval('budget.user_id')").Scan(&userId)
	if err != nil {
		log.Printf("Failed to assign user id. Reason; %s", err)
		return 0, fmt.Errorf("Failed to assign user id. Reason: %w", err)
	}
	return userId, err
}

func (d *DefaultUserDao) SaveTx(u ledger.User, tx *sql.Tx) error {
	epoch := time.Time{}
	_, err := tx.Exec(
		"INSERT INTO budget.user (id, email, created_by, created_at, last_modified_by, last_modified_at, version) VALUES ($1, $2, $3, $4, $5, $6, $7)",
		u.Id(),
		u.Email().Address,
		u.CreatedBy().String(),
		u.CreatedAtUTC(),
		sql.NullString{
			String: u.ModifiedBy().String(),
			Valid:  u.ModifiedBy() != ledger.UpdatedBy{},
		},
		sql.NullTime{
			Time:  u.ModifiedAtUTC(),
			Valid: epoch != u.ModifiedAtUTC(),
		},
		u.Version(),
	)
	if err != nil {
		return fmt.Errorf("Failed to save user. Reason: %w", err)
	}
	return nil
}

func (d *DefaultUserDao) GetUserById(queryId ledger.UserId) (ledger.User, error) {
	var ur userRecord
	err := d.db.QueryRow("SELECT id, email, created_by, created_at, last_modified_by, last_modified_at, version FROM budget.user WHERE id = $1", queryId).
		Scan(&ur.id, &ur.email, &ur.createdBy, &ur.createdAt, &ur.modifiedBy, &ur.modifiedAt, &ur.version)

	if err == sql.ErrNoRows {
		return ledger.User{}, pkg.ValidationErrorWithError(pkg.ErrUserNotFound, fmt.Sprintf("User with id %d not found", queryId), err)
	} else if err != nil {
		return ledger.User{}, fmt.Errorf("User with id %d not found. Reason: %w", queryId, err)
	}

	return ledger.NewUserFromRecord(ur)
}

func (d *DefaultUserDao) Save(u ledger.User) error {
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	err = d.SaveTx(u, tx)
	if err == nil {
		err = tx.Commit()
	}
	return err
}
