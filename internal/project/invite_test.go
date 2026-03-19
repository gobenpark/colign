package project

import (
	"testing"

	"github.com/gobenpark/colign/internal/models"
)

func TestValidateInviteRole(t *testing.T) {
	tests := []struct {
		role    string
		valid   bool
	}{
		{"owner", true},
		{"editor", true},
		{"viewer", true},
		{"admin", false},
		{"", false},
	}

	for _, tt := range tests {
		got := ValidateRole(tt.role)
		if got != tt.valid {
			t.Errorf("ValidateRole(%q) = %v, want %v", tt.role, got, tt.valid)
		}
	}
}

func TestParseRole(t *testing.T) {
	role, err := ParseRole("editor")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if role != models.RoleEditor {
		t.Errorf("expected RoleEditor, got %s", role)
	}

	_, err = ParseRole("invalid")
	if err == nil {
		t.Error("expected error for invalid role")
	}
}
