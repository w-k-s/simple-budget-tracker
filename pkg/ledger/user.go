package ledger

import (
	"fmt"
	"net/mail"
	"time"

	"github.com/w-k-s/simple-budget-tracker/pkg"
)

type UserId uint64
type User struct {
	auditInfo
	id    UserId
	email *mail.Address
}

type UserRecord interface {
	Id() UserId
	Email() *mail.Address
	CreatedBy() UpdatedBy
	CreatedAtUTC() time.Time
	ModifiedBy() UpdatedBy
	ModifiedAtUTC() time.Time
	Version() Version
}

func NewUser(id UserId, email *mail.Address) (User, error) {
	var (
		updatedBy UpdatedBy
		audit     auditInfo
		err       error
	)
	if updatedBy, err = MakeUpdatedByUserId(id); err != nil {
		return User{}, nil
	}
	if audit, err = makeAuditForCreation(updatedBy); err != nil {
		return User{}, nil
	}

	return newUser(id, email, audit), nil
}

func NewUserWithEmailString(id UserId, emailString string) (User, error) {
	email, err := mail.ParseAddress(emailString)
	if err != nil {
		return User{}, pkg.ValidationErrorWithFields(pkg.ErrUserEmailInvalid, "", err, nil)
	}
	return NewUser(id, email)
}

func NewUserFromRecord(record UserRecord) (User, error) {
	var (
		auditInfo auditInfo
		err       error
	)

	if auditInfo, err = makeAuditForModification(
		record.CreatedBy(),
		record.CreatedAtUTC(),
		record.ModifiedBy(),
		record.ModifiedAtUTC(),
		record.Version(),
	); err != nil {
		return User{}, err
	}

	return newUser(record.Id(), record.Email(), auditInfo), nil
}

func newUser(id UserId, email *mail.Address, auditInfo auditInfo) User {
	return User{
		auditInfo: auditInfo,
		id:        id,
		email:     email,
	}
}

func (u User) Id() UserId {
	return u.id
}

func (u User) Email() *mail.Address {
	return u.email
}

func (u User) String() string {
	return fmt.Sprintf("User{id: %d, email: %s}", u.id, u.email.Address)
}
