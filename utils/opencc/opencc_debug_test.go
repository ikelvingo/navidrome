package opencc

import (
	"strings"
	"testing"

	"github.com/deluan/sanitize"
)

func TestSearchQueryProcessing(t *testing.T) {
	// Simulate the processing flow from searching.go
	testCases := []string{
		"周杰伦", // simplified
		"周杰倫", // traditional
	}

	for _, original := range testCases {
		// Simulate searching.go preprocessing
		processed := sanitize.Accents(strings.ToLower(strings.TrimSuffix(original, "*")))
		queries := GetSearchQueries(processed)

		t.Logf("original: %s", original)
		t.Logf("processed: %s", processed)
		t.Logf("variants: %v", queries)
		t.Logf("---")
	}
}

func TestDirectConversion(t *testing.T) {
	// Test traditional input directly
	traditional := "周杰倫"
	simplified, traditionalOut := ConvertBoth(traditional)

	t.Logf("input: %s", traditional)
	t.Logf("simplified output: %s", simplified)
	t.Logf("traditional output: %s", traditionalOut)

	// Test simplified input directly
	simplifiedIn := "周杰伦"
	simplifiedOut2, traditionalOut2 := ConvertBoth(simplifiedIn)

	t.Logf("input: %s", simplifiedIn)
	t.Logf("simplified output: %s", simplifiedOut2)
	t.Logf("traditional output: %s", traditionalOut2)
}

func TestSanitizeAccents(t *testing.T) {
	// Test the effect of sanitize.Accents on Chinese text
	testCases := []string{
		"周杰伦",
		"周杰倫",
		"Hello",
		"Café",
	}

	for _, s := range testCases {
		result := sanitize.Accents(s)
		t.Logf("sanitize.Accents(%q) = %q", s, result)
	}
}

func TestFullSearchFlow(t *testing.T) {
	// Simulate the complete search flow
	original := "周杰倫"

	// Step 1: Original preprocessing
	q := sanitize.Accents(strings.ToLower(strings.TrimSuffix(original, "*")))
	t.Logf("Step 1 - preprocessed: %s", q)

	// Step 2: Get query variants
	queries := GetSearchQueries(q)
	t.Logf("Step 2 - query variants: %v", queries)

	// Step 3: Simulate searching each variant
	for _, query := range queries {
		t.Logf("Step 3 - searching: %s", query)
	}
}
