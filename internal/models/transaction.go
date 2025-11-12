package models

import "time"

type TransactionType string

const (
	TransactionTypeIncome   TransactionType = "INCOME"
	TransactionTypeExpense  TransactionType = "EXPENSE"
	TransactionTypeTransfer TransactionType = "TRANSFER"
)

type Transaction struct {
	ID          string
	AccountID   string
	UserID      string
	Type        TransactionType
	Amount      int64
	Currency    string
	MCC         *int32
	Description string
	CreatedAt   time.Time
}
