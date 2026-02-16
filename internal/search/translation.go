package search

import (
	"strings"
	"unicode"
)

// TranslationMap содержит переводы ключевых слов с русского на румынский
var translationMap = map[string][]string{
	// Соревнования
	"соревнование": {"competitie", "competitii", "etapa", "etape"},
	"соревнования": {"competitie", "competitii", "etapa", "etape"},
	"соревнований": {"competitie", "competitii", "etapa", "etape"}, // Родительный падеж (мн.ч.)
	"соревновании": {"competitie", "competitii", "etapa", "etape"}, // Предложный падеж
	"соревнованию": {"competitie", "competitii", "etapa", "etape"}, // Дательный падеж
	"competitie": {"соревнование", "соревнования", "соревнований", "соревновании"}, // Обратный перевод
	"competitii": {"соревнование", "соревнования", "соревнований", "соревновании"}, // Обратный перевод
	
	"чемпионат": {"campionat", "campionate", "cm"},
	"чемпионаты": {"campionat", "campionate", "cm"},
	"чемпионата": {"campionat", "campionate", "cm"}, // Родительный падеж
	"чемпионате": {"campionat", "campionate", "cm"}, // Предложный падеж
	"чемпионатом": {"campionat", "campionate", "cm"}, // Творительный падеж
	"чемпионат мира": {"campionat mondial", "cm", "campionatul mondial"},
	"чемпионата мира": {"campionat mondial", "cm", "campionatul mondial"}, // Родительный падеж
	"чемпионате мира": {"campionat mondial", "cm", "campionatul mondial"}, // Предложный падеж
	"campionat": {"чемпионат", "чемпионаты", "чемпионата", "чемпионате"}, // Обратный перевод
	"campionate": {"чемпионат", "чемпионаты", "чемпионата", "чемпионате"}, // Обратный перевод
	"cm": {"чемпионат", "чемпионат мира", "чемпионата мира"}, // Сокращение
	
	"этап": {"etapa", "etape"},
	"этапы": {"etapa", "etape"},
	"этапа": {"etapa", "etape"}, // Родительный падеж
	"этапе": {"etapa", "etape"}, // Предложный падеж
	"этапом": {"etapa", "etape"}, // Творительный падеж
	"etapa": {"этап", "этапы", "этапа", "этапе", "этапом"}, // Обратный перевод
	"etape": {"этап", "этапы", "этапа", "этапе", "этапом"}, // Обратный перевод
	
	"регистрация": {"inregistrare", "inscriere", "integistrare"}, // integistrare - опечатка в базе, но нужно учитывать
	"регистрации": {"inregistrare", "inscriere", "integistrare"}, // Множественное число
	"регистрацию": {"inregistrare", "inscriere", "integistrare"}, // Винительный падеж
	"регистрацией": {"inregistrare", "inscriere", "integistrare"}, // Творительный падеж
	"inregistrare": {"регистрация", "регистрации", "регистрацию"}, // Обратный перевод
	"inscriere": {"регистрация", "регистрации", "регистрацию"}, // Обратный перевод
	"integistrare": {"регистрация", "регистрации", "регистрацию"}, // Обратный перевод (с опечаткой)
	
	"издание": {"editie", "editia", "editiile"},
	"издания": {"editie", "editia", "editiile"}, // Множественное число
	"изданию": {"editie", "editia", "editiile"}, // Дательный падеж
	"издании": {"editie", "editia", "editiile"}, // Предложный падеж
	"editia": {"издание", "издания", "изданию", "издании"}, // Обратный перевод
	"editie": {"издание", "издания", "изданию", "издании"}, // Обратный перевод
	"editiile": {"издание", "издания", "изданию", "издании"}, // Обратный перевод
	
	// Отчеты
	"отчет": {"raport", "raporturi"},
	"отчеты": {"raport", "raporturi"},
	"отчета": {"raport", "raporturi"}, // Родительный падеж
	"отчете": {"raport", "raporturi"}, // Предложный падеж
	"отчет о": {"raport", "despre"},
	"отчет по": {"raport", "despre"},
	"материал": {"material", "articol"},
	"материалы": {"material", "articol", "materiale"},
	"материала": {"material", "articol", "materiale"}, // Родительный падеж
	"материале": {"material", "articol", "materiale"}, // Предложный падеж
	"сюжет": {"articol", "material", "raport"},
	"сюжеты": {"articol", "material", "raport", "articole"},
	"сюжета": {"articol", "material", "raport", "articole"}, // Родительный падеж
	"сюжете": {"articol", "material", "raport", "articole"}, // Предложный падеж
	
	// Места (транслитерация и переводы)
	"данчены": {"danceni", "dancheni"},
	"днестр": {"dnestr", "nistru"},
	"днестра": {"dnestr", "nistru"}, // Родительный падеж
	"днестре": {"dnestr", "nistru"}, // Предложный падеж
	"днестром": {"dnestr", "nistru"}, // Творительный падеж
	"nistru": {"днестр", "днестра", "днестре"}, // Обратный перевод
	"dnestr": {"днестр", "днестра", "днестре"}, // Обратный перевод
	
	"пырыта": {"pîrîta", "pyrata"},
	"пырыты": {"pîrîta", "pyrata"}, // Множественное число
	"пырыте": {"pîrîta", "pyrata"}, // Предложный падеж
	"pîrîta": {"пырыта", "пырыты", "пырыте"}, // Обратный перевод
	
	"ципала": {"ţipala", "tipala"},
	"ципалы": {"ţipala", "tipala"}, // Множественное число
	"ципале": {"ţipala", "tipala"}, // Предложный падеж
	"ţipala": {"ципала", "ципалы", "ципале"}, // Обратный перевод
	"tipala": {"ципала", "ципалы", "ципале"}, // Обратный перевод
	
	"хыржаука": {"hîrjauca", "hirjauca"},
	"хыржауки": {"hîrjauca", "hirjauca"}, // Множественное число
	"хыржауке": {"hîrjauca", "hirjauca"}, // Предложный падеж
	"hîrjauca": {"хыржаука", "хыржауки", "хыржауке"}, // Обратный перевод
	"hirjauca": {"хыржаука", "хыржауки", "хыржауке"}, // Обратный перевод
	
	"дамба": {"baraj", "barajul"},
	"дамбу": {"baraj", "barajul"}, // Винительный падеж
	"дамбы": {"baraj", "barajul"}, // Родительный падеж
	"дамбе": {"baraj", "barajul"}, // Предложный падеж
	"дамбой": {"baraj", "barajul"}, // Творительный падеж
	"barajul": {"дамба", "дамбу", "дамбы", "дамбе", "дамбой"}, // Обратный перевод
	"barajului": {"дамба", "дамбу", "дамбы", "дамбе", "дамбой"}, // Родительный падеж (румынский)
	
	// Дамба озера Данчены (специфичное место)
	"дамба озера данчены": {"barajul lacului danceni", "baraj lacului danceni"},
	"дамбы озера данчены": {"barajul lacului danceni", "baraj lacului danceni"},
	"дамбе озера данчены": {"barajul lacului danceni", "baraj lacului danceni"},
	"barajul lacului danceni": {"дамба озера данчены", "дамбы озера данчены", "дамбе озера данчены"}, // Обратный перевод
	
	"озеро": {"lac", "lacul"},
	"озеру": {"lac", "lacul"},
	"озера": {"lac", "lacul"}, // Множественное число
	"озере": {"lac", "lacul"}, // Предложный падеж
	"озером": {"lac", "lacul"}, // Творительный падеж
	"lacul": {"озеро", "озеру", "озера", "озере", "озером"}, // Обратный перевод
	"lacului": {"озеро", "озеру", "озера", "озере", "озером"}, // Родительный падеж (румынский)
	
	// Озеро Данчены (специфичное место)
	"озеро данчены": {"lacul danceni", "lac danceni"},
	"озера данчены": {"lacul danceni", "lac danceni"},
	"озере данчены": {"lacul danceni", "lac danceni"},
	"lacul danceni": {"озеро данчены", "озера данчены", "озере данчены"}, // Обратный перевод
	"lacului danceni": {"озеро данчены", "озера данчены", "озере данчены"}, // Обратный перевод
	
	// Рыбалка и снасти
	"рыбалка": {"pescuit", "pescuitul"},
	"рыбалки": {"pescuit", "pescuitul"}, // Множественное число
	"рыбалке": {"pescuit", "pescuitul"}, // Предложный падеж
	"pescuit": {"рыбалка", "рыбалки", "рыбалке"}, // Обратный перевод
	
	"рыба": {"peşte", "peste"},
	"рыбы": {"peşte", "peste"}, // Множественное число
	"рыбу": {"peşte", "peste"}, // Винительный падеж
	"рыбе": {"peşte", "peste"}, // Предложный падеж
	"peşte": {"рыба", "рыбы", "рыбу", "рыбе"}, // Обратный перевод
	
	"поклевка": {"trasatura", "trasaturi"},
	"поклевки": {"trasatura", "trasaturi"}, // Множественное число
	"поклевок": {"trasatura", "trasaturi"}, // Родительный падеж (мн.ч.)
	"trasaturi": {"поклевка", "поклевки", "поклевок"}, // Обратный перевод
	
	"метод": {"method", "metoda"},
	"метода": {"method", "metoda"}, // Родительный падеж
	"методе": {"method", "metoda"}, // Предложный падеж
	"method": {"метод", "метода", "методе"}, // Обратный перевод
	
	"фидер": {"feeder"},
	"фидера": {"feeder"}, // Родительный падеж
	"фидере": {"feeder"}, // Предложный падеж
	"feeder": {"фидер", "фидера", "фидере"}, // Обратный перевод
	
	"копка": {"copca", "copcă"},
	"копки": {"copca", "copcă"}, // Множественное число
	"копке": {"copca", "copcă"}, // Предложный падеж
	"copcă": {"копка", "копки", "копке"}, // Обратный перевод
	
	"сезон": {"sezon", "sezonul"},
	"сезона": {"sezon", "sezonul"}, // Родительный падеж
	"сезоне": {"sezon", "sezonul"}, // Предложный падеж
	"sezon": {"сезон", "сезона", "сезоне"}, // Обратный перевод
	
	"зима": {"iarnă", "iarna"},
	"зимы": {"iarnă", "iarna"}, // Родительный падеж
	"зиме": {"iarnă", "iarna"}, // Предложный падеж
	"зимой": {"iarnă", "iarna"}, // Творительный падеж
	"iarnă": {"зима", "зимы", "зиме", "зимой"}, // Обратный перевод
	"iarna": {"зима", "зимы", "зиме", "зимой"}, // Обратный перевод
	
	// Бренды и снасти
	"dovit": {"dovit"}, // Бренд, без перевода
	"wafter": {"wafter"}, // Снасть, без перевода
	"пелетс": {"peleti", "peleţi"},
	"пелетсы": {"peleti", "peleţi"}, // Множественное число
	"peleţi": {"пелетс", "пелетсы"}, // Обратный перевод
	
	// Времена года (месяцы)
	"ноябрь": {"noiembrie", "novembrie"},
	"ноября": {"noiembrie", "novembrie"}, // Родительный падеж
	"декабрь": {"decembrie"},
	"декабря": {"decembrie"}, // Родительный падеж
	"январь": {"ianuarie"},
	"января": {"ianuarie"}, // Родительный падеж
	
	// Общие слова (служебные - используются для контекста, но не для строгого поиска)
	"на": {"la", "pe"},
	"в": {"la", "pe", "in"},
	"о": {"despre"},
	"по": {"despre", "pe"},
	"и": {"si"},
	"с": {"cu"},
}

// stopWords содержит служебные слова, которые не должны использоваться для строгого поиска
var stopWords = map[string]bool{
	"на": true, "в": true, "о": true, "по": true, "и": true, "с": true,
	"la": true, "pe": true, "in": true, "despre": true, "si": true, "cu": true,
	"the": true, "a": true, "an": true, "of": true, "and": true, "or": true,
}

// IsStopWord проверяет, является ли слово служебным
func IsStopWord(word string) bool {
	return stopWords[strings.ToLower(word)]
}

// transliterateCyrillic транслитерирует кириллицу в латиницу
func transliterateCyrillic(text string) string {
	translitMap := map[rune]string{
		'а': "a", 'б': "b", 'в': "v", 'г': "g", 'д': "d", 'е': "e", 'ё': "yo",
		'ж': "zh", 'з': "z", 'и': "i", 'й': "y", 'к': "k", 'л': "l", 'м': "m",
		'н': "n", 'о': "o", 'п': "p", 'р': "r", 'с': "s", 'т': "t", 'у': "u",
		'ф': "f", 'х': "h", 'ц': "ts", 'ч': "ch", 'ш': "sh", 'щ': "sch",
		'ъ': "", 'ы': "y", 'ь': "", 'э': "e", 'ю': "yu", 'я': "ya",
		'А': "A", 'Б': "B", 'В': "V", 'Г': "G", 'Д': "D", 'Е': "E", 'Ё': "Yo",
		'Ж': "Zh", 'З': "Z", 'И': "I", 'Й': "Y", 'К': "K", 'Л': "L", 'М': "M",
		'Н': "N", 'О': "O", 'П': "P", 'Р': "R", 'С': "S", 'Т': "T", 'У': "U",
		'Ф': "F", 'Х': "H", 'Ц': "Ts", 'Ч': "Ch", 'Ш': "Sh", 'Щ': "Sch",
		'Ъ': "", 'Ы': "Y", 'Ь': "", 'Э': "E", 'Ю': "Yu", 'Я': "Ya",
	}

	var result strings.Builder
	for _, r := range text {
		if translit, ok := translitMap[r]; ok {
			result.WriteString(translit)
		} else if unicode.IsLetter(r) || unicode.IsDigit(r) || unicode.IsSpace(r) {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// ExpandQuery расширяет поисковый запрос, добавляя переводы и транслитерации
func ExpandQuery(query string) []string {
	query = strings.ToLower(strings.TrimSpace(query))
	if query == "" {
		return []string{query}
	}

	// Собираем все варианты поиска
	variants := make(map[string]bool)
	variants[query] = true // Оригинальный запрос

	// Разбиваем на слова
	words := strings.Fields(query)
	
	// Генерируем варианты для каждого слова
	var expandedWords [][]string
	for _, word := range words {
		wordVariants := []string{word}
		
		// Добавляем переводы
		if translations, ok := translationMap[word]; ok {
			wordVariants = append(wordVariants, translations...)
		}
		
		// Добавляем транслитерацию (если слово содержит кириллицу)
		if containsCyrillic(word) {
			translit := transliterateCyrillic(word)
			if translit != word {
				wordVariants = append(wordVariants, translit)
			}
		}
		
		expandedWords = append(expandedWords, wordVariants)
	}

	// Генерируем все комбинации вариантов слов
	if len(expandedWords) > 0 {
		combinations := generateCombinations(expandedWords)
		for _, combo := range combinations {
			variants[strings.Join(combo, " ")] = true
		}
	}

	// Добавляем транслитерацию всего запроса
	if containsCyrillic(query) {
		translit := transliterateCyrillic(query)
		if translit != query {
			variants[translit] = true
		}
	}

	// Преобразуем map в slice
	result := make([]string, 0, len(variants))
	for variant := range variants {
		if variant != "" {
			result = append(result, variant)
		}
	}

	return result
}

// generateCombinations генерирует все комбинации вариантов слов
func generateCombinations(words [][]string) [][]string {
	if len(words) == 0 {
		return [][]string{{}}
	}

	var result [][]string
	rest := generateCombinations(words[1:])
	
	for _, variant := range words[0] {
		for _, combo := range rest {
			newCombo := make([]string, len(combo)+1)
			newCombo[0] = variant
			copy(newCombo[1:], combo)
			result = append(result, newCombo)
		}
	}
	
	return result
}

// containsCyrillic проверяет, содержит ли строка кириллические символы
func containsCyrillic(s string) bool {
	for _, r := range s {
		if unicode.Is(unicode.Cyrillic, r) {
			return true
		}
	}
	return false
}
