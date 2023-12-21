package persistence

import (
	"database/sql"
	"fmt"
	"log"
	"net/mail"
	"time"

	"github.com/w-k-s/simple-budget-tracker/pkg/ledger"
	dao "github.com/w-k-s/simple-budget-tracker/pkg/persistence"
)

type userRecord struct {
	id         ledger.UserId
	email      string
	createdBy  string
	createdAt  time.Time
	modifiedBy sql.NullString
	modifiedAt sql.NullTime
	version    ledger.Version
}

func (ur userRecord) Id() ledger.UserId {
	return ur.id
}

func (ur userRecord) Email() *mail.Address {
	var (
		email *mail.Address
		err   error
	)
	if email, err = mail.ParseAddress(ur.email); err != nil {
		log.Fatalf("Expected valid email to be saved in the database. Reason: %s", err)
	}
	return email
}

func (ur userRecord) CreatedBy() ledger.UpdatedBy {
	var (
		updatedBy ledger.UpdatedBy
		err       error
	)
	if updatedBy, err = ledger.ParseUpdatedBy(ur.createdBy); err != nil {
		log.Fatalf("Invalid createdBy persisted for record %d: %s", ur.id, ur.createdBy)
	}
	return updatedBy
}

func (ur userRecord) CreatedAtUTC() time.Time {
	return ur.createdAt
}

func (ur userRecord) ModifiedBy() ledger.UpdatedBy {
	if !ur.modifiedBy.Valid {
		return ledger.UpdatedBy{}
	}
	var (
		updatedBy ledger.UpdatedBy
		err       error
	)
	if updatedBy, err = ledger.ParseUpdatedBy(ur.modifiedBy.String); err != nil {
		log.Fatalf("Invalid modifiedBy persisted for record %d: %s", ur.id, ur.ModifiedBy())
	}
	return updatedBy
}

func (ur userRecord) ModifiedAtUTC() time.Time {
	if ur.modifiedAt.Valid {
		return ur.modifiedAt.Time
	}
	return time.Time{}
}

func (ur userRecord) Version() ledger.Version {
	return ur.version
}

type DefaultUserDao struct {
	*RootDao
}

func MustOpenUserDao(db *sql.DB) dao.UserDao {
	return &DefaultUserDao{
		&RootDao{db},
	}
}

func (d DefaultUserDao) Close() error {
	return d.db.Close()
}

func (d *DefaultUserDao) NewUserId() (ledger.UserId, error) {
	var userId ledger.UserId
	err := d.db.QueryRow("SELECT nextval('budget.user_id')").Scan(&userId)
	if err != nil {
		log.Printf("Failed to assign user id. Reason; %s", err)
		return 0, ledger.NewError(ledger.ErrDatabaseState, "Failed to assign user id", err)
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
		log.Printf("Failed to save user %v. Reason: %s", u, err)
		if message, ok := d.isDuplicateKeyError(err); ok {
			return ledger.NewError(ledger.ErrUserEmailDuplicated, message, err)
		}
		return ledger.NewError(ledger.ErrDatabaseState, "Failed to save user", err)
	}
	return nil
}

func (d *DefaultUserDao) GetUserById(queryId ledger.UserId) (ledger.User, error) {
	var ur userRecord
	err := d.db.QueryRow("SELECT id, email, created_by, created_at, last_modified_by, last_modified_at, version FROM budget.user WHERE id = $1", queryId).
		Scan(&ur.id, &ur.email, &ur.createdBy, &ur.createdAt, &ur.modifiedBy, &ur.modifiedAt, &ur.version)
	if err != nil {
		if err == sql.ErrNoRows {
			return ledger.User{}, ledger.NewError(ledger.ErrUserNotFound, fmt.Sprintf("User with id %d not found", queryId), err)
		}
		return ledger.User{}, ledger.NewError(ledger.ErrDatabaseState, fmt.Sprintf("User with id %d not found", queryId), err)
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
