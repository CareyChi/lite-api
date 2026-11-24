package httpserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := NewRouter()

	tests := []struct {
		name  string
		path  string
		field string
		want  string
	}{
		{name: "liveness", path: "/healthz", field: "status", want: "ok"},
		{name: "system info", path: "/api/v1/system/info", field: "status", want: "scaffold"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			router.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Fatalf("status = %d, want 200", w.Code)
			}
			var body map[string]string
			if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
				t.Fatal(err)
			}
			if body[tt.field] != tt.want {
				t.Fatalf("%s = %q, want %q", tt.field, body[tt.field], tt.want)
			}
		})
	}
}
