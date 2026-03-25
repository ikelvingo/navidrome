package opencc

import (
	"testing"
)

func TestContainsChinese(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"周杰伦", true},
		{"Hello", false},
		{"Hello世界", true},
		{"", false},
		{"123", false},
		{"周杰倫", true},
	}

	for _, test := range tests {
		result := ContainsChinese(test.input)
		if result != test.expected {
			t.Errorf("ContainsChinese(%q) = %v, expected %v", test.input, result, test.expected)
		}
	}
}

func TestGetSearchQueries(t *testing.T) {
	// 测试非中文查询
	nonChineseQueries := GetSearchQueries("beatles")
	if len(nonChineseQueries) != 1 || nonChineseQueries[0] != "beatles" {
		t.Errorf("Expected single query for non-Chinese, got %v", nonChineseQueries)
	}

	// 测试中文查询（简体）
	simplifiedQueries := GetSearchQueries("周杰伦")
	if len(simplifiedQueries) < 1 {
		t.Errorf("Expected at least one query for Chinese input, got %v", simplifiedQueries)
	}

	// 测试中文查询（繁体）
	traditionalQueries := GetSearchQueries("周杰倫")
	if len(traditionalQueries) < 1 {
		t.Errorf("Expected at least one query for Chinese input, got %v", traditionalQueries)
	}

	t.Logf("Simplified input queries: %v", simplifiedQueries)
	t.Logf("Traditional input queries: %v", traditionalQueries)
}

func TestConvertBoth(t *testing.T) {
	// 测试简体转繁体
	simplified, traditional := ConvertBoth("周杰伦")
	t.Logf("简体: %s, 繁体: %s", simplified, traditional)

	// 测试繁体转简体
	simplified2, traditional2 := ConvertBoth("周杰倫")
	t.Logf("简体: %s, 繁体: %s", simplified2, traditional2)

	// 测试非中文
	s, t2 := ConvertBoth("Hello")
	if s != "Hello" || t2 != "Hello" {
		t.Errorf("Expected same output for non-Chinese input, got %s, %s", s, t2)
	}
}

func TestNormalizeQuery(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello*", "hello"},
		{"  hello  ", "hello"},
		{"hello", "hello"},
		{"", ""},
	}

	for _, test := range tests {
		result := NormalizeQuery(test.input)
		if result != test.expected {
			t.Errorf("NormalizeQuery(%q) = %q, expected %q", test.input, result, test.expected)
		}
	}
}
