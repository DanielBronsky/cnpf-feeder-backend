package gemini

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

// Client handles Gemini API interactions
type Client struct {
	apiKey  string
	baseURL string
	client  *http.Client
}

// NewClient creates a new Gemini client
func NewClient() (*Client, error) {
	apiKey := os.Getenv("GOOGLE_GEMINI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("GOOGLE_GEMINI_API_KEY environment variable is not set")
	}

	return &Client{
		apiKey:  apiKey,
		baseURL: "https://generativelanguage.googleapis.com/v1beta/models/gemini-pro:generateContent",
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

// GenerateContentRequest represents the request payload
type GenerateContentRequest struct {
	Contents []Content `json:"contents"`
}

type Content struct {
	Parts []Part `json:"parts"`
}

type Part struct {
	Text string `json:"text"`
}

// GenerateContentResponse represents the response payload
type GenerateContentResponse struct {
	Candidates []Candidate `json:"candidates"`
}

type Candidate struct {
	Content Content `json:"content"`
}

// GenerateContent sends a prompt to Gemini and returns the generated text
func (c *Client) GenerateContent(prompt string) (string, error) {
	reqBody := GenerateContentRequest{
		Contents: []Content{
			{
				Parts: []Part{
					{Text: prompt},
				},
			},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s?key=%s", c.baseURL, c.apiKey)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var response GenerateContentResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if len(response.Candidates) == 0 || len(response.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no content in response")
	}

	return response.Candidates[0].Content.Parts[0].Text, nil
}

// AnalyzeContent analyzes a report or competition using Gemini
func (c *Client) AnalyzeContent(title, text string, contentType string) (string, error) {
	var prompt string
	if contentType == "report" {
		prompt = fmt.Sprintf(`Ты помощник для сайта о рыбалке. Проанализируй эту статью и дай краткий, но информативный анализ (2-3 предложения). Будь дружелюбным и заинтересованным.

Заголовок: %s

Текст статьи:
%s

Дай анализ статьи:`, title, text)
	} else {
		prompt = fmt.Sprintf(`Ты помощник для сайта о рыбалке. Проанализируй информацию о соревновании и дай краткое описание (2-3 предложения). Будь дружелюбным и информативным.

Название соревнования: %s
%s

Дай краткое описание соревнования:`, title, text)
	}

	return c.GenerateContent(prompt)
}

// GenerateSmallTalkResponse generates a friendly response when user is not searching.
// intent examples: greet, status, whoami, capabilities, howto, help, thanks, bye, smalltalk
func (c *Client) GenerateSmallTalkResponse(userQuery string, intent string) (string, error) {
	if intent == "" {
		intent = "smalltalk"
	}

	prompt := fmt.Sprintf(`Ты дружелюбный помощник на сайте о рыбалке.
Пользователь написал: "%s"
Намерение (intent): %s

Ответь по‑русски, коротко и дружелюбно (1-4 предложения), без фразы "не нашел информацию".

Требования по intent:
- greet/status: поприветствуй/ответь как дела, затем предложи 2-3 примера запросов.
- whoami: объясни кто ты и чем полезен, затем 2 примера запросов.
- capabilities: перечисли 4-6 возможностей (поиск отчетов и соревнований, русский/румынский, транслит, подсказки по переформулировке), затем 2 примера.
- howto/help: дай короткую инструкцию в 3 шага и примеры.
- thanks/bye: коротко и дружелюбно.

Примеры запросов (можно выбирать):
- "Отчет о Днестре"
- "соревнования в Данченах"
- "дамба озера Данчены"
- "Ţipala"
- "Hîrjauca"`, userQuery, intent)

	return c.GenerateContent(prompt)
}

// GenerateNoResultsResponse generates a helpful response when nothing found.
func (c *Client) GenerateNoResultsResponse(userQuery string) (string, error) {
	prompt := fmt.Sprintf(`Ты дружелюбный помощник на сайте о рыбалке.
Пользователь спросил: "%s"

Поиск по базе сейчас не нашел подходящих материалов.
Ответь по‑русски, дружелюбно и полезно (2-4 предложения):
- предложи 2-3 варианта как переформулировать запрос
- предложи примеры ("Отчет о Днестре", "соревнования в Данченах", "Ţipala", "Hîrjauca")
- задай один уточняющий вопрос (что именно ищем: отчет или соревнование, и какое место).`, userQuery)

	return c.GenerateContent(prompt)
}

// ExtractSearchQuery extracts and optimizes search query from user input using AI
// Returns optimized search terms in Romanian and transliterated forms
func (c *Client) ExtractSearchQuery(userQuery string) (string, error) {
	prompt := fmt.Sprintf(`Ты помощник для сайта о рыбалке в Молдове. Пользователь задал вопрос: "%s"

Твоя задача - извлечь ключевые слова для поиска в базе данных:
1. Извлеки ключевые слова (исключи служебные слова: "в", "на", "о", "по", "и", "с", "для", "про" и т.д.)
2. Переведи ключевые слова на румынский язык (если они на русском)
3. Добавь транслитерацию кириллицы в латиницу (например: "Данчены" -> "danceni", "Днестр" -> "dnestr")
4. ВАЖНО: Верни ТОЛЬКО ключевые слова через пробел, БЕЗ объяснений, БЕЗ дополнительного текста

Примеры правильных ответов:
"Соревнования в Данченах" -> competitie danceni etapa
"Отчет о Днестре" -> raport dnestr nistru
"Lacul Danceni" -> lacul danceni

Верни ТОЛЬКО ключевые слова (без кавычек, без объяснений):`, userQuery)

	result, err := c.GenerateContent(prompt)
	if err != nil {
		return "", fmt.Errorf("failed to extract search query: %w", err)
	}

	// Clean up the result - remove extra whitespace, newlines, and any explanations
	result = strings.TrimSpace(result)
	// Remove common prefixes that AI might add
	result = strings.TrimPrefix(result, "Ключевые слова:")
	result = strings.TrimPrefix(result, "Ключевые слова для поиска:")
	result = strings.TrimPrefix(result, "Ответ:")
	result = strings.TrimPrefix(result, "->")
	result = strings.TrimPrefix(result, "→")
	// Remove quotes if present
	result = strings.Trim(result, `"'`)
	result = strings.ReplaceAll(result, "\n", " ")
	// Remove multiple spaces
	space := regexp.MustCompile(`\s+`)
	result = space.ReplaceAllString(result, " ")
	result = strings.TrimSpace(result)

	// If result is empty or too short, fallback to original query processing
	if result == "" || len(result) < 2 {
		// Fallback: try to extract words manually
		stopWords := map[string]bool{
			"что": true, "где": true, "когда": true, "как": true,
			"в": true, "на": true, "о": true, "по": true, "и": true,
			"с": true, "для": true, "про": true, "там": true, "тут": true,
			"это": true, "был": true, "была": true, "было": true, "были": true,
			"у": true, "к": true, "из": true, "от": true, "за": true,
			"а": true, "но": true, "или": true, "да": true, "не": true,
		}
		words := strings.Fields(strings.ToLower(userQuery))
		var keyWords []string
		for _, word := range words {
			if len(word) >= 2 && !stopWords[word] {
				keyWords = append(keyWords, word)
			}
		}
		if len(keyWords) > 0 {
			result = strings.Join(keyWords, " ")
		} else {
			result = userQuery // Last resort: use original query
		}
	}

	return result, nil
}

// GenerateChatResponse generates a chat response based on search results
func (c *Client) GenerateChatResponse(userQuery string, results []SearchResult) (string, error) {
	if len(results) == 0 {
		return "К сожалению, я не нашел информацию по вашему запросу. Попробуйте переформулировать вопрос.", nil
	}

	resultsText := ""
	for i, r := range results {
		if r.Type == "report" {
			textPreview := r.Text
			if len(textPreview) > 200 {
				textPreview = textPreview[:200] + "..."
			}
			resultsText += fmt.Sprintf("%d. 📄 Отчет: \"%s\"\n   Текст: %s\n\n", i+1, r.Title, textPreview)
		} else if r.Type == "competition" {
			locationInfo := ""
			if r.Location != "" {
				locationInfo = fmt.Sprintf("\n   Место: %s", r.Location)
			}
			resultsText += fmt.Sprintf("%d. 🏆 Соревнование: \"%s\"%s\n\n", i+1, r.Title, locationInfo)
		}
	}

	prompt := fmt.Sprintf(`Ты дружелюбный помощник на сайте о рыбалке. Пользователь спросил: "%s"

Я нашел следующие результаты:
%s

ВАЖНО: 
- Если результат имеет тип "competition" (соревнование) - говори "соревнование" или "соревновании", НЕ говори "сюжет" или "отчет"
- Если результат имеет тип "report" (отчет) - говори "отчет" или "отчете", НЕ говори "соревнование"

Ответь пользователю дружелюбно, упомяни найденные материалы с правильными названиями типов (соревнование/отчет) и предложи перейти к ним для просмотра подробностей и фотографий. Ответ должен быть кратким (3-4 предложения) и на русском языке.`, userQuery, resultsText)

	return c.GenerateContent(prompt)
}

// SearchResult represents a search result for chat
type SearchResult struct {
	Type       string
	Title      string
	Text       string
	Location   string
	HasPhotos  bool
	PhotosCount int
}
