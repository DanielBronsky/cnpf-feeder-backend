package errors

import (
	"fmt"
	"strings"
)

// TranslateError translates common MongoDB and system errors to Russian
func TranslateError(err error) error {
	if err == nil {
		return nil
	}

	errStr := err.Error()

	// MongoDB errors
	if strings.Contains(errStr, "database name cannot be empty") {
		return fmt.Errorf("имя базы данных не может быть пустым. Проверьте MONGODB_URI в переменных окружения")
	}
	if strings.Contains(errStr, "no such host") || strings.Contains(errStr, "connection refused") {
		return fmt.Errorf("не удалось подключиться к MongoDB. Убедитесь, что MongoDB запущен")
	}
	if strings.Contains(errStr, "server selection timeout") {
		return fmt.Errorf("таймаут подключения к MongoDB. Проверьте, что MongoDB запущен и доступен")
	}
	if strings.Contains(errStr, "authentication failed") {
		return fmt.Errorf("ошибка аутентификации в MongoDB. Проверьте учетные данные")
	}
	if strings.Contains(errStr, "E11000") || strings.Contains(errStr, "duplicate key") {
		return fmt.Errorf("запись с таким значением уже существует")
	}

	// Common system errors
	if strings.Contains(errStr, "context deadline exceeded") || strings.Contains(errStr, "timeout") {
		return fmt.Errorf("превышено время ожидания операции")
	}
	if strings.Contains(errStr, "connection closed") {
		return fmt.Errorf("соединение с базой данных закрыто")
	}
	if strings.Contains(errStr, "not found") {
		return fmt.Errorf("запись не найдена")
	}

	// Return original error if no translation found
	return err
}

// WrapError wraps error with Russian message
func WrapError(msg string, err error) error {
	if err == nil {
		return nil
	}
	translated := TranslateError(err)
	return fmt.Errorf("%s: %w", msg, translated)
}
