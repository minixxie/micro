package micro

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRedoc(t *testing.T) {

	req := httptest.NewRequest("GET", "/docs", nil)
	recorder := httptest.NewRecorder()

	redoc(recorder, req, map[string]string{})

	assert.Equal(t, 200, recorder.Code)
	assert.Equal(t, "text/html; charset=utf-8", recorder.Header().Get("Content-Type"))
	assert.Contains(t, recorder.Body.String(), "<title>API documentation</title>")
}

func TestRedoc2(t *testing.T) {

	req := httptest.NewRequest("GET", "/docs", nil)
	recorder := httptest.NewRecorder()

	RedocOptions.AddSpec("PetStore", "https://rebilly.github.io/ReDoc/swagger.yaml")
	RedocOptions.AddSpec("Instagram", "https://api.apis.guru/v2/specs/instagram.com/1.0.0/swagger.yaml")
	RedocOptions.AddSpec("Google Calendar", "https://api.apis.guru/v2/specs/googleapis.com/calendar/v3/swagger.yaml")

	redoc(recorder, req, map[string]string{})

	assert.Equal(t, 200, recorder.Code)
	assert.Equal(t, "text/html; charset=utf-8", recorder.Header().Get("Content-Type"))
	assert.Contains(t, recorder.Body.String(), "<title>API documentation</title>")
	assert.Contains(t, recorder.Body.String(), "Google Calendar")
}
