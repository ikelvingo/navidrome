package opencc

import (
	"strings"
	"testing"

	"github.com/deluan/sanitize"
)

func TestSearchQueryProcessing(t *testing.T) {
	// 模拟 searching.go 中的处理流程
	testCases := []string{
		"周杰伦", // 简体
		"周杰倫", // 繁体
	}

	for _, original := range testCases {
		// 模拟 searching.go 的处理
		processed := sanitize.Accents(strings.ToLower(strings.TrimSuffix(original, "*")))
		queries := GetSearchQueries(processed)

		t.Logf("原始输入: %s", original)
		t.Logf("处理后: %s", processed)
		t.Logf("查询变体: %v", queries)
		t.Logf("---")
	}
}

func TestDirectConversion(t *testing.T) {
	// 直接测试繁体输入
	traditional := "周杰倫"
	simplified, traditionalOut := ConvertBoth(traditional)

	t.Logf("输入: %s", traditional)
	t.Logf("简体输出: %s", simplified)
	t.Logf("繁体输出: %s", traditionalOut)

	// 直接测试简体输入
	simplifiedIn := "周杰伦"
	simplifiedOut2, traditionalOut2 := ConvertBoth(simplifiedIn)

	t.Logf("输入: %s", simplifiedIn)
	t.Logf("简体输出: %s", simplifiedOut2)
	t.Logf("繁体输出: %s", traditionalOut2)
}

func TestSanitizeAccents(t *testing.T) {
	// 测试 sanitize.Accents 对中文的影响
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
	// 模拟完整的搜索流程
	original := "周杰倫"

	// Step 1: 原始处理
	q := sanitize.Accents(strings.ToLower(strings.TrimSuffix(original, "*")))
	t.Logf("Step 1 - 处理后: %s", q)

	// Step 2: 获取查询变体
	queries := GetSearchQueries(q)
	t.Logf("Step 2 - 查询变体: %v", queries)

	// Step 3: 模拟对每个变体进行搜索
	for _, query := range queries {
		t.Logf("Step 3 - 搜索: %s", query)
	}
}
