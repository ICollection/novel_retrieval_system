package common

import "strings"

// If 三目运算函数。
func If[T any](condition bool, value1 T, value2 T) T {
	if condition {
		return value1
	} else {
		return value2
	}
}

// Trim 除去文本两端空白。
func Trim(text string) *string {
	result := strings.TrimSpace(text)
	if result == "" {
		return nil
	} else {
		return &result
	}
}

// Split 分割文本切片。
func Split(text *string, separtor string) []string {
	if text == nil {
		return []string{}
	} else {
		return strings.Split(*text, separtor)
	}
}

// Join 组合文本切片。
func Join(part []string, separtor string) *string {
	if part == nil || len(part) == 0 {
		return nil
	}
	result := strings.Join(part, separtor)
	return &result
}

// Truncate 截取文本前指定数量的字符。
func Truncate(text string, length int) string {
	if length <= 0 {
		return ""
	}
	runes := []rune(text)
	if len(runes) <= length {
		return text
	} else {
		return string(runes[:length])
	}
}
