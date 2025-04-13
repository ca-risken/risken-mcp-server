package helper

import (
	"testing"
)

func TestParseMCPArgs(t *testing.T) {
	cases := []struct {
		name     string
		key      string
		mcpArgs  map[string]any
		expected any
		wantErr  bool
	}{
		{
			name:     "success string",
			key:      "test_key",
			mcpArgs:  map[string]any{"test_key": "test_value"},
			expected: Pointer("test_value"),
			wantErr:  false,
		},
		{
			name:     "success int",
			key:      "test_key",
			mcpArgs:  map[string]any{"test_key": 123},
			expected: Pointer(123),
			wantErr:  false,
		},
		{
			name:     "success float",
			key:      "test_key",
			mcpArgs:  map[string]any{"test_key": 1.23},
			expected: Pointer(1.23),
			wantErr:  false,
		},
		{
			name:     "success bool",
			key:      "test_key",
			mcpArgs:  map[string]any{"test_key": true},
			expected: Pointer(true),
			wantErr:  false,
		},
		{
			name:     "success array",
			key:      "test_key",
			mcpArgs:  map[string]any{"test_key": []string{"a", "b", "c"}},
			expected: Pointer([]string{"a", "b", "c"}),
			wantErr:  false,
		},
		{
			name:     "key not found",
			key:      "not_exist",
			mcpArgs:  map[string]any{"test_key": "test_value"},
			expected: nil,
			wantErr:  false,
		},
		{
			name:     "type mismatch",
			key:      "test_key",
			mcpArgs:  map[string]any{"test_key": 123},
			expected: nil,
			wantErr:  false,
		},
		{
			name:     "empty map",
			key:      "test_key",
			mcpArgs:  map[string]any{},
			expected: nil,
			wantErr:  false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			switch tc.expected.(type) {
			case *string:
				got, err := ParseMCPArgs[string](tc.key, tc.mcpArgs)
				if (err != nil) != tc.wantErr {
					t.Errorf("ParseMCPArgs() error = %v, wantErr %v", err, tc.wantErr)
					return
				}
				if tc.expected != nil && (got == nil || *got != *tc.expected.(*string)) {
					t.Errorf("ParseMCPArgs() = %v, want %v", got, tc.expected)
				}
			case *int:
				got, err := ParseMCPArgs[int](tc.key, tc.mcpArgs)
				if (err != nil) != tc.wantErr {
					t.Errorf("ParseMCPArgs() error = %v, wantErr %v", err, tc.wantErr)
					return
				}
				if tc.expected != nil && (got == nil || *got != *tc.expected.(*int)) {
					t.Errorf("ParseMCPArgs() = %v, want %v", got, tc.expected)
				}
			case *float64:
				got, err := ParseMCPArgs[float64](tc.key, tc.mcpArgs)
				if (err != nil) != tc.wantErr {
					t.Errorf("ParseMCPArgs() error = %v, wantErr %v", err, tc.wantErr)
					return
				}
				if tc.expected != nil && (got == nil || *got != *tc.expected.(*float64)) {
					t.Errorf("ParseMCPArgs() = %v, want %v", got, tc.expected)
				}
			case *bool:
				got, err := ParseMCPArgs[bool](tc.key, tc.mcpArgs)
				if (err != nil) != tc.wantErr {
					t.Errorf("ParseMCPArgs() error = %v, wantErr %v", err, tc.wantErr)
					return
				}
				if tc.expected != nil && (got == nil || *got != *tc.expected.(*bool)) {
					t.Errorf("ParseMCPArgs() = %v, want %v", got, tc.expected)
				}
			case *[]string:
				got, err := ParseMCPArgs[[]string](tc.key, tc.mcpArgs)
				if (err != nil) != tc.wantErr {
					t.Errorf("ParseMCPArgs() error = %v, wantErr %v", err, tc.wantErr)
					return
				}
				if tc.expected != nil && got != nil {
					expected := tc.expected.(*[]string)
					if len(*got) != len(*expected) {
						t.Errorf("ParseMCPArgs() = %v, want %v", got, tc.expected)
					}
					for i := range *got {
						if (*got)[i] != (*expected)[i] {
							t.Errorf("ParseMCPArgs() = %v, want %v", got, tc.expected)
						}
					}
				}
			}
		})
	}

}
