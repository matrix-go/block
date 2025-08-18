package api

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestServer_SetRouter(t *testing.T) {
	server := NewServer(ServerConfig{}, nil)
	router := server.SetRouter()

	recorder := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/test", nil)
	require.NoError(t, err)
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "{\"msg\":\"success\"}", recorder.Body.String())
}
