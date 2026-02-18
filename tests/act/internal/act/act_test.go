package act

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAnnotation_ParseLogFmtMessage(t *testing.T) {
	tests := []struct {
		name     string
		message  string
		expected map[string]string
	}{
		{
			name:    "single key-value pair",
			message: `key=value`,
			expected: map[string]string{
				"key": "value",
			},
		},
		{
			name:    "multiple key-value pairs",
			message: `key1=value1 key2=value2`,
			expected: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
		},
		{
			name:    "values with spaces (quoted)",
			message: `key1="value with spaces" key2="another value"`,
			expected: map[string]string{
				"key1": "value with spaces",
				"key2": "another value",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			annotation := Annotation{Message: tt.message}
			result := annotation.ParseLogFmtMessage()
			require.Equal(t, tt.expected, result)
		})
	}
}
