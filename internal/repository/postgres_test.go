package repository

import (
	"github.com/stretchr/testify/assert"
	"regexp"
	"testing"
)

const GENERATED_STRING_MAX_LENGTH = 48

func TestGenerateUUIDCode(t *testing.T) {
	base62Regex := regexp.MustCompile(`^[0-9a-zA-Z]*$`)

	tests := []struct {
		name        string
		length      int
		wantLen     int
		expectError bool
	}{
		{"Normal length", 8, 8, false},
		{"Zero length", 0, 0, false},
		{"Large length", 100, GENERATED_STRING_MAX_LENGTH, false},
		{"Negative length", -5, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil && !tt.expectError {
					t.Errorf("generateUUIDCode() panicked unexpectedly: %v", r)
				}
			}()

			result, err := generateUUIDCode(tt.length)
			if tt.length < 0 {
				assert.ErrorContains(t, err, "invalid length")
			}

			if len(result) != tt.wantLen {
				t.Errorf("Length mismatch: got %d, want %d", len(result), tt.wantLen)
			}

			if tt.length > 0 && !base62Regex.MatchString(result) {
				t.Errorf("Invalid characters in code: %s", result)
			}

			// Проверка уникальности генерации
			if tt.length > 0 {
				another, _ := generateUUIDCode(tt.length)
				if result == another {
					t.Error("Generated identical codes, uniqueness not guaranteed")
				}
			}
		})
	}
}

func FuzzGenerateUUIDCode(f *testing.F) {
	f.Fuzz(func(t *testing.T, length int) {
		res, err := generateUUIDCode(length)

		if length < 0 && !assert.ErrorContains(t, err, "invalid length") {
			t.Errorf("invalid input length is not covered, length = %v", length)
		}

		if length > 0 && err != nil {
			t.Errorf("generated value = %q, err = %v", res, err)
		}

		if length > 0 && length <= GENERATED_STRING_MAX_LENGTH && len(res) != length {
			t.Errorf("length missmatch, want = %v, got = %v", length, len(res))
		}

		if length > GENERATED_STRING_MAX_LENGTH && len(res) != GENERATED_STRING_MAX_LENGTH {
			t.Errorf("generated string has length over max, string = %q, length = %v", res, err)
		}
	})
}
