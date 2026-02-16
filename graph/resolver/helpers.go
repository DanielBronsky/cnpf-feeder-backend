package resolver

import (
	"context"
	"os"
	"strings"
	"unicode"

	"github.com/gin-gonic/gin"

	"github.com/cnpf/feeder-backend/internal/auth"
)

// GetGinContext extracts Gin context from GraphQL context
func GetGinContext(ctx context.Context) *gin.Context {
	if gc, ok := ctx.Value("ginContext").(*gin.Context); ok {
		return gc
	}
	return nil
}

// getCurrentUserFromContext extracts current user from context
func getCurrentUserFromContext(ctx context.Context) (*auth.CurrentUser, error) {
	ginCtx := GetGinContext(ctx)
	if ginCtx == nil {
		return nil, nil
	}
	return auth.GetCurrentUser(ginCtx)
}

// setAuthCookie sets authentication cookie
func setAuthCookie(ctx context.Context, token string) {
	if ginCtx := GetGinContext(ctx); ginCtx != nil {
		// Build cookie string with SameSite attribute
		isSecure := os.Getenv("GIN_MODE") == "release"
		cookieValue := auth.AuthCookieName + "=" + token
		cookieValue += "; Path=/"
		cookieValue += "; Max-Age=2592000" // 30 days
		cookieValue += "; HttpOnly"
		if isSecure {
			cookieValue += "; Secure"
		}
		cookieValue += "; SameSite=Lax" // Для поддержки cross-origin запросов
		
		ginCtx.Header("Set-Cookie", cookieValue)
	}
}

// clearAuthCookie clears authentication cookie
func clearAuthCookie(ctx context.Context) {
	if ginCtx := GetGinContext(ctx); ginCtx != nil {
		// Build cookie string with SameSite attribute to match setAuthCookie
		isSecure := os.Getenv("GIN_MODE") == "release"
		cookieValue := auth.AuthCookieName + "="
		cookieValue += "; Path=/"
		cookieValue += "; Max-Age=0" // Expire immediately
		cookieValue += "; HttpOnly"
		if isSecure {
			cookieValue += "; Secure"
		}
		cookieValue += "; SameSite=Lax" // Match setAuthCookie for consistency
		
		ginCtx.Header("Set-Cookie", cookieValue)
	}
}

// isAllowed checks if user is allowed to perform action
func isAllowed(user *auth.CurrentUser, authorID string) bool {
	if user == nil {
		return false
	}
	return user.IsAdmin || user.ID == authorID
}

func normalizeQueryForIntent(q string) string {
	q = strings.ToLower(strings.TrimSpace(q))
	if q == "" {
		return ""
	}

	// Replace punctuation with spaces and collapse whitespace.
	var b strings.Builder
	b.Grow(len(q))
	for _, r := range q {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || unicode.IsSpace(r) {
			b.WriteRune(r)
			continue
		}
		// punctuation / symbols -> space
		b.WriteRune(' ')
	}

	space := strings.Fields(b.String())
	return strings.Join(space, " ")
}

// getSmallTalkIntent detects small-talk intent and returns (intent, ok).
// intent values: greet, status, whoami, capabilities, howto, help, thanks, bye, smalltalk
func getSmallTalkIntent(q string) (string, bool) {
	q = normalizeQueryForIntent(q)
	if q == "" {
		return "smalltalk", true
	}

	// Exact matches
	switch q {
	case "привет", "прив", "здравствуй", "здравствуйте",
		"добрый день", "добрый вечер", "доброе утро",
		"hi", "hello", "hey":
		return "greet", true
	case "как дела", "как ты", "как поживаешь", "что нового":
		return "status", true
	case "кто ты", "ты кто", "кто вы", "ты бот", "это бот":
		return "whoami", true
	case "что ты умеешь", "что ты можешь", "что умеешь", "что можешь", "возможности":
		return "capabilities", true
	case "как пользоваться", "как пользоваться ботом", "как пользоваться чатом", "как искать", "как искать тут", "как найти":
		return "howto", true
	case "помоги", "помощь", "help", "инструкция", "команды":
		return "help", true
	case "спасибо", "спс", "thanks", "thank you":
		return "thanks", true
	case "пока", "до свидания", "до встречи", "bye":
		return "bye", true
	}

	// Contains-based heuristics
	if strings.Contains(q, "кто ты") || strings.Contains(q, "ты кто") || strings.Contains(q, "кто вы") {
		return "whoami", true
	}
	if strings.Contains(q, "что ты уме") || strings.Contains(q, "что ты мож") || strings.Contains(q, "возможност") {
		return "capabilities", true
	}
	if strings.Contains(q, "как польз") || strings.Contains(q, "как искать") || strings.Contains(q, "как найти") {
		return "howto", true
	}
	if strings.Contains(q, "помог") || strings.Contains(q, "help") || strings.Contains(q, "инструк") {
		return "help", true
	}
	if strings.Contains(q, "спасибо") || strings.Contains(q, "спс") {
		return "thanks", true
	}
	if strings.Contains(q, "пока") || strings.Contains(q, "до свид") {
		return "bye", true
	}

	// Very short messages: treat as generic small talk
	if len(strings.Fields(q)) <= 3 && len(q) <= 16 {
		return "smalltalk", true
	}

	return "", false
}

func isSmallTalkQuery(q string) bool {
	_, ok := getSmallTalkIntent(q)
	return ok
}
