package models

import "errors"

var (
	ErrClientError = errors.New("ошибка клиента (4xx)") // Client error (4xx)
	ErrServerError = errors.New("ошибка сервера (5xx)") // Server error (5xx)
	ErrTimeout     = errors.New("таймаут")              // Timeout
)
