package normalization

import (
	"testing"
)

func TestParseLanguageEntry(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectedID  string
		expectedVal string
	}{
		{"English simple", "English", "english", "English"},
		{"Simplified Chinese", "Simplified Chinese", "simplified_chinese", "Simplified Chinese"},
		{"Traditional Chinese", "Traditional Chinese", "traditional_chinese", "Traditional Chinese"},
		{"Simplified Chinese Alt", "Chinese (Simplified)", "simplified_chinese", "Simplified Chinese"},
		{"Traditional Chinese Alt", "Chinese (Traditional)", "traditional_chinese", "Traditional Chinese"},
		{"Spanish Spain", "Spanish - Spain", "spanish_spain", "Spanish - Spain"},
		{"Spanish Latin America", "Spanish - Latin America", "spanish_latam", "Spanish - Latin America"},
		{"Portuguese Brazil", "Portuguese - Brazil", "portuguese_brazil", "Portuguese - Brazil"},
		{"Portuguese Portugal", "Portuguese - Portugal", "portuguese_portugal", "Portuguese - Portugal"},
		{"HTML entity", "Fran&ccedil;ais", "french", "French"},
		{"With asterisk", "English*", "english", "English"},
		{"Japanese kanji", "日本語", "japanese", "Japanese"},
		{"Korean hangul", "한국어", "korean", "Korean"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			info, ok := ParseLanguageEntry(tc.input)
			if !ok {
				t.Fatalf("expected ok=true for input %q", tc.input)
			}
			if info.ID != tc.expectedID {
				t.Errorf("expected ID %q, got %q", tc.expectedID, info.ID)
			}
			if info.DisplayName != tc.expectedVal {
				t.Errorf("expected DisplayName %q, got %q", tc.expectedVal, info.DisplayName)
			}
		})
	}
}

func TestParseLanguagesList(t *testing.T) {
	input := []string{
		"English",
		"French",
		"English", // duplicate
		"<strong>*languages with full audio support</strong>",
		"Simplified Chinese, Traditional Chinese",
		"Spanish - Spain; Spanish - Latin America",
	}

	results := ParseLanguagesList(input)
	expectedIDs := []string{
		"english",
		"french",
		"simplified_chinese",
		"traditional_chinese",
		"spanish_spain",
		"spanish_latam",
	}

	if len(results) != len(expectedIDs) {
		t.Fatalf("expected %d results, got %d: %+v", len(expectedIDs), len(results), results)
	}

	for i, exp := range expectedIDs {
		if results[i].ID != exp {
			t.Errorf("index %d: expected %q, got %q", i, exp, results[i].ID)
		}
	}
}

func TestParseLanguagesFromString(t *testing.T) {
	raw := "[English, French, German, Italian, Spanish - Spain]"
	results := ParseLanguagesFromString(raw)
	if len(results) != 5 {
		t.Fatalf("expected 5 languages, got %d", len(results))
	}
	if results[0].ID != "english" || results[4].ID != "spanish_spain" {
		t.Errorf("unexpected parsing: %+v", results)
	}
}
