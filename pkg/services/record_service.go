package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ayush6624/go-chatgpt"
	"github.com/w-k-s/simple-budget-tracker/pkg/ledger"
	dao "github.com/w-k-s/simple-budget-tracker/pkg/persistence"
)

type CreateRecordRequest struct {
	Note     string `json:"note"`
	Category struct {
		Id   uint64 `json:"id"`
		Name string `json:"name"`
	} `json:"category"`
	Amount struct {
		Currency string `json:"currency"`
		Value    int64  `json:"value"`
	} `json:"amount"`
	DateUTC  string `json:"date"`
	Type     string `json:"type"`
	Transfer struct {
		Beneficiary struct {
			Id uint64 `json:"id"`
		} `json:"beneficiary"`
	} `json:"transfer,omitempty"`
}

type CreateRecordPrompt struct{
	Prompt string `json:"prompt"`
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
	Transfer *TransferResponse `json:"transfer,omitempty"`

	// Account is only set when a single record is updated.
	Account *AccountBalanceResponse `json:"account,omitempty"`
}

func makeRecordResponse(record ledger.Record, account ledger.Account) (RecordResponse, error) {
	amountValue, _ := record.Amount().MinorUnits()

	resp := RecordResponse{}
	resp.Id = uint64(record.Id())
	resp.Note = record.Note()
	resp.Category.Id = uint64(record.Category().Id())
	resp.Category.Name = record.Category().Name()
	resp.Amount.Currency = record.Amount().Currency().CurrencyCode()
	resp.Amount.Value = amountValue
	resp.DateUTC = record.DateUTCString()
	resp.Type = string(record.Type())

	emptyAccount := ledger.Account{}
	if account != emptyAccount {
		currentBalanceValue, _ := account.CurrentBalance().MinorUnits()

		resp.Account = new(AccountBalanceResponse)
		resp.Account.Id = uint64(account.Id())
		resp.Account.Balance.Currency = account.CurrentBalance().Currency().CurrencyCode()
		resp.Account.Balance.Value = currentBalanceValue
	}

	if record.Type() == ledger.Transfer {
		resp.Transfer = new(TransferResponse)
		resp.Transfer.Beneficiary.Id = uint64(record.BeneficiaryId())
	}

	return resp, nil
}

type TransferResponse struct {
	Beneficiary struct {
		Id uint64 `json:"id"`
	} `json:"beneficiary"`
}

type AccountBalanceResponse struct {
	Id      uint64         `json:"id"`
	Balance AmountResponse `json:"currentBalance"`
}

type AmountResponse struct {
	Currency string `json:"currency"`
	Value    int64  `json:"value"`
}

type RecordsResponse struct {
	Records []RecordResponse `json:"records"`
	Summary struct {
		TotalExpenses AmountResponse `json:"totalExpenses"`
		TotalIncome   AmountResponse `json:"totalIncome"`
		TotalSavings  AmountResponse `json:"totalSavings"`
	} `json:"summary"`
	SearchParameters struct {
		From time.Time `json:"from"`
		To   time.Time `json:"to"`
	} `json:"search"`
}

func makeRecordsResponse(records ledger.Records) (RecordsResponse, error) {
	if len(records) == 0 {
		return RecordsResponse{}, nil
	}

	moneyToAmountResponse := func(money ledger.Money, err error) (AmountResponse, error) {
		if err != nil {
			return AmountResponse{}, err
		}
		var minorUnits int64
		if minorUnits, err = money.MinorUnits(); err != nil {
			return AmountResponse{}, err
		}
		return AmountResponse{
			Currency: money.Currency().CurrencyCode(),
			Value:    minorUnits,
		}, nil
	}

	var (
		totalExpenses AmountResponse
		totalIncome   AmountResponse
		totalSavings  AmountResponse

		from time.Time
		to   time.Time

		err error
	)

	if totalIncome, err = moneyToAmountResponse(records.TotalIncome()); err != nil {
		return RecordsResponse{}, err.(ledger.Error)
	}
	if totalExpenses, err = moneyToAmountResponse(records.TotalExpenses()); err != nil {
		return RecordsResponse{}, err.(ledger.Error)
	}
	if totalSavings, err = moneyToAmountResponse(records.TotalSavings()); err != nil {
		return RecordsResponse{}, err.(ledger.Error)
	}

	if from, to, err = records.Period(); err != nil {
		return RecordsResponse{}, err.(ledger.Error)
	}

	var recordsResponse RecordsResponse
	recordsResponse.Summary.TotalExpenses = totalExpenses
	recordsResponse.Summary.TotalIncome = totalIncome
	recordsResponse.Summary.TotalSavings = totalSavings
	recordsResponse.SearchParameters.From = from
	recordsResponse.SearchParameters.To = to

	for _, record := range records {
		var recordResponse RecordResponse

		if recordResponse, err = makeRecordResponse(record, ledger.Account{}); err != nil {
			return RecordsResponse{}, err.(ledger.Error)
		}

		recordsResponse.Records = append(recordsResponse.Records, recordResponse)
	}
	return recordsResponse, nil
}

type RecordService interface {
	CreateRecord(ctx context.Context, request CreateRecordRequest) (RecordResponse, error)
	GetRecords(ctx context.Context, accountId ledger.AccountId) (RecordsResponse, error)
	CreateRecordRequestWithChatGPT(ctx context.Context, prompt CreateRecordPrompt) (CreateRecordRequest, error)
}

type recordService struct {
	recordDao   dao.RecordDao
	accountDao  dao.AccountDao
	categoryDao dao.CategoryDao
	gptApiKey   string
}

func NewRecordService(
	recordDao dao.RecordDao, 
	accountDao dao.AccountDao, 
	categoryDao dao.CategoryDao,
	gptApiKey string,
) (RecordService, error) {
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
		gptApiKey: gptApiKey,
	}, nil
}

func (svc recordService) CreateRecord(ctx context.Context, request CreateRecordRequest) (RecordResponse, error) {
	var (
		userId    ledger.UserId
		accountId ledger.AccountId
		tx        *sql.Tx
		err       error
	)

	if userId, err = RequireUserId(ctx); err != nil {
		return RecordResponse{}, err.(ledger.Error)
	}

	if accountId, err = RequireAccountId(ctx); err != nil {
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

	// TODO: Check account belongs to user id
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
	sourceAccountId := ledger.NoSourceAccount
	var beneficiaryAccount ledger.Account
	if ledger.RecordType(request.Type) == ledger.Transfer {

		if amount, err = amount.Negate(); err != nil {
			return RecordResponse{}, err.(ledger.Error)
		}

		if beneficiaryAccount, err = svc.accountDao.GetAccountById(ctx, ledger.AccountId(request.Transfer.Beneficiary.Id), userId, tx); err != nil {
			return RecordResponse{}, err.(ledger.Error)
		}

		transferReference = ledger.MakeTransferReference()
		sourceAccountId = accountId
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
		sourceAccountId,
		beneficiaryAccount.Id(),
		beneficiaryAccount.Type(),
		transferReference,
		ledger.MustMakeUpdatedByUserId(userId),
	); err != nil {
		return RecordResponse{}, err.(ledger.Error)
	}

	// Save Record(s)
	if err = svc.recordDao.SaveTx(ctx, accountId, record, tx); err != nil {
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
			record.Note(),
			category,
			credit,
			date.In(time.UTC),
			ledger.RecordType(request.Type),
			sourceAccountId,
			beneficiaryAccount.Id(),
			beneficiaryAccount.Type(),
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
	if account, err = svc.accountDao.GetAccountById(ctx, accountId, userId, tx); err != nil {
		return RecordResponse{}, err.(ledger.Error)
	}

	if err = dao.Commit(tx); err != nil {
		return RecordResponse{}, err.(ledger.Error)
	}

	return makeRecordResponse(record, account)
}

func (svc recordService) CreateRecordRequestWithChatGPT(ctx context.Context, prompt CreateRecordPrompt) (CreateRecordRequest, error){
	if len(svc.gptApiKey) == 0{
		return CreateRecordRequest{}, fmt.Errorf("ChatGPT API Key not configured")
	}

	var(
		userId ledger.UserId
		err error
	)
	
	userId, err = RequireUserId(ctx)
	if err != nil {
		return CreateRecordRequest{}, err.(ledger.Error)
	}

	tx, err := svc.categoryDao.BeginTx()
	if err != nil{
		return CreateRecordRequest{}, err.(ledger.Error)
	}
	defer dao.DeferRollback(tx, fmt.Sprintf("CreateRecordRequestWithChatGPT: %d", userId))

	// Get User Categories Names
	categories, err := svc.categoryDao.GetCategoriesForUser(ctx, userId, tx)
	if err != nil{
		return CreateRecordRequest{}, err.(ledger.Error)
	}

	// Get User Account Names
	accounts, err := svc.accountDao.GetAccountsByUserId(ctx, userId, tx)
	if err != nil{
		return CreateRecordRequest{}, err.(ledger.Error)
	}
	
	jsonStructure, err := json.Marshal(CreateRecordRequest{})
	if err != nil{
		return CreateRecordRequest{}, fmt.Errorf("Failed to marshal empty create record request")
	}

	// Create Prompt
	gptRequest := fmt.Sprintf(`
		Populate the fields in JSON structure using the provided prompt.
		The prompt describes an income, expense or a transfer of money from one account to another.

		JSON Structure: """%s"""
		Prompt: """%s""".
		
		Details:
		- Note: Required. this is what the money was spent on e.g. McDonalds
		- Category: Required. the category of the transaction. One of: '%s'.
		- Amount: Required. the value of the transaction. For expenses, this must be negative. By default the currency is: '%s'.
		- Date: Required. The date of the transaction formatted as yyyy-MM-dd'T'HH:mm:ssX in UTC timezone. Default: '%s'.
		- Type: Required. One of INCOME, EXPENSE or TRANSFER. Default: EXPENSE.
		- Transfer.Beneficiary.Id: Remove transfer field if type is not TRANSFER. The id of the account the money is being transferred to. Available accounts are: '%s'
		
		Output:
		- Output only the populated JSON.
		- If required fields are blank, respond with a json object with a field error and a value listing all missing details separated by commas.
		`,
		string(jsonStructure),
		prompt.Prompt,
		categories.String(),
		accounts[0].Currency(),
		time.Now().UTC().Format("2006-01-02T15:04:05.999Z"),
		accounts.String(),
	)

	// Send Request to Chat-GPT
	client, err := chatgpt.NewClient(svc.gptApiKey)
	if err != nil {
		return CreateRecordRequest{}, fmt.Errorf("Failed to create GPT Client")
	}
	
	res, err := client.SimpleSend(ctx, gptRequest)
	if err != nil {
		return CreateRecordRequest{}, fmt.Errorf("Failed to create GPT Client: %w", err)
	}
	
	if len(res.Choices) == 0 {
		return CreateRecordRequest{}, fmt.Errorf("No response from ChatGPT")
	}

	populatedRequest := CreateRecordRequest{}
	
	err = json.Unmarshal([]byte(res.Choices[0].Message.Content), &populatedRequest)
	if err != nil {
		return CreateRecordRequest{}, fmt.Errorf("Failed to parse GPT response: %w", err)
	}

	return populatedRequest, nil
}

func (svc recordService) GetRecords(ctx context.Context, accountId ledger.AccountId) (RecordsResponse, error) {

	userId, err := RequireUserId(ctx); 
	if err != nil {
		return RecordsResponse{}, err.(ledger.Error)
	}

	tx, err := svc.recordDao.BeginTx(); 
	if err != nil {
		return RecordsResponse{}, err.(ledger.Error)
	}

	defer dao.DeferRollback(tx, fmt.Sprintf("GetRecords: %d", userId))

	if _, err = svc.accountDao.GetAccountById(ctx, accountId, userId, tx); err != nil {
		return RecordsResponse{}, err.(ledger.Error)
	}

	records, err := svc.recordDao.GetRecordsForLastPeriod(ctx, accountId, tx); 
	if err != nil {
		return RecordsResponse{}, err.(ledger.Error)
	}

	return makeRecordsResponse(records)
}
