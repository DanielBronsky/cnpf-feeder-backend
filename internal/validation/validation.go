package validation

import (
	"fmt"
	"regexp"
	"strings"
)

var usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)

// ValidateEmail validates email format
func ValidateEmail(email string) error {
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" {
		return fmt.Errorf("Email обязателен")
	}
	// Simple email validation
	if !strings.Contains(email, "@") || !strings.Contains(email, ".") {
		return fmt.Errorf("Неверный формат email")
	}
	return nil
}

// ValidateUsername validates username format
func ValidateUsername(username string) error {
	username = strings.TrimSpace(strings.ToLower(username))
	if len(username) < 3 || len(username) > 24 {
		return fmt.Errorf("Имя пользователя должно быть от 3 до 24 символов")
	}
	if !usernameRegex.MatchString(username) {
		return fmt.Errorf("Имя пользователя должно содержать только буквы, цифры и подчеркивания")
	}
	return nil
}

// ValidatePassword validates password
func ValidatePassword(password string) error {
	if len(password) < 8 || len(password) > 72 {
		return fmt.Errorf("Пароль должен быть от 8 до 72 символов")
	}
	return nil
}

// ValidateRegisterInput validates registration input
func ValidateRegisterInput(email, username, password, passwordConfirm string) error {
	if err := ValidateEmail(email); err != nil {
		return err
	}
	if err := ValidateUsername(username); err != nil {
		return err
	}
	if err := ValidatePassword(password); err != nil {
		return err
	}
	if password != passwordConfirm {
		return fmt.Errorf("Пароли не совпадают")
	}
	return nil
}

// ValidateLoginInput validates login input
func ValidateLoginInput(login, password string) error {
	login = strings.TrimSpace(strings.ToLower(login))
	if len(login) < 3 || len(login) > 254 {
		return fmt.Errorf("Логин должен быть от 3 до 254 символов")
	}
	if len(password) < 1 || len(password) > 72 {
		return fmt.Errorf("Пароль обязателен")
	}
	return nil
}
