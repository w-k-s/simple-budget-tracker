package services

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/w-k-s/simple-budget-tracker/pkg/ledger"
	dao "github.com/w-k-s/simple-budget-tracker/pkg/persistence"
)

type CreateRecordRequest struct {
	Account struct {
		Id uint64 `json:"id"`
	} `json:"account"`
	Note     string `json:"note"`
	Category struct {
		Id   uint64 `json:"id"`
		Name string `json:"name"`
	} `json:"category"`
	Amount struct {
		Currency string `json:"currency"`
		Value    uint64 `json:"value"`
	} `json:"amount"`
	DateUTC  string `json:"date"`
	Type     string `json:"type"`
	Transfer struct {
		Beneficiary struct {
			Id uint64 `json:"id"`
		} `json:"beneficiary"`
	} `json:"transfer"`
}

type RecordResponse struct {
	Id       uint64 `json:"id"`
	Note     string `json:"note"`
	Category struct {
		Id   uint64 `json:"id"`
		Name string `json:"name"`
	} `json:"category"`
	Amount  AmountResponse `json:"amount"`
	DateUTC string         `json:"date"`
	Type    string         `json:"type"`

	// Transfer is only set when record type is transfer
	Transfer struct {
		Beneficiary struct {
			Id uint64 `json:"id"`
		} `json:"beneficiary"`
	} `json:"transfer"`

	// Account is only set when a single record is updated.
	Account AccountBalanceResponse `json:"account"`
}

type AccountBalanceResponse struct {
	Id      uint64         `json:"id"`
	Balance AmountResponse `json:"currentBalance"`
}

type AmountResponse struct {
	Currency string `json:"currency"`
	Value    uint64 `json:"value"`
}

type CalendarMonthRecordsResponse struct {
	Records []RecordResponse `json:"records"`
	Summary struct {
		TotalExpenses AmountResponse `json:"totalExpenses"`
		TotalIncome   AmountResponse `json:"totalIncome"`
	} `json:"summary"`
}

type RecordService interface {
	CreateRecord(ctx context.Context, request CreateRecordRequest) (RecordResponse, error)
	GetRecords(ctx context.Context) (CalendarMonthRecordsResponse, error)
}

type recordService struct {
	recordDao   dao.RecordDao
	accountDao  dao.AccountDao
	categoryDao dao.CategoryDao
}

func NewRecordService(recordDao dao.RecordDao, accountDao dao.AccountDao, categoryDao dao.CategoryDao) (RecordService, error) {
	if recordDao == nil {
		return nil, fmt.Errorf("can not create record service. recordDao is nil")
	}
	if accountDao == nil {
		return nil, fmt.Errorf("can not create record service. accountDao is nil")
	}
	if categoryDao == nil {
		return nil, fmt.Errorf("can not create record service. categoryDao is nil")
	}

	return &recordService{
		recordDao:   recordDao,
		accountDao:  accountDao,
		categoryDao: categoryDao,
	}, nil
}

func (svc recordService) CreateRecord(ctx context.Context, request CreateRecordRequest) (RecordResponse, error) {
	var (
		userId ledger.UserId
		tx     *sql.Tx
		err    error
	)

	if userId, err = RequireUserId(ctx); err != nil {
		return RecordResponse{}, err.(ledger.Error)
	}

	if tx, err = svc.recordDao.BeginTx(); err != nil {
		return RecordResponse{}, err.(ledger.Error)
	}

	defer dao.DeferRollback(tx, fmt.Sprintf("CreateRecord: %d", userId))

	var (
		recordId ledger.RecordId
		category ledger.Category
		account  ledger.Account
		amount   ledger.Money
		date     time.Time
		record   ledger.Record
	)

	if recordId, err = svc.recordDao.NewRecordId(tx); err != nil {
		return RecordResponse{}, err.(ledger.Error)
	}

	if category, err = svc.categoryDao.GetCategoryById(ctx, ledger.CategoryId(request.Category.Id), userId, tx); err != nil {
		return RecordResponse{}, err.(ledger.Error)
	}

	if amount, err = ledger.NewMoney(request.Amount.Currency, int64(request.Amount.Value)); err != nil {
		return RecordResponse{}, err.(ledger.Error)
	}

	transferReference := ledger.NoTransferReference
	if ledger.RecordType(request.Type) == ledger.Transfer {
		if amount, err = amount.Negate(); err != nil {
			return RecordResponse{}, err.(ledger.Error)
		}
		transferReference = ledger.MakeTransferReference()
	}

	if date, err = time.Parse(time.RFC3339, request.DateUTC); err != nil {
		return RecordResponse{}, ledger.NewError(ledger.ErrRecordValidation, fmt.Sprintf("Date '%s' does not match format '%s'", request.DateUTC, time.RFC3339), err)
	}

	if record, err = ledger.NewRecord(
		recordId,
		request.Note,
		category,
		amount,
		date.In(time.UTC),
		ledger.RecordType(request.Type),
		account.Id(),
		ledger.AccountId(request.Transfer.Beneficiary.Id),
		transferReference,
		ledger.MustMakeUpdatedByUserId(userId),
	); err != nil {
		return RecordResponse{}, err.(ledger.Error)
	}

	// Save Record(s)
	if err = svc.recordDao.SaveTx(ctx, ledger.AccountId(request.Account.Id), record, tx); err != nil {
		return RecordResponse{}, err.(ledger.Error)
	}

	// In case of transfer, credit receiving account
	if record.Type() == ledger.Transfer {
		var (
			transferId ledger.RecordId
			transfer   ledger.Record
			credit     ledger.Money
		)

		if credit, err = amount.Abs(); err != nil {
			return RecordResponse{}, err.(ledger.Error)
		}

		if transferId, err = svc.recordDao.NewRecordId(tx); err != nil {
			return RecordResponse{}, err.(ledger.Error)
		}

		if transfer, err = ledger.NewRecord(
			transferId,
			fmt.Sprintf("From %s: %s", account.Name(), record.Note()),
			category,
			credit,
			date.In(time.UTC),
			ledger.RecordType(request.Type),
			account.Id(),
			ledger.AccountId(request.Transfer.Beneficiary.Id),
			transferReference,
			ledger.MustMakeUpdatedByUserId(userId),
		); err != nil {
			return RecordResponse{}, err.(ledger.Error)
		}

		if err = svc.recordDao.SaveTx(ctx, ledger.AccountId(request.Transfer.Beneficiary.Id), transfer, tx); err != nil {
			return RecordResponse{}, err.(ledger.Error)
		}
	}

	// Update last used category
	if err = svc.categoryDao.UpdateCategoryLastUsed(ctx, category.Id(), date.In(time.UTC), tx); err != nil {
		return RecordResponse{}, err.(ledger.Error)
	}

	// Get account balance
	if account, err = svc.accountDao.GetAccountById(ctx, ledger.AccountId(request.Account.Id), tx); err != nil {
		return RecordResponse{}, err.(ledger.Error)
	}

	if err = dao.Commit(tx); err != nil {
		return RecordResponse{}, err.(ledger.Error)
	}

	// Return response
	amountValue, _ := record.Amount().MinorUnits()
	currentBalanceValue, _ := account.CurrentBalance().MinorUnits()

	resp := RecordResponse{}
	resp.Id = uint64(recordId)
	resp.Note = record.Note()
	resp.Category.Id = uint64(category.Id())
	resp.Category.Name = category.Name()
	resp.Amount.Currency = record.Amount().Currency().CurrencyCode()
	resp.Amount.Value = uint64(amountValue)
	resp.DateUTC = record.DateUTCString()
	resp.Type = string(record.Type())
	resp.Account.Id = uint64(account.Id())
	resp.Account.Balance.Currency = account.CurrentBalance().Currency().CurrencyCode()
	resp.Account.Balance.Value = uint64(currentBalanceValue)

	if record.Type() == ledger.Transfer {
		resp.Transfer.Beneficiary.Id = uint64(record.BeneficiaryId())
	}

	return resp, nil
}

func (svc recordService) GetRecords(ctx context.Context) (CalendarMonthRecordsResponse, error) {
	return CalendarMonthRecordsResponse{}, nil
}
