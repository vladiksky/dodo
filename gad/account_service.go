package services

import (
	"bankapp/errors"
	"bankapp/interfaces"
	"bankapp/models"
	"fmt"
	"strings"
	"time"
)

// AccountServiceImpl реализация AccountService
type AccountServiceImpl struct {
	account *models.Account
	storage interfaces.Storage
}

// NewAccountService создает новый сервис для работы со счетом
func NewAccountService(account *models.Account, storage interfaces.Storage) interfaces.AccountService {
	return &AccountServiceImpl{
		account: account,
		storage: storage,
	}
}

// Deposit пополнение счета
func (s *AccountServiceImpl) Deposit(amount float64) error {
	if amount <= 0 {
		return errors.ErrInvalidAmount
	}

	s.account.Balance += amount

	transaction := models.Transaction{
		ID:        fmt.Sprintf("TX%d", time.Now().UnixNano()),
		Type:      models.DepositTransaction,
		Amount:    amount,
		Timestamp: time.Now(),
		Message:   fmt.Sprintf("Пополнение счета на %.2f", amount),
	}

	s.account.Transactions = append(s.account.Transactions, transaction)

	return s.storage.SaveAccount(s.account)
}

// Withdraw снятие средств
func (s *AccountServiceImpl) Withdraw(amount float64) error {
	if amount <= 0 {
		return errors.ErrInvalidAmount
	}

	if s.account.Balance < amount {
		return errors.ErrInsufficientFunds
	}

	s.account.Balance -= amount

	transaction := models.Transaction{
		ID:        fmt.Sprintf("TX%d", time.Now().UnixNano()),
		Type:      models.WithdrawTransaction,
		Amount:    amount,
		Timestamp: time.Now(),
		Message:   fmt.Sprintf("Снятие средств на %.2f", amount),
	}

	s.account.Transactions = append(s.account.Transactions, transaction)

	return s.storage.SaveAccount(s.account)
}

// Transfer перевод другому счету
func (s *AccountServiceImpl) Transfer(to *models.Account, amount float64) error {
	if amount <= 0 {
		return errors.ErrInvalidAmount
	}

	if s.account.Balance < amount {
		return errors.ErrInsufficientFunds
	}

	if s.account.ID == to.ID {
		return errors.ErrSameAccountTransfer
	}

	// Снимаем средства с текущего счета
	s.account.Balance -= amount

	transaction := models.Transaction{
		ID:        fmt.Sprintf("TX%d", time.Now().UnixNano()),
		Type:      models.TransferTransaction,
		Amount:    amount,
		Timestamp: time.Now(),
		Message:   fmt.Sprintf("Перевод счету %s на %.2f", to.ID, amount),
	}

	s.account.Transactions = append(s.account.Transactions, transaction)

	// Зачисляем средства на целевой счет
	to.Balance += amount

	toTransaction := models.Transaction{
		ID:        fmt.Sprintf("TX%d", time.Now().UnixNano()),
		Type:      models.TransferTransaction,
		Amount:    amount,
		Timestamp: time.Now(),
		Message:   fmt.Sprintf("Перевод от счета %s на %.2f", s.account.ID, amount),
	}

	to.Transactions = append(to.Transactions, toTransaction)

	// Сохраняем оба счета
	if err := s.storage.SaveAccount(s.account); err != nil {
		return err
	}

	return s.storage.SaveAccount(to)
}

// GetBalance получение баланса
func (s *AccountServiceImpl) GetBalance() float64 {
	return s.account.Balance
}

// GetStatement получение выписки
func (s *AccountServiceImpl) GetStatement() string {
	if len(s.account.Transactions) == 0 {
		return "История транзакций пуста"
	}

	var sb strings.Builder
	sb.WriteString("Выписка по счету:\n")
	sb.WriteString("========================================\n")
	sb.WriteString(fmt.Sprintf("Владелец: %s\n", s.account.OwnerName))
	sb.WriteString(fmt.Sprintf("ID счета: %s\n", s.account.ID))
	sb.WriteString("========================================\n")

	for _, tx := range s.account.Transactions {
		sb.WriteString(fmt.Sprintf("%s | %s | %.2f | %s\n",
			tx.Timestamp.Format("2006-01-02 15:04:05"),
			tx.Type,
			tx.Amount,
			tx.Message))
	}

	sb.WriteString("========================================\n")
	sb.WriteString(fmt.Sprintf("Текущий баланс: %.2f\n", s.account.Balance))

	return sb.String()
}
