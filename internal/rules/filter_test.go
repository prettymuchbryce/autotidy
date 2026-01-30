package rules

import (
	"errors"
	"testing"

	"github.com/prettymuchbryce/autotidy/internal/testutil"
)

// mockEvaluable is a test helper that implements Evaluable.
type mockEvaluable struct {
	result bool
	err    error
}

func (m *mockEvaluable) Evaluate(path string) (bool, error) {
	return m.result, m.err
}

func TestFilter_Evaluate(t *testing.T) {
	testPath := testutil.Path("/", "test", "path")

	tests := []struct {
		name     string
		inner    Evaluable
		expected bool
		wantErr  bool
	}{
		{
			name:     "delegates to inner returning true",
			inner:    &mockEvaluable{result: true},
			expected: true,
			wantErr:  false,
		},
		{
			name:     "delegates to inner returning false",
			inner:    &mockEvaluable{result: false},
			expected: false,
			wantErr:  false,
		},
		{
			name:    "delegates error from inner",
			inner:   &mockEvaluable{err: errors.New("test error")},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &Filter{
				Name:  "test",
				Inner: tt.inner,
			}
			got, err := f.Evaluate(testPath)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if got != tt.expected {
				t.Errorf("Evaluate() = %v, want %v", got, tt.expected)
			}
		})
	}
}
