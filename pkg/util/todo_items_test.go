package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCountCheckedItems(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected int
	}{
		{
			name:     "Markdown",
			text:     `[x] Foo [ ]bar [x] fooBar`,
			expected: 2,
		},
		{
			name:     "Last index at eof",
			text:     `[x] Foo [ ]bar [x]`,
			expected: 2,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, CountCheckedItems(test.text))
		})
	}
}

func TestItemChecked(t *testing.T) {
	tests := []struct {
		name      string
		oldText   string
		newText   string
		triggered bool
	}{
		{
			name:      "New Unchecked",
			oldText:   `[x] Foo [ ]bar [x] fooBar`,
			newText:   `[ ] Foo [ ]bar [ ] fooBar`,
			triggered: false,
		},
		{
			name:      "New Unchanged",
			oldText:   `[x] Foo [ ]bar [x] fooBar`,
			newText:   `[x] Foo [ ]bar [x] fooBar`,
			triggered: false,
		},
		{
			name:      "New Checked",
			oldText:   `[x] Foo [ ]bar [x] fooBar`,
			newText:   `[x] Foo [x]bar [x] fooBar`,
			triggered: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.triggered, ItemChecked(test.oldText, test.newText))
		})
	}
}
