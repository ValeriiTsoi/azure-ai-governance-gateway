package governance

import "testing"

func TestEvaluatePolicy(t *testing.T) {
	tests := []struct {
		classification string
		expected       string
	}{
		{"public", "allow"},
		{"internal", "allow"},
		{"confidential", "review"},
		{"restricted", "deny"},
	}

	for _, test := range tests {
		t.Run(test.classification, func(t *testing.T) {
			result := EvaluatePolicy(test.classification)

			if result.Decision != test.expected {
				t.Fatalf(
					"expected %q, got %q",
					test.expected,
					result.Decision,
				)
			}
		})
	}
}
