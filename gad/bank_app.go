package app

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"bankapp/errors"
	"bankapp/interfaces"
	"bankapp/models"
	"bankapp/services"
	"bankapp/storage"
)

// BankApp структура банковского приложения
type BankApp struct {
	storage        interfaces.Storage
	accounts       map[string]interfaces.AccountService
	currentAccount interfaces.AccountService
	scanner        *bufio.Scanner
}

// NewBankApp создает новое банковское приложение
func NewBankApp() *BankApp {
	storage := storage.NewMemoryStorage()
	return &BankApp{
		storage:  storage,
		accounts: make(map[string]interfaces.AccountService),
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

	account := models.NewAccount(ownerName)
	accountService := services.NewAccountService(account, app.storage)

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
			fmt.Printf("Ошибка: %v\n", errors.ErrAccountNotFound)
			return
		}
		accountService = services.NewAccountService(account, app.storage)
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
		fmt.Printf("Ошибка: %v\n", errors.ErrInvalidAmount)
		return 0, errors.ErrInvalidAmount
	}

	return amount, nil
}
