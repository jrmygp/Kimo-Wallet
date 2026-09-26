package domain

import "testing"

func TestNewCreateWidgetInput(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr error
	}{
		{name: "valid name", input: "My Widget"},
		{name: "trims surrounding whitespace", input: "  My Widget  "},
		{name: "empty after trim", input: "   ", wantErr: ErrInvalidName},
		{name: "too long", input: string(make([]byte, maxNameLength+1)), wantErr: ErrInvalidName},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewCreateWidgetInput(tt.input)
			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Fatalf("NewCreateWidgetInput(%q) error = %v, want %v", tt.input, err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("NewCreateWidgetInput(%q) error = %v, want nil", tt.input, err)
			}
		})
	}
}

func TestNewGetWidgetByIDInput(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr error
	}{
		{name: "valid uuid", input: "11111111-1111-4111-8111-111111111111"},
		{name: "empty", input: "", wantErr: ErrInvalidID},
		{name: "malformed", input: "not-a-uuid", wantErr: ErrInvalidID},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewGetWidgetByIDInput(tt.input)
			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Fatalf("NewGetWidgetByIDInput(%q) error = %v, want %v", tt.input, err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("NewGetWidgetByIDInput(%q) error = %v, want nil", tt.input, err)
			}
		})
	}
}
