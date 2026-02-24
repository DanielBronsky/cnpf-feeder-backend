package mail

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

// SendResetEmail sends password reset email via Resend API.
// If RESEND_API_KEY is not set, logs the link to stdout (for dev).
func SendResetEmail(to, resetLink string) error {
	apiKey := os.Getenv("RESEND_API_KEY")
	if apiKey == "" {
		log.Printf("[mail] RESEND_API_KEY not set - would send reset link to %s: %s", to, resetLink)
		return nil
	}

	from := os.Getenv("MAIL_FROM")
	if from == "" {
		from = "CNPF Feeder <onboarding@resend.dev>"
	}

	body := map[string]interface{}{
		"from":    from,
		"to":      []string{to},
		"subject": "Сброс пароля — CNPF Feeder",
		"html": fmt.Sprintf(`
			<p>Здравствуйте!</p>
			<p>Вы запросили сброс пароля. Перейдите по ссылке для создания нового пароля:</p>
			<p><a href="%s">%s</a></p>
			<p>Ссылка действительна 1 час.</p>
			<p>Если вы не запрашивали сброс — просто проигнорируйте это письмо.</p>
		`, resetLink, resetLink),
	}

	data, err := json.Marshal(body)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", "https://api.resend.com/emails", bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("Resend API error: status %d", resp.StatusCode)
	}
	return nil
}
