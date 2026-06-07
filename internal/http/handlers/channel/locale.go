package channel

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/NexaCard/API/internal/models"
)

var htmlTagRe = regexp.MustCompile(`<[^>]*>`)

// resolveLocalizedJSON 从 models.JSON (map[string]interface{}) 按 locale 优先级取值
// 优先 locale → defaultLocale → 第一个非空值
func resolveLocalizedJSON(m models.JSON, locale, defaultLocale string) string {
	if len(m) == 0 {
		return ""
	}
	if v, ok := m[locale]; ok {
		if s := fmt.Sprintf("%v", v); s != "" && s != "<nil>" {
			return s
		}
	}
	if v, ok := m[defaultLocale]; ok {
		if s := fmt.Sprintf("%v", v); s != "" && s != "<nil>" {
			return s
		}
	}
	for _, v := range m {
		if s := fmt.Sprintf("%v", v); s != "" && s != "<nil>" {
			return s
		}
	}
	return ""
}

// stripHTML 剥离 HTML 标签，返回纯文本
func stripHTML(s string) string {
	text := htmlTagRe.ReplaceAllString(s, "")
	// 合并多余空白行
	lines := strings.Split(text, "\n")
	var result []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return strings.Join(result, "\n")
}

// truncate 截取字符串前 n 个 rune，超出则追加 "..."
func truncate(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n]) + "..."
}
