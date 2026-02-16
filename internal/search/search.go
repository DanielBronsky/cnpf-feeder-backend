package search

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// SearchResult represents a search result
type SearchResult struct {
	Type        string
	ID          string
	Title       string
	Text        string
	Location    string
	HasPhotos   bool
	PhotosCount int
}

// scoredResult represents a search result with relevance score
type scoredResult struct {
	result SearchResult
	score  int
}

// SearchReports searches for reports matching the query
// Query should already be optimized by AI (key words extracted, translated, transliterated)
// Uses relevance scoring to filter out irrelevant results
func SearchReports(ctx context.Context, db *mongo.Database, query string) ([]SearchResult, error) {
	// Query is already optimized by AI, just extract key words
	keyWords := extractKeyWordsFromQuery(query)
	if len(keyWords) == 0 {
		return []SearchResult{}, nil
	}

	// Build regex patterns for key words only
	var orConditions []bson.M
	for _, word := range keyWords {
		if word == "" {
			continue
		}
		regex := bson.M{"$regex": regexp.QuoteMeta(word), "$options": "i"}
		orConditions = append(orConditions,
			bson.M{"title": regex},
			bson.M{"text": regex},
		)
	}

	if len(orConditions) == 0 {
		return []SearchResult{}, nil
	}

	cursor, err := db.Collection("reports").Find(ctx, bson.M{
		"$or": orConditions,
	}, options.Find().SetLimit(20).SetSort(bson.M{"createdAt": -1})) // Get more results for filtering
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	// Score and filter results by relevance
	var scoredResults []scoredResult

	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			continue
		}

		var id string
		if objID, ok := doc["_id"].(primitive.ObjectID); ok {
			id = objID.Hex()
		} else {
			continue
		}

		title := getString(doc, "title")
		text := getString(doc, "text")
		
		// Calculate relevance score
		content := title + " " + text
		score := calculateRelevanceScore(content, keyWords)
		matchingWordsCount := countMatchingWords(content, keyWords)
		
		// Require matching words based on query length:
		// - 1-2 words: require all words to match
		// - 3+ words: require at least 2 words to match
		minRequiredWords := 2
		if len(keyWords) <= 2 {
			minRequiredWords = len(keyWords) // Require all words for short queries
		}
		
		// Only include results with sufficient relevance
		if score > 0 && matchingWordsCount >= minRequiredWords {
			photos, _ := doc["photos"].(bson.A)
			photosCount := len(photos)

			scoredResults = append(scoredResults, scoredResult{
				result: SearchResult{
					Type:        "report",
					ID:          id,
					Title:       title,
					Text:        text,
					HasPhotos:   photosCount > 0,
					PhotosCount: photosCount,
				},
				score: score,
			})
		}
	}

	// Sort by score (descending) and take top 5
	results := sortAndLimitResults(scoredResults, 5)

	return results, nil
}

// SearchCompetitions searches for competitions matching the query
// Query should already be optimized by AI (key words extracted, translated, transliterated)
// Uses relevance scoring to filter out irrelevant results
func SearchCompetitions(ctx context.Context, db *mongo.Database, query string) ([]SearchResult, error) {
	// Query is already optimized by AI, just extract key words
	keyWords := extractKeyWordsFromQuery(query)
	if len(keyWords) == 0 {
		return []SearchResult{}, nil
	}

	// Build regex patterns for key words only
	var orConditions []bson.M
	for _, word := range keyWords {
		if word == "" {
			continue
		}
		regex := bson.M{"$regex": regexp.QuoteMeta(word), "$options": "i"}
		orConditions = append(orConditions,
			bson.M{"title": regex},
			bson.M{"location": regex},
		)
	}

	if len(orConditions) == 0 {
		return []SearchResult{}, nil
	}

	cursor, err := db.Collection("competitions").Find(ctx, bson.M{
		"$or": orConditions,
	}, options.Find().SetLimit(20).SetSort(bson.M{"createdAt": -1})) // Get more results for filtering
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	// Score and filter results by relevance
	var scoredResults []scoredResult

	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			continue
		}

		var id string
		if objID, ok := doc["_id"].(primitive.ObjectID); ok {
			id = objID.Hex()
		} else {
			continue
		}

		title := getString(doc, "title")
		location := getString(doc, "location")
		content := title + " " + location

		// Calculate relevance score
		score := calculateRelevanceScore(content, keyWords)
		
		// Only include results with score > 0 (at least one key word found)
		if score > 0 {
			scoredResults = append(scoredResults, scoredResult{
				result: SearchResult{
					Type:     "competition",
					ID:       id,
					Title:    title,
					Location: location,
				},
				score: score,
			})
		}
	}

	// Sort by score (descending) and take top 5
	results := sortAndLimitResults(scoredResults, 5)

	return results, nil
}

// SearchAll searches both reports and competitions
// If query is empty or same as original, also tries fallback with ExpandQuery for compatibility
func SearchAll(ctx context.Context, db *mongo.Database, query string) ([]SearchResult, error) {
	// If query is empty, return empty results
	if strings.TrimSpace(query) == "" {
		return []SearchResult{}, nil
	}

	reports, err1 := SearchReports(ctx, db, query)
	competitions, err2 := SearchCompetitions(ctx, db, query)

	if err1 != nil {
		return nil, err1
	}
	if err2 != nil {
		return nil, err2
	}

	results := append(reports, competitions...)
	
	// If no results found and query might need expansion, try fallback
	if len(results) == 0 {
		// Try with expanded query variants as fallback
		queryVariants := ExpandQuery(query)
		for _, variant := range queryVariants {
			if variant == query {
				continue // Skip if same as original
			}
			fallbackReports, _ := SearchReports(ctx, db, variant)
			fallbackCompetitions, _ := SearchCompetitions(ctx, db, variant)
			if len(fallbackReports) > 0 || len(fallbackCompetitions) > 0 {
				results = append(fallbackReports, fallbackCompetitions...)
				break // Found results, stop trying variants
			}
		}
	}

	return results, nil
}

func getString(doc bson.M, key string) string {
	if val, ok := doc[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

// extractKeyWordsFromQuery extracts meaningful words from a single optimized query
// AI should have already removed stop words, but we filter them again just in case
func extractKeyWordsFromQuery(query string) []string {
	keyWordsMap := make(map[string]bool)
	
	words := strings.Fields(strings.ToLower(strings.TrimSpace(query)))
	addDateTokens(words, keyWordsMap)
	for _, word := range words {
		// Skip empty words and stop words (AI should have removed them, but double-check)
		if word == "" || IsStopWord(word) {
			continue
		}
		// Skip very short words (less than 2 characters)
		if len(word) < 2 {
			continue
		}
		keyWordsMap[word] = true
	}
	
	// Convert map to slice
	keyWords := make([]string, 0, len(keyWordsMap))
	for word := range keyWordsMap {
		keyWords = append(keyWords, word)
	}
	
	return keyWords
}

func addDateTokens(words []string, keyWordsMap map[string]bool) {
	// Supports queries like:
	// - "18 января"
	// - "18 ianuarie"
	// - "18.01.2026"
	// Adds tokens like "18.01" so it matches titles formatted as DD.MM.YYYY.

	monthNameToNumber := map[string]int{
		// Russian (common forms)
		"январь": 1, "января": 1, "январе": 1,
		"февраль": 2, "февраля": 2, "феврале": 2,
		"март": 3, "марта": 3, "марте": 3,
		"апрель": 4, "апреля": 4, "апреле": 4,
		"май": 5, "мая": 5, "мае": 5,
		"июнь": 6, "июня": 6, "июне": 6,
		"июль": 7, "июля": 7, "июле": 7,
		"август": 8, "августа": 8, "августе": 8,
		"сентябрь": 9, "сентября": 9, "сентябре": 9,
		"октябрь": 10, "октября": 10, "октябре": 10,
		"ноябрь": 11, "ноября": 11, "ноябре": 11,
		"декабрь": 12, "декабря": 12, "декабре": 12,
		// Romanian
		"ianuarie": 1,
		"februarie": 2,
		"martie": 3,
		"aprilie": 4,
		"mai": 5,
		"iunie": 6,
		"iulie": 7,
		"august": 8,
		"septembrie": 9,
		"octombrie": 10,
		"noiembrie": 11,
		"decembrie": 12,
		// Romanian variants we might see in titles / text
		"novembrie": 11,
	}

	var day, month, year int
	day = 0
	month = 0
	year = 0

	for _, w := range words {
		w = strings.TrimSpace(w)
		if w == "" {
			continue
		}

		// Try parse year/day/month from numeric tokens.
		if n, err := strconv.Atoi(trimNonDigits(w)); err == nil {
			if n >= 1900 && n <= 2100 {
				year = n
				continue
			}
			if n >= 1 && n <= 31 && day == 0 {
				day = n
				continue
			}
			if n >= 1 && n <= 12 && month == 0 {
				month = n
				continue
			}
		}

		// Month by name.
		if m, ok := monthNameToNumber[w]; ok && month == 0 {
			month = m
			continue
		}

		// Try parse formats like 18.01.2026 / 18-01-2026 / 18/01/2026
		if strings.Contains(w, ".") || strings.Contains(w, "-") || strings.Contains(w, "/") {
			parts := splitDateParts(w)
			if len(parts) >= 2 {
				if d, err := strconv.Atoi(parts[0]); err == nil && d >= 1 && d <= 31 && day == 0 {
					day = d
				}
				if m, err := strconv.Atoi(parts[1]); err == nil && m >= 1 && m <= 12 && month == 0 {
					month = m
				}
				if len(parts) >= 3 {
					if y, err := strconv.Atoi(parts[2]); err == nil && y >= 1900 && y <= 2100 && year == 0 {
						year = y
					}
				}
			}
		}
	}

	if day == 0 || month == 0 {
		return
	}

	dd := fmt.Sprintf("%02d", day)
	mm := fmt.Sprintf("%02d", month)

	// Most common title format in this project: DD.MM.YYYY
	keyWordsMap[dd+"."+mm] = true
	keyWordsMap[dd+"-"+mm] = true
	keyWordsMap[dd+"/"+mm] = true

	if year != 0 {
		yyyy := fmt.Sprintf("%04d", year)
		keyWordsMap[dd+"."+mm+"."+yyyy] = true
		keyWordsMap[dd+"-"+mm+"-"+yyyy] = true
		keyWordsMap[dd+"/"+mm+"/"+yyyy] = true
	}
}

func trimNonDigits(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func splitDateParts(s string) []string {
	s = strings.ReplaceAll(s, "-", ".")
	s = strings.ReplaceAll(s, "/", ".")
	parts := strings.Split(s, ".")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		out = append(out, p)
	}
	return out
}

// calculateRelevanceScore calculates how relevant a document is to the search query
// Returns score based on number of key words found and their positions
func calculateRelevanceScore(content string, keyWords []string) int {
	contentLower := strings.ToLower(content)
	score := 0
	
	for _, word := range keyWords {
		// Count occurrences of the word
		count := strings.Count(contentLower, word)
		if count > 0 {
			// Base score for finding the word
			score += count * 2
			
			// Bonus if word appears in title (first part of content)
			// Assuming title is first ~100 characters
			if len(content) > 100 {
				titlePart := contentLower[:100]
				if strings.Contains(titlePart, word) {
					score += 5 // Bonus for title match
				}
			} else if strings.Contains(contentLower, word) {
				score += 5 // Bonus if entire content is short (likely title)
			}
		}
	}
	
	return score
}

// countMatchingWords counts how many key words are found in content
func countMatchingWords(content string, keyWords []string) int {
	contentLower := strings.ToLower(content)
	count := 0
	
	for _, word := range keyWords {
		if strings.Contains(contentLower, word) {
			count++
		}
	}
	
	return count
}

// sortAndLimitResults sorts results by score and returns top N
func sortAndLimitResults(scoredResults []scoredResult, limit int) []SearchResult {
	if len(scoredResults) == 0 {
		return []SearchResult{}
	}
	
	// Simple bubble sort by score (descending) - fine for small datasets
	for i := 0; i < len(scoredResults)-1; i++ {
		for j := 0; j < len(scoredResults)-i-1; j++ {
			if scoredResults[j].score < scoredResults[j+1].score {
				scoredResults[j], scoredResults[j+1] = scoredResults[j+1], scoredResults[j]
			}
		}
	}
	
	// Take top N results
	resultCount := limit
	if len(scoredResults) < limit {
		resultCount = len(scoredResults)
	}
	
	results := make([]SearchResult, 0, resultCount)
	for i := 0; i < resultCount; i++ {
		results = append(results, scoredResults[i].result)
	}
	
	return results
}
