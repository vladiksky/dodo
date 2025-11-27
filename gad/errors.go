package errors

import "errors"

// Кастомные ошибки
var (
	ErrInsufficientFunds   = errors.New("недостаточно средств на счете")
	ErrInvalidAmount       = errors.New("некорректная сумма (отрицательная или нулевая)")
	ErrAccountNotFound     = errors.New("счет не найден")
	ErrSameAccountTransfer = errors.New("попытка перевода на тот же счёт")
)
