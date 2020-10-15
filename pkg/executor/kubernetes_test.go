package executor

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatNamePrefix(t *testing.T) {
	const maxRepoNameLen = 10
	prefix := strings.Repeat("p", 253-11-1-maxRepoNameLen-1) + "-"

	tests := []struct {
		name     string
		repo     string
		expected string
	}{
		{
			name:     "Normal Repo Name",
			repo:     strings.Repeat("r", maxRepoNameLen-4) + "/foo",
			expected: prefix + strings.Repeat("r", maxRepoNameLen-4) + "-foo-",
		},
		{
			name:     "Normal Multi-Org Repo Name",
			repo:     strings.Repeat("/r", maxRepoNameLen/2),
			expected: prefix + strings.Repeat("-r", maxRepoNameLen/2) + "-",
		},
		{
			name:     "Long Repo Name with Normal Name",
			repo:     strings.Repeat("r", maxRepoNameLen*2) + "/" + strings.Repeat("r", maxRepoNameLen),
			expected: prefix + strings.Repeat("r", maxRepoNameLen) + "-",
		},
		{
			name:     "Long Repo Name with Long Name",
			repo:     strings.Repeat("r", maxRepoNameLen*2) + "/" + strings.Repeat("r", maxRepoNameLen*2),
			expected: prefix + strings.Repeat("r", maxRepoNameLen) + "-",
		},
		{
			name:     "Long Multi-Org Repo Name",
			repo:     strings.Repeat("r/", maxRepoNameLen*2),
			expected: prefix + strings.Repeat("r-", maxRepoNameLen/2) + "-",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, formatNamePrefix(prefix, test.repo))
		})
	}
}
