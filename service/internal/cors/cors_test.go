package cors

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCors(t *testing.T) {
	req, _ := http.NewRequest("GET", "/health-check", nil)
	req.Header.Add("Origin", "1.2.3.4")

	blat := AllowCors(http.FileServer(http.Dir(".")))

	rr := httptest.NewRecorder()

	blat.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code, "HTTP 404 on CORS")
	assert.Equal(t, "1.2.3.4", rr.Header().Get("Access-Control-Allow-Origin"), "CORS Header set")
}
