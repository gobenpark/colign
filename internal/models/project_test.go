package models

import (
	"testing"
	"time"
)

func TestProjectModel(t *testing.T) {
	p := &Project{
		ID:          1,
		Name:        "My App",
		Slug:        "my-app",
		Description: "Test project",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if p.Name != "My App" {
		t.Errorf("expected name 'My App', got '%s'", p.Name)
	}
	if p.Slug != "my-app" {
		t.Errorf("expected slug 'my-app', got '%s'", p.Slug)
	}
}

func TestProjectMemberModel(t *testing.T) {
	pm := &ProjectMember{
		ID:        1,
		ProjectID: 1,
		UserID:    1,
		Role:      RoleOwner,
	}

	if pm.Role != RoleOwner {
		t.Errorf("expected role '%s', got '%s'", RoleOwner, pm.Role)
	}
}

func TestProjectMemberRoles(t *testing.T) {
	roles := []Role{RoleOwner, RoleEditor, RoleViewer}
	expected := []string{"owner", "editor", "viewer"}

	for i, role := range roles {
		if string(role) != expected[i] {
			t.Errorf("expected role '%s', got '%s'", expected[i], role)
		}
	}
}
