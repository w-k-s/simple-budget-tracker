package ledger

import (
	"fmt"
	"net/mail"
	"time"
)

type UserId uint64
type User struct {
	AuditInfo
	id    UserId
	email *mail.Address
}

type UserRecord interface {
	Id() UserId
	Email() *mail.Address
	CreatedBy() UserId
	CreatedAtUTC() time.Time
	ModifiedBy() UserId
	ModifiedAtUTC() time.Time
	Version() Version
}

func NewUser(id UserId, email *mail.Address) (User, error) {
	var (
		audit AuditInfo
		err   error
	)
	if audit, err = MakeAuditForCreation(id); err != nil {
		return User{}, nil
	}

	return newUser(id, email, audit), nil
}

func NewUserWithEmailString(id UserId, emailString string) (User, error) {
	email, err := mail.ParseAddress(emailString)
	if err != nil {
		return User{}, NewError(ErrUserEmailInvalid, err.Error(), err)
	}
	return NewUser(id, email)
}

func NewUserFromRecord(record UserRecord) (User, error) {
	var (
		auditInfo AuditInfo
		err       error
	)

	if auditInfo, err = MakeAuditForModification(
		record.CreatedBy(),
		record.CreatedAtUTC(),
		record.ModifiedBy(),
		record.ModifiedAtUTC(),
		record.Version(),
	); err != nil {
		return User{}, err
	}

	return newUser(record.CreatedBy(), record.Email(), auditInfo), nil
}

func newUser(id UserId, email *mail.Address, auditInfo AuditInfo) User {
	return User{
		AuditInfo: auditInfo,
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
