package textfa

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestNormalize(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"arabic ye and kaf", "استامينوفن", "استامینوفن"},
		{"persian ye already", "استامینوفن", "استامینوفن"},
		{"zwnj becomes space", "ضد\u200cچروک", "ضد چروک"},
		{"persian digits", "کرم ۵۰۰", "کرم 500"},
		{"arabic digits", "کرم ٥٠٠", "کرم 500"},
		{"extra spaces", "  کرم   دور   چشم  ", "کرم دور چشم"},
		{"zero width dropped", "کرم\u200bدورچشم", "کرمدورچشم"},
		{"control chars dropped", "استامینوفن\u0001", "استامینوفن"},
		{"empty", "", ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := Normalize(tc.input); got != tc.want {
				t.Fatalf("Normalize(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestQueryFoldsSpellings(t *testing.T) {
	wantFa := Query("استامینوفن")
	if got := Query("استامينوفن"); got != wantFa {
		t.Fatalf("arabic ye: %q vs %q", got, wantFa)
	}
	if Query("استا مینوفن") == "" {
		t.Fatal("spaced query should survive normalisation")
	}
	if Query("Acetaminophen") != "acetaminophen" {
		t.Fatalf("latin fold = %q", Query("Acetaminophen"))
	}
}

func TestQueryRejectsBlankAndCapsLength(t *testing.T) {
	if Query("   \n\t  ") != "" {
		t.Fatal("blank query")
	}
	long := strings.Repeat("ا", MaxQueryRunes+20)
	if got := Query(long); utf8.RuneCountInString(got) != MaxQueryRunes {
		t.Fatalf("len = %d", utf8.RuneCountInString(got))
	}
}

func TestDocumentJoinsNormalizedFields(t *testing.T) {
	got := Document("استامينوفن ۵۰۰", "Acetaminophen", "استامینوفن ۵۰۰")
	if !strings.Contains(got, "استامینوفن 500") || !strings.Contains(got, "acetaminophen") {
		t.Fatalf("document = %q", got)
	}
}

func TestSlugify(t *testing.T) {
	got := Slugify("Cream Dr Jila Eye Contour 1664")
	want := "cream-dr-jila-eye-contour-1664"
	if got != want {
		t.Fatalf("Slugify = %q, want %q", got, want)
	}
}
