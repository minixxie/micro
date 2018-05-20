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
	assert.Contains(t, recorder.Body.String(), "<redoc spec-url='/swagger.json'></redoc>")
}
