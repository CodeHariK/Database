package secretary

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/codeharik/secretary/utils"
)

// Test /getallbtree
func TestGetAllBTreeHandler(t *testing.T) {
	s, err := New()
	if err != nil {
		t.Fatal(err)
	}
	router := s.setupRouter()

	req := httptest.NewRequest(http.MethodGet, "/getallbtree", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req) // Call handler directly

	resp := rec.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status OK; got %v", resp.Status)
	}
}

// Test /getbtree with query params
func TestGetBTreeHandler(t *testing.T) {
	s, err := New()
	if err != nil {
		t.Fatal(err)
	}

	router := s.setupRouter()

	u, _ := s.Tree("users")
	key := []byte(utils.GenerateSeqRandomString(16, 4))
	err = u.Insert(key, key)
	if err != nil {
		t.Errorf("Insert failed: %s", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/getbtree?table=123", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	resp := rec.Result()
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected status StatusNotFound; got %v", resp.Status)
	}

	req = httptest.NewRequest(http.MethodGet, "/getbtree?table=users", nil)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	resp = rec.Result()
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status OK; got %v", resp.Status)
	}
}

// Test /insert with POST data
func TestInsertHandler(t *testing.T) {
	s, err := New()
	if err != nil {
		t.Fatal(err)
	}
	router := s.setupRouter()

	body := `{"value": "123"}`
	req := httptest.NewRequest(http.MethodPost, "/insert?table=users", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	resp := rec.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status OK; got %v", resp.Status)
	}
}
