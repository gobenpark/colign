package oauth

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRegisterHandler_GETMethodNotAllowed(t *testing.T) {
	handler := &RegisterHandler{}

	req := httptest.NewRequest(http.MethodGet, "/oauth/register", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
}

func TestRegisterHandler_InvalidJSON(t *testing.T) {
	handler := &RegisterHandler{}

	req := httptest.NewRequest(http.MethodPost, "/oauth/register",
		strings.NewReader("this is not json"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}
