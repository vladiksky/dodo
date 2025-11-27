package models

import (
	"fmt"
	"time"
)

// TransactionType тип транзакции
type TransactionType string

const (
	DepositTransaction  TransactionType = "DEPOSIT"
	WithdrawTransaction TransactionType = "WITHDRAW"
	TransferTransaction TransactionType = "TRANSFER"
)

// Transaction структура транзакции
type Transaction struct {
	ID        string
	Type      TransactionType
	Amount    float64
	Timestamp time.Time
	Message   string
}

// Account структура счета
type Account struct {
	ID           string
	OwnerName    string
	Balance      float64
	Transactions []Transaction
	CreatedAt    time.Time
}

// NewAccount создает новый счет
func NewAccount(ownerName string) *Account {
	return &Account{
		ID:        generateID(),
		OwnerName: ownerName,
		Balance:   0,
		CreatedAt: time.Now(),
	}
}

// generateID генерирует уникальный ID для счета
func generateID() string {
	return fmt.Sprintf("ACC%d", time.Now().UnixNano())
}
