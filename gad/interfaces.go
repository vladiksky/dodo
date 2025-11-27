package interfaces

import "bankapp/models"

// AccountService - основной интерфейс для работы со счетом
type AccountService interface {
	Deposit(amount float64) error
	Withdraw(amount float64) error
	Transfer(to *models.Account, amount float64) error
	GetBalance() float64
	GetStatement() string
}

// Storage - интерфейс для работы с хранилищем данных
type Storage interface {
	SaveAccount(account *models.Account) error
	LoadAccount(accountID string) (*models.Account, error)
	GetAllAccounts() ([]*models.Account, error)
}
