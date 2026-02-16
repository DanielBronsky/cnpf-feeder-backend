package gemini

import (
	"fmt"
	"strings"
)

// FishTranslations - словарь переводов названий рыб (русский -> румынский)
var FishTranslations = map[string][]string{
	"плотва":   {"roșioară", "rosioara", "roșcovan"},
	"лещ":      {"plătică", "platica", "platicea"},
	"карп":     {"crap", "caras"},
	"сазан":    {"crap sălbatic", "crap salbatic"},
	"карась":   {"caras"},
	"окунь":    {"biban"},
	"щука":     {"știucă", "stiuca"},
	"судак":    {"șalău", "salau"},
	"сом":      {"somn"},
	"толстолобик": {"tolstolobik", "crap argintiu"},
	"амур":     {"amur"},
	"линь":     {"lin", "lin de balta"},
	"густера":  {"gălbenuș", "galbanus"},
	"подлещик": {"plătică mică", "platica mica"},
	"красноперка": {"roșioară", "rosioara"},
	"уклейка":  {"boarță", "boarta"},
	"ерш":      {"scorpie"},
	"налим":    {"mihalț", "mihalt"},
}

// LocationTranslations - словарь переводов мест рыбалки
var LocationTranslations = map[string][]string{
	"днестр":   {"nistru", "dnestr"},
	"прут":     {"prut"},
	"дубоссары": {"dubăsari", "dubasari"},
	"данчены":  {"dănceni", "danceni"},
	"вадул луй водэ": {"vadul lui vodă", "vadul lui voda"},
	"рыбница":  {"rîbnița", "ribnita", "rybnitsa"},
	"сороки":   {"soroca"},
	"каменка":  {"camenca"},
	"кишинев":  {"chișinău", "chisinau"},
	"озеро":    {"lac", "lacul"},
	"река":     {"râu", "rau", "riul"},
	"пруд":     {"iaz"},
	"водохранилище": {"bazin", "lac de acumulare"},
}

// FishingTerms - рыболовные термины (русский -> румынский)
var FishingTerms = map[string][]string{
	"рыбалка":     {"pescuit"},
	"соревнование": {"competiție", "competitie", "concurs"},
	"отчет":       {"raport"},
	"улов":        {"captură", "captura", "pradă", "prada"},
	"фидер":       {"feeder"},
	"поплавок":    {"plutitor", "dobă", "doba"},
	"спиннинг":    {"spinning"},
	"удочка":      {"undiță", "undita"},
	"прикормка":   {"momeală", "nada", "momeala"},
	"насадка":     {"momitură", "nadă", "momitura", "nada"},
	"крючок":      {"cârlig", "carlig"},
	"леска":       {"fir"},
	"катушка":     {"mulinetă", "mulineta"},
	"берег":       {"mal", "țărm", "tarm"},
	"лодка":       {"barcă", "barca"},
	"этап":        {"etapă", "etapa"},
	"тур":         {"tur"},
	"чемпионат":   {"campionat"},
	"кубок":       {"cupă", "cupa"},
	"зима":        {"iarnă", "iarna"},
	"весна":       {"primăvară", "primavara"},
	"лето":        {"vară", "vara"},
	"осень":       {"toamnă", "toamna"},
}

// MonthTranslations - переводы месяцев
var MonthTranslations = map[string][]string{
	"январь":   {"ianuarie"},
	"февраль":  {"februarie"},
	"март":     {"martie"},
	"апрель":   {"aprilie"},
	"май":      {"mai"},
	"июнь":     {"iunie"},
	"июль":     {"iulie"},
	"август":   {"august"},
	"сентябрь": {"septembrie"},
	"октябрь":  {"octombrie"},
	"ноябрь":   {"noiembrie"},
	"декабрь":  {"decembrie"},
}

// TranslateToRomanian переводит текст на румынский используя словари
// Возвращает оригинальный текст + все румынские варианты
func TranslateToRomanian(text string) []string {
	text = strings.ToLower(text)
	words := strings.Fields(text)
	
	results := []string{text} // Включаем оригинал
	translatedWords := make(map[string]bool)
	
	// Проверяем каждое слово в словарях
	for _, word := range words {
		word = strings.Trim(word, ".,!?;:")
		
		// Проверяем во всех словарях
		allDicts := []map[string][]string{
			FishTranslations,
			LocationTranslations,
			FishingTerms,
			MonthTranslations,
		}
		
		for _, dict := range allDicts {
			if translations, found := dict[word]; found {
				for _, translation := range translations {
					if !translatedWords[translation] {
						results = append(results, translation)
						translatedWords[translation] = true
					}
				}
			}
		}
	}
	
	return results
}

// ExtractKeywords извлекает ключевые слова и переводит их
func ExtractKeywords(text string) []string {
	text = strings.ToLower(text)
	
	// Стоп-слова (служебные слова)
	stopWords := map[string]bool{
		"что": true, "где": true, "когда": true, "как": true,
		"в": true, "на": true, "о": true, "по": true, "и": true,
		"с": true, "для": true, "про": true, "там": true, "тут": true,
		"это": true, "был": true, "была": true, "было": true, "были": true,
		"у": true, "к": true, "из": true, "от": true, "за": true,
		"а": true, "но": true, "или": true, "да": true, "не": true,
		"ли": true, "же": true, "бы": true, "ну": true,
		"ловля": true, "ловил": true, "поймал": true, "рыбачил": true, // глаголы не нужны
	}
	
	// Нормализация склонений (для ключевых слов)
	wordNormalizations := map[string]string{
		"днестре": "днестр",
		"днестра": "днестр",
		"днестром": "днестр",
		"пруте": "прут",
		"прута": "прут",
		"данченах": "данчены",
		"данченами": "данчены",
		"озере": "озеро",
		"озера": "озеро",
		"реке": "река",
		"реки": "река",
		"рекой": "река",
		"леща": "лещ",
		"лещом": "лещ",
		"плотвы": "плотва",
		"плотвой": "плотва",
		"карпа": "карп",
		"карпом": "карп",
		"соревнованиях": "соревнование",
		"соревнованиям": "соревнование",
		"отчете": "отчет",
		"отчета": "отчет",
		"рыбалке": "рыбалка",
		"рыбалки": "рыбалка",
	}
	
	words := strings.Fields(text)
	keywords := make(map[string]bool)
	
	for _, word := range words {
		word = strings.Trim(word, ".,!?;:")
		
		// Пропускаем стоп-слова и короткие слова
		if len(word) < 2 || stopWords[word] {
			continue
		}
		
		// Нормализуем слово (приводим к базовой форме)
		if normalized, found := wordNormalizations[word]; found {
			word = normalized
		}
		
		// Добавляем оригинальное слово
		keywords[word] = true
		
		// Добавляем переводы из словарей
		allDicts := []map[string][]string{
			FishTranslations,
			LocationTranslations,
			FishingTerms,
			MonthTranslations,
		}
		
		for _, dict := range allDicts {
			if translations, found := dict[word]; found {
				for _, translation := range translations {
					keywords[translation] = true
				}
			}
		}
	}
	
	// Преобразуем map в slice
	result := make([]string, 0, len(keywords))
	for keyword := range keywords {
		result = append(result, keyword)
	}
	
	return result
}

// TranslateRussianToRomanian переводит русский текст на румынский через Gemini API
// Используется как fallback когда словаря недостаточно
func (c *Client) TranslateRussianToRomanian(text string) (string, error) {
	prompt := fmt.Sprintf(`Переведи на румынский язык. Верни ТОЛЬКО перевод без объяснений:

%s

Перевод:`, text)
	
	return c.GenerateContent(prompt)
}

// TranslateRomanianToRussian переводит румынский текст на русский через Gemini API
func (c *Client) TranslateRomanianToRussian(text string) (string, error) {
	prompt := fmt.Sprintf(`Переведи на русский язык. Верни ТОЛЬКО перевод без объяснений:

%s

Перевод:`, text)
	
	return c.GenerateContent(prompt)
}
