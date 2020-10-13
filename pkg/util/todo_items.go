package util

import "strings"

func ItemChecked(oldBody, newBody string) bool {
	return CountCheckedItems(oldBody) < CountCheckedItems(newBody)
}

func CountCheckedItems(s string) int {
	oldCount := 0
	lastIndex := 0
	for i := strings.Index(s, "[x]"); i >= 0; i = strings.Index(s[lastIndex:], "[x]") {
		oldCount++
		lastIndex += i + 3
	}
	return oldCount
}
