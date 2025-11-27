package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Кастомные ошибки
var (
	ErrInsufficientFunds   = errors.New("недостаточно средств на счете")
	ErrInvalidAmount       = errors.New("некорректная сумма (отрицательная или нулевая)")
	ErrAccountNotFound     = errors.New("счет не найден")
	ErrSameAccountTransfer = errors.New("попытка перевода на тот же счёт")
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

// AccountService - основной интерфейс для работы со счетом
type AccountService interface {
	Deposit(amount float64) error
	Withdraw(amount float64) error
	Transfer(to *Account, amount float64) error
	GetBalance() float64
	GetStatement() string
}

// Storage - интерфейс для работы с хранилищем данных
type Storage interface {
	SaveAccount(account *Account) error
	LoadAccount(accountID string) (*Account, error)
	GetAllAccounts() ([]*Account, error)
}

// AccountServiceImpl реализация AccountService
type AccountServiceImpl struct {
	account *Account
	storage Storage
}

// NewAccountService создает новый сервис для работы со счетом
func NewAccountService(account *Account, storage Storage) AccountService {
	return &AccountServiceImpl{
		account: account,
		storage: storage,
	}
}

// Deposit пополнение счета
func (s *AccountServiceImpl) Deposit(amount float64) error {
	if amount <= 0 {
		return ErrInvalidAmount
	}

	s.account.Balance += amount

	transaction := Transaction{
		ID:        fmt.Sprintf("TX%d", time.Now().UnixNano()),
		Type:      DepositTransaction,
		Amount:    amount,
		Timestamp: time.Now(),
		Message:   fmt.Sprintf("Пополнение счета на %.2f", amount),
	}

	s.account.Transactions = append(s.account.Transactions, transaction)

	// Сохраняем изменения в хранилище
	return s.storage.SaveAccount(s.account)
}

// Withdraw снятие средств
func (s *AccountServiceImpl) Withdraw(amount float64) error {
	if amount <= 0 {
		return ErrInvalidAmount
	}

	if s.account.Balance < amount {
		return ErrInsufficientFunds
	}

	s.account.Balance -= amount

	transaction := Transaction{
		ID:        fmt.Sprintf("TX%d", time.Now().UnixNano()),
		Type:      WithdrawTransaction,
		Amount:    amount,
		Timestamp: time.Now(),
		Message:   fmt.Sprintf("Снятие средств на %.2f", amount),
	}

	s.account.Transactions = append(s.account.Transactions, transaction)

	return s.storage.SaveAccount(s.account)
}

// Transfer перевод другому счету
func (s *AccountServiceImpl) Transfer(to *Account, amount float64) error {
	if amount <= 0 {
		return ErrInvalidAmount
	}

	if s.account.Balance < amount {
		return ErrInsufficientFunds
	}

	if s.account.ID == to.ID {
		return ErrSameAccountTransfer
	}

	// Снимаем средства с текущего счета
	s.account.Balance -= amount

	transaction := Transaction{
		ID:        fmt.Sprintf("TX%d", time.Now().UnixNano()),
		Type:      TransferTransaction,
		Amount:    amount,
		Timestamp: time.Now(),
		Message:   fmt.Sprintf("Перевод счету %s на %.2f", to.ID, amount),
	}

	s.account.Transactions = append(s.account.Transactions, transaction)

	// Зачисляем средства на целевой счет
	to.Balance += amount

	toTransaction := Transaction{
		ID:        fmt.Sprintf("TX%d", time.Now().UnixNano()),
		Type:      TransferTransaction,
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

// MemoryStorage реализация хранилища в памяти
type MemoryStorage struct {
	accounts map[string]*Account
}

// NewMemoryStorage создает новое хранилище в памяти
func NewMemoryStorage() Storage {
	return &MemoryStorage{
		accounts: make(map[string]*Account),
	}
}

// SaveAccount сохраняет счет
func (s *MemoryStorage) SaveAccount(account *Account) error {
	s.accounts[account.ID] = account
	return nil
}

// LoadAccount загружает счет по ID
func (s *MemoryStorage) LoadAccount(accountID string) (*Account, error) {
	account, exists := s.accounts[accountID]
	if !exists {
		return nil, ErrAccountNotFound
	}

	return account, nil
}

// GetAllAccounts возвращает все счета
func (s *MemoryStorage) GetAllAccounts() ([]*Account, error) {
	accounts := make([]*Account, 0, len(s.accounts))
	for _, account := range s.accounts {
		accounts = append(accounts, account)
	}

	return accounts, nil
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

// BankApp структура банковского приложения
type BankApp struct {
	storage        Storage
	accounts       map[string]AccountService
	currentAccount AccountService
	scanner        *bufio.Scanner
}

// NewBankApp создает новое банковское приложение
func NewBankApp() *BankApp {
	storage := NewMemoryStorage()
	return &BankApp{
		storage:  storage,
		accounts: make(map[string]AccountService),
		scanner:  bufio.NewScanner(os.Stdin),
	}
}

// Run запускает приложение
func (app *BankApp) Run() {
	fmt.Println("=== Банковское приложение ===")

	for {
		if app.currentAccount == nil {
			app.showMainMenu()
		} else {
			app.showAccountMenu()
		}
	}
}

// showMainMenu показывает главное меню
func (app *BankApp) showMainMenu() {
	fmt.Println("\n--- Главное меню ---")
	fmt.Println("1. Создать счет")
	fmt.Println("2. Выбрать счет")
	fmt.Println("3. Показать все счета")
	fmt.Println("4. Выйти")
	fmt.Print("Выберите опцию: ")

	app.scanner.Scan()
	choice := app.scanner.Text()

	switch choice {
	case "1":
		app.createAccount()
	case "2":
		app.selectAccount()
	case "3":
		app.showAllAccounts()
	case "4":
		fmt.Println("До свидания!")
		os.Exit(0)
	default:
		fmt.Println("Неверный выбор. Попробуйте снова.")
	}
}

// showAccountMenu показывает меню счета
func (app *BankApp) showAccountMenu() {
	fmt.Println("\n--- Меню счета ---")
	fmt.Println("1. Пополнить счет")
	fmt.Println("2. Снять средства")
	fmt.Println("3. Перевести другому счету")
	fmt.Println("4. Просмотреть баланс")
	fmt.Println("5. Получить выписку")
	fmt.Println("6. Вернуться в главное меню")
	fmt.Print("Выберите опцию: ")

	app.scanner.Scan()
	choice := app.scanner.Text()

	switch choice {
	case "1":
		app.deposit()
	case "2":
		app.withdraw()
	case "3":
		app.transfer()
	case "4":
		app.showBalance()
	case "5":
		app.showStatement()
	case "6":
		app.currentAccount = nil
		fmt.Println("Возврат в главное меню...")
	default:
		fmt.Println("Неверный выбор. Попробуйте снова.")
	}
}

// createAccount создает новый счет
func (app *BankApp) createAccount() {
	fmt.Print("Введите имя владельца счета: ")
	app.scanner.Scan()
	ownerName := strings.TrimSpace(app.scanner.Text())

	if ownerName == "" {
		fmt.Println("Имя владельца не может быть пустым")
		return
	}

	account := NewAccount(ownerName)
	accountService := NewAccountService(account, app.storage)

	// Сохраняем счет
	if err := app.storage.SaveAccount(account); err != nil {
		fmt.Printf("Ошибка при создании счета: %v\n", err)
		return
	}

	app.accounts[account.ID] = accountService

	fmt.Printf("Счет успешно создан!\n")
	fmt.Printf("ID счета: %s\n", account.ID)
	fmt.Printf("Владелец: %s\n", account.OwnerName)
}

// selectAccount выбирает счет для работы
func (app *BankApp) selectAccount() {
	fmt.Print("Введите ID счета: ")
	app.scanner.Scan()
	accountID := strings.TrimSpace(app.scanner.Text())

	accountService, exists := app.accounts[accountID]
	if !exists {
		// Попробуем загрузить из хранилища
		account, err := app.storage.LoadAccount(accountID)
		if err != nil {
			fmt.Printf("Ошибка: %v\n", ErrAccountNotFound)
			return
		}
		accountService = NewAccountService(account, app.storage)
		app.accounts[accountID] = accountService
	}

	app.currentAccount = accountService
	fmt.Printf("Счет %s выбран для работы\n", accountID)
}

// showAllAccounts показывает все счета
func (app *BankApp) showAllAccounts() {
	accounts, err := app.storage.GetAllAccounts()
	if err != nil {
		fmt.Printf("Ошибка при получении счетов: %v\n", err)
		return
	}

	if len(accounts) == 0 {
		fmt.Println("Счета не найдены")
		return
	}

	fmt.Println("\n--- Все счета ---")
	for _, account := range accounts {
		fmt.Printf("ID: %s | Владелец: %s | Баланс: %.2f\n",
			account.ID, account.OwnerName, account.Balance)
	}
}

// deposit пополняет счет
func (app *BankApp) deposit() {
	amount, err := app.readAmount("Введите сумму для пополнения: ")
	if err != nil {
		return
	}

	if err := app.currentAccount.Deposit(amount); err != nil {
		fmt.Printf("Ошибка при пополнении: %v\n", err)
		return
	}

	fmt.Printf("Счет успешно пополнен на %.2f\n", amount)
}

// withdraw снимает средства
func (app *BankApp) withdraw() {
	amount, err := app.readAmount("Введите сумму для снятия: ")
	if err != nil {
		return
	}

	if err := app.currentAccount.Withdraw(amount); err != nil {
		fmt.Printf("Ошибка при снятии: %v\n", err)
		return
	}

	fmt.Printf("Со счета успешно снято %.2f\n", amount)
}

// transfer переводит средства другому счету
func (app *BankApp) transfer() {
	amount, err := app.readAmount("Введите сумму для перевода: ")
	if err != nil {
		return
	}

	fmt.Print("Введите ID целевого счета: ")
	app.scanner.Scan()
	toAccountID := strings.TrimSpace(app.scanner.Text())

	// Загружаем целевой счет
	toAccount, err := app.storage.LoadAccount(toAccountID)
	if err != nil {
		fmt.Printf("Ошибка: %v\n", err)
		return
	}

	if err := app.currentAccount.Transfer(toAccount, amount); err != nil {
		fmt.Printf("Ошибка при переводе: %v\n", err)
		return
	}

	fmt.Printf("Успешно переведено %.2f на счет %s\n", amount, toAccountID)
}

// showBalance показывает баланс
func (app *BankApp) showBalance() {
	balance := app.currentAccount.GetBalance()
	fmt.Printf("Текущий баланс: %.2f\n", balance)
}

// showStatement показывает выписку
func (app *BankApp) showStatement() {
	statement := app.currentAccount.GetStatement()
	fmt.Println(statement)
}

// readAmount читает сумму из ввода
func (app *BankApp) readAmount(prompt string) (float64, error) {
	fmt.Print(prompt)
	app.scanner.Scan()
	input := strings.TrimSpace(app.scanner.Text())

	amount, err := strconv.ParseFloat(input, 64)
	if err != nil || amount <= 0 {
		fmt.Printf("Ошибка: %v\n", ErrInvalidAmount)
		return 0, ErrInvalidAmount
	}

	return amount, nil
}

func main() {
	app := NewBankApp()
	app.Run()
}
