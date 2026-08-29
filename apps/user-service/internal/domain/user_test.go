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

func TestNewGetUserByKimoIDInput(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		want    string
		wantErr error
	}{
		{
			name: "valid KimoID",
			id:   "ABCDEF123456",
			want: "ABCDEF123456",
		},
		{
			name: "lowercase is upper-cased, not rejected",
			id:   "abcdef123456",
			want: "ABCDEF123456",
		},
		{
			name: "trims surrounding whitespace",
			id:   "  ABCDEF123456  ",
			want: "ABCDEF123456",
		},
		{
			name:    "empty id",
			id:      "",
			wantErr: ErrInvalidKimoID,
		},
		{
			name:    "a phone number, not a KimoID",
			id:      "+6281234567890",
			wantErr: ErrInvalidKimoID,
		},
		{
			name:    "a UUID, not a KimoID",
			id:      "11111111-1111-4111-8111-111111111111",
			wantErr: ErrInvalidKimoID,
		},
		{
			name:    "one character short",
			id:      "ABCDEF12345",
			wantErr: ErrInvalidKimoID,
		},
		{
			name:    "one character too long",
			id:      "ABCDEF1234567",
			wantErr: ErrInvalidKimoID,
		},
		{
			name:    "contains a non-alphanumeric character",
			id:      "ABCDEF-23456",
			wantErr: ErrInvalidKimoID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewGetUserByKimoIDInput(tt.id)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("NewGetUserByKimoIDInput(%q) error = %v, want %v", tt.id, err, tt.wantErr)
			}
			if err == nil && got != tt.want {
				t.Fatalf("NewGetUserByKimoIDInput(%q) = %q, want %q", tt.id, got, tt.want)
			}
		})
	}
}
