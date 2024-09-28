package common

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBoolToInteger(t *testing.T) {
	cases := []struct {
		name          string
		input         bool
		expectedValue int
	}{
		{
			name:          "true",
			input:         true,
			expectedValue: 1,
		},
		{
			name:          "false",
			input:         false,
			expectedValue: 0,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			val := BoolToInteger(c.input)
			require.Equal(t, c.expectedValue, val)
		})
	}
}
