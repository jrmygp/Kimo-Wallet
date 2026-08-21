package domain

import (
	"errors"
	"strings"
	"testing"
)

func TestNewRegisterInput(t *testing.T) {
	longestValidName := strings.Repeat("a", maxFullNameLength)

	tests := []struct {
		name        string
		phoneNumber string
		fullName    string
		wantErr     error
	}{
		{
			name:        "valid Indonesian E.164 number",
			phoneNumber: "+6281234567890",
			fullName:    "Jane Doe",
		},
		{
			name:        "trims surrounding whitespace",
			phoneNumber: "  +6281234567890  ",
			fullName:    "  Jane Doe  ",
		},
		{
			name:        "full name at max length boundary",
			phoneNumber: "+6281234567890",
			fullName:    longestValidName,
		},
		{
			name:        "missing leading plus",
			phoneNumber: "6281234567890",
			fullName:    "Jane Doe",
			wantErr:     ErrInvalidPhoneNumber,
		},
		{
			name:        "leading zero after plus",
			phoneNumber: "+0281234567890",
			fullName:    "Jane Doe",
			wantErr:     ErrInvalidPhoneNumber,
		},
		{
			name:        "too short to be E.164",
			phoneNumber: "+62812",
			fullName:    "Jane Doe",
			wantErr:     ErrInvalidPhoneNumber,
		},
		{
			name:        "contains non-digit characters",
			phoneNumber: "+6281-234-5678",
			fullName:    "Jane Doe",
			wantErr:     ErrInvalidPhoneNumber,
		},
		{
			name:        "empty phone number",
			phoneNumber: "",
			fullName:    "Jane Doe",
			wantErr:     ErrInvalidPhoneNumber,
		},
		{
			name:        "empty full name",
			phoneNumber: "+6281234567890",
			fullName:    "",
			wantErr:     ErrInvalidFullName,
		},
		{
			name:        "whitespace-only full name",
			phoneNumber: "+6281234567890",
			fullName:    "   ",
			wantErr:     ErrInvalidFullName,
		},
		{
			name:        "full name over max length",
			phoneNumber: "+6281234567890",
			fullName:    longestValidName + "a",
			wantErr:     ErrInvalidFullName,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewRegisterInput(tt.phoneNumber, tt.fullName)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("NewRegisterInput(%q, %q) error = %v, want %v", tt.phoneNumber, tt.fullName, err, tt.wantErr)
			}
		})
	}
}
