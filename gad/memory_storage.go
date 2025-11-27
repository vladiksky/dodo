package storage

import (
	"bankapp/errors"
	"bankapp/interfaces"
	"bankapp/models"
)

// MemoryStorage реализация хранилища в памяти
type MemoryStorage struct {
	accounts map[string]*models.Account
}

// NewMemoryStorage создает новое хранилище в памяти
func NewMemoryStorage() interfaces.Storage {
	return &MemoryStorage{
		accounts: make(map[string]*models.Account),
	}
}

// SaveAccount сохраняет счет
func (s *MemoryStorage) SaveAccount(account *models.Account) error {
	s.accounts[account.ID] = account
	return nil
}

// LoadAccount загружает счет по ID
func (s *MemoryStorage) LoadAccount(accountID string) (*models.Account, error) {
	account, exists := s.accounts[accountID]
	if !exists {
		return nil, errors.ErrAccountNotFound
	}

	return account, nil
}

// GetAllAccounts возвращает все счета
func (s *MemoryStorage) GetAllAccounts() ([]*models.Account, error) {
	accounts := make([]*models.Account, 0, len(s.accounts))
	for _, account := range s.accounts {
		accounts = append(accounts, account)
	}

	return accounts, nil
}
