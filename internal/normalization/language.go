package normalization

import (
	"html"
	"regexp"
	"strings"
)

// LanguageInfo represents a normalized language with canonical ID and display name.
type LanguageInfo struct {
	ID          string // lowercase canonical snake_case id, e.g. "simplified_chinese"
	DisplayName string // Canonical human readable name, e.g. "Simplified Chinese"
	RawLabel    string // Original parsed raw token
}

var (
	htmlTagRegex    = regexp.MustCompile(`<[^>]*>`)
	extraNotesRegex = regexp.MustCompile(`(?i)\*languages with full audio support\*?|\*full audio\*?`)
	bracketRegex    = regexp.MustCompile(`\[.*?\]`)
)

// LanguageMapping holds predefined canonical mappings.
type languageDefinition struct {
	id          string
	displayName string
	aliases     []string
}

var knownLanguages = []languageDefinition{
	{id: "english", displayName: "English", aliases: []string{"english", "en"}},
	{id: "simplified_chinese", displayName: "Simplified Chinese", aliases: []string{"simplified chinese", "chinese simplified", "chinese (simplified)", "simplified-chinese", "zh-cn", "zh_hans", "简体中文"}},
	{id: "traditional_chinese", displayName: "Traditional Chinese", aliases: []string{"traditional chinese", "chinese traditional", "chinese (traditional)", "traditional-chinese", "zh-tw", "zh_hant", "繁體中文"}},
	{id: "french", displayName: "French", aliases: []string{"french", "français", "francais", "fr"}},
	{id: "german", displayName: "German", aliases: []string{"german", "deutsch", "de"}},
	{id: "spanish_spain", displayName: "Spanish - Spain", aliases: []string{"spanish - spain", "spanish (spain)", "spanish", "español - españa", "español", "es"}},
	{id: "spanish_latam", displayName: "Spanish - Latin America", aliases: []string{"spanish - latin america", "spanish (latin america)", "español - latinoamérica", "español latino", "es-419", "spanish-latam"}},
	{id: "italian", displayName: "Italian", aliases: []string{"italian", "italiano", "it"}},
	{id: "japanese", displayName: "Japanese", aliases: []string{"japanese", "日本語", "ja"}},
	{id: "korean", displayName: "Korean", aliases: []string{"korean", "한국어", "ko"}},
	{id: "russian", displayName: "Russian", aliases: []string{"russian", "русский", "ru"}},
	{id: "portuguese_brazil", displayName: "Portuguese - Brazil", aliases: []string{"portuguese - brazil", "portuguese (brazil)", "português - brasil", "pt-br", "brazilian portuguese"}},
	{id: "portuguese_portugal", displayName: "Portuguese - Portugal", aliases: []string{"portuguese - portugal", "portuguese (portugal)", "portuguese", "português - portugal", "português", "pt-pt", "pt"}},
	{id: "polish", displayName: "Polish", aliases: []string{"polish", "polski", "pl"}},
	{id: "turkish", displayName: "Turkish", aliases: []string{"turkish", "türkçe", "turkce", "tr"}},
	{id: "dutch", displayName: "Dutch", aliases: []string{"dutch", "nederlands", "nl"}},
	{id: "ukrainian", displayName: "Ukrainian", aliases: []string{"ukrainian", "українська", "uk"}},
	{id: "czech", displayName: "Czech", aliases: []string{"czech", "čeština", "cestina", "cs"}},
	{id: "swedish", displayName: "Swedish", aliases: []string{"swedish", "svenska", "sv"}},
	{id: "hungarian", displayName: "Hungarian", aliases: []string{"hungarian", "magyar", "hu"}},
	{id: "arabic", displayName: "Arabic", aliases: []string{"arabic", "العربية", "ar"}},
	{id: "thai", displayName: "Thai", aliases: []string{"thai", "ไทย", "th"}},
	{id: "vietnamese", displayName: "Vietnamese", aliases: []string{"vietnamese", "tiếng việt", "vi"}},
	{id: "danish", displayName: "Danish", aliases: []string{"danish", "dansk", "da"}},
	{id: "finnish", displayName: "Finnish", aliases: []string{"finnish", "suomi", "fi"}},
	{id: "norwegian", displayName: "Norwegian", aliases: []string{"norwegian", "norsk", "no"}},
	{id: "romanian", displayName: "Romanian", aliases: []string{"romanian", "română", "romana", "ro"}},
	{id: "greek", displayName: "Greek", aliases: []string{"greek", "ελληνικά", "el"}},
	{id: "bulgarian", displayName: "Bulgarian", aliases: []string{"bulgarian", "български", "bg"}},
	{id: "hindi", displayName: "Hindi", aliases: []string{"hindi", "हिन्दी", "hi"}},
	{id: "indonesian", displayName: "Indonesian", aliases: []string{"indonesian", "bahasa indonesia", "id"}},
	{id: "hebrew", displayName: "Hebrew", aliases: []string{"hebrew", "עברית", "he"}},
	{id: "malay", displayName: "Malay", aliases: []string{"malay", "bahasa melayu", "ms"}},
	{id: "persian", displayName: "Persian", aliases: []string{"persian", "farsi", "fa"}},
	{id: "catalan", displayName: "Catalan", aliases: []string{"catalan", "català", "ca"}},
	{id: "croatian", displayName: "Croatian", aliases: []string{"croatian", "hrvatski", "hr"}},
	{id: "serbian", displayName: "Serbian", aliases: []string{"serbian", "srpski", "sr"}},
	{id: "slovak", displayName: "Slovak", aliases: []string{"slovak", "slovenčina", "sk"}},
	{id: "slovenian", displayName: "Slovenian", aliases: []string{"slovenian", "slovenščina", "sl"}},
	{id: "estonian", displayName: "Estonian", aliases: []string{"estonian", "eesti", "et"}},
	{id: "latvian", displayName: "Latvian", aliases: []string{"latvian", "latviešu", "lv"}},
	{id: "lithuanian", displayName: "Lithuanian", aliases: []string{"lithuanian", "lietuvių", "lt"}},
	{id: "belarusian", displayName: "Belarusian", aliases: []string{"belarusian", "беларуская", "be"}},
	{id: "georgian", displayName: "Georgian", aliases: []string{"georgian", "ქართული", "ka"}},
	{id: "armenian", displayName: "Armenian", aliases: []string{"armenian", "հայերენ", "hy"}},
	{id: "kazakh", displayName: "Kazakh", aliases: []string{"kazakh", "қазақ", "kk"}},
	{id: "tagalog", displayName: "Tagalog", aliases: []string{"tagalog", "filipino", "tl"}},
	{id: "bengali", displayName: "Bengali", aliases: []string{"bengali", "বাংলা", "bn"}},
	{id: "urdu", displayName: "Urdu", aliases: []string{"urdu", "اردو", "ur"}},
	{id: "afrikaans", displayName: "Afrikaans", aliases: []string{"afrikaans", "af"}},
	{id: "basque", displayName: "Basque", aliases: []string{"basque", "euskara", "eu"}},
	{id: "galician", displayName: "Galician", aliases: []string{"galician", "galego", "gl"}},
	{id: "icelandic", displayName: "Icelandic", aliases: []string{"icelandic", "íslenska", "is"}},
	{id: "irish", displayName: "Irish", aliases: []string{"irish", "gaeilge", "ga"}},
	{id: "welsh", displayName: "Welsh", aliases: []string{"welsh", "cymraeg", "cy"}},
	{id: "albanian", displayName: "Albanian", aliases: []string{"albanian", "shqip", "sq"}},
	{id: "macedonian", displayName: "Macedonian", aliases: []string{"macedonian", "македонски", "mk"}},
}

var aliasMap map[string]languageDefinition
var idToDisplayName map[string]string

func init() {
	aliasMap = make(map[string]languageDefinition)
	idToDisplayName = make(map[string]string)
	for _, def := range knownLanguages {
		idToDisplayName[def.id] = def.displayName
		aliasMap[def.id] = def
		aliasMap[strings.ToLower(def.displayName)] = def
		for _, alias := range def.aliases {
			aliasMap[strings.ToLower(strings.TrimSpace(alias))] = def
		}
	}
}

// GetLanguageDisplayName returns the canonical clean display name for a language ID.
func GetLanguageDisplayName(id string) string {
	if name, ok := idToDisplayName[id]; ok {
		return name
	}
	return toTitleCase(strings.ReplaceAll(id, "_", " "))
}

// CleanLanguageString cleans an individual raw string or token.
func CleanLanguageString(input string) string {
	s := html.UnescapeString(input)
	s = htmlTagRegex.ReplaceAllString(s, " ")
	s = extraNotesRegex.ReplaceAllString(s, "")
	s = strings.ReplaceAll(s, "*", "")
	s = strings.ReplaceAll(s, "'", "")
	s = strings.ReplaceAll(s, `"`, "")
	s = strings.Trim(s, " \t\r\n&;,.")
	return s
}

// ParseLanguageEntry normalizes a single language label into a canonical LanguageInfo.
func ParseLanguageEntry(raw string) (LanguageInfo, bool) {
	cleaned := CleanLanguageString(raw)
	if cleaned == "" {
		return LanguageInfo{}, false
	}

	normKey := strings.ToLower(cleaned)
	normKey = strings.Trim(normKey, " ,;./-_*&")

	if def, ok := aliasMap[normKey]; ok {
		return LanguageInfo{
			ID:          def.id,
			DisplayName: def.displayName,
			RawLabel:    raw,
		}, true
	}

	// Dynamic fallback for any unknown or exotic language:
	id := toSnakeCase(cleaned)
	if id == "" {
		return LanguageInfo{}, false
	}

	return LanguageInfo{
		ID:          id,
		DisplayName: toTitleCase(cleaned),
		RawLabel:    raw,
	}, true
}

// ParseLanguagesList parses a slice or raw string into deduplicated normalized languages.
func ParseLanguagesList(items []string) []LanguageInfo {
	if len(items) == 0 {
		return nil
	}

	seen := make(map[string]bool)
	var result []LanguageInfo

	for _, item := range items {
		tokens := splitLanguageTokens(item)
		for _, token := range tokens {
			info, ok := ParseLanguageEntry(token)
			if !ok {
				continue
			}
			if !seen[info.ID] {
				seen[info.ID] = true
				result = append(result, info)
			}
		}
	}

	return result
}

// ParseLanguagesFromString parses a string that might be comma-separated, bracket-delimited, or HTML-tagged.
func ParseLanguagesFromString(str string) []LanguageInfo {
	cleaned := html.UnescapeString(str)
	cleaned = htmlTagRegex.ReplaceAllString(cleaned, ",")
	cleaned = extraNotesRegex.ReplaceAllString(cleaned, "")
	cleaned = strings.Trim(cleaned, "[]")

	tokens := splitLanguageTokens(cleaned)
	return ParseLanguagesList(tokens)
}

func splitLanguageTokens(s string) []string {
	s = strings.ReplaceAll(s, "\r\n", ",")
	s = strings.ReplaceAll(s, "\n", ",")
	s = strings.ReplaceAll(s, ";", ",")

	var tokens []string
	for _, part := range strings.Split(s, ",") {
		part = strings.Trim(part, " \t\r\n*")
		if part != "" {
			tokens = append(tokens, part)
		}
	}
	return tokens
}

func toSnakeCase(s string) string {
	s = strings.ToLower(s)
	reg := regexp.MustCompile(`[^a-z0-9]+`)
	s = reg.ReplaceAllString(s, "_")
	return strings.Trim(s, "_")
}

func toTitleCase(s string) string {
	parts := strings.Fields(s)
	for i, p := range parts {
		if len(p) > 0 {
			parts[i] = strings.ToUpper(p[:1]) + strings.ToLower(p[1:])
		}
	}
	return strings.Join(parts, " ")
}
