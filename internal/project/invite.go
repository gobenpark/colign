package project

import (
	"fmt"

	"github.com/gobenpark/colign/internal/models"
)

var validRoles = map[string]models.Role{
	"owner":  models.RoleOwner,
	"editor": models.RoleEditor,
	"viewer": models.RoleViewer,
}

func ValidateRole(role string) bool {
	_, ok := validRoles[role]
	return ok
}

func ParseRole(role string) (models.Role, error) {
	r, ok := validRoles[role]
	if !ok {
		return "", fmt.Errorf("invalid role: %s", role)
	}
	return r, nil
}
