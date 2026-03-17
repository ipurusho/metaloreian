package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWriteJSON(t *testing.T) {
	tests := []struct {
		name       string
		status     int
		data       any
		wantStatus int
	}{
		{"ok with map", http.StatusOK, map[string]string{"hello": "world"}, 200},
		{"created with struct", http.StatusCreated, struct {
			ID int `json:"id"`
		}{42}, 201},
		{"empty slice", http.StatusOK, []string{}, 200},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			writeJSON(w, tt.status, tt.data)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}
			if ct := w.Header().Get("Content-Type"); ct != "application/json" {
				t.Errorf("Content-Type = %q, want application/json", ct)
			}
		})
	}
}

func TestWriteError(t *testing.T) {
	w := httptest.NewRecorder()
	writeError(w, http.StatusBadRequest, "something broke")

	if w.Code != 400 {
		t.Errorf("status = %d, want 400", w.Code)
	}

	var resp errorResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Error != "something broke" {
		t.Errorf("error = %q, want %q", resp.Error, "something broke")
	}
}
