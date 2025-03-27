package secretary

import (
	"encoding/json"
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
	mux := http.NewServeMux()
	router := s.setupRouter(mux)

	req := httptest.NewRequest(http.MethodGet, "/getalltree", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req) // Call handler directly

	resp := rec.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status OK; got %v", resp.Status)
	}
}

// Test /getbtree with query params
func TestGetBTreeHandler(t *testing.T) {
	s, err := New()
	if err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	router := s.setupRouter(mux)

	users, err := s.Tree("users")
	if err != nil {
		t.Fatal(err)
	}

	var keySeq uint64 = 0
	key := []byte(utils.GenerateSeqRandomString(&keySeq, 16, 5, 4))
	err = users.Set(key, key)
	if err != nil {
		t.Fatalf("Insert failed: %s", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/gettree/123", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	resp := rec.Result()
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected status StatusNotFound; got %v", resp.Status)
	}

	req = httptest.NewRequest(http.MethodGet, "/gettree/users", nil)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	resp = rec.Result()
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status OK; got %v", resp.Status)
	}
}

// Test /insert with POST data
func TestSetHandler(t *testing.T) {
	s, err := New()
	if err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	router := s.setupRouter(mux)

	body := `{"value": "123"}`
	req := httptest.NewRequest(http.MethodPost, "/set/users", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	resp := rec.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status OK; got %v", resp.Status)
	}
}

// Test /insert with POST data
func TestNewTreeHandler(t *testing.T) {
	s, err := New()
	if err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	router := s.setupRouter(mux)

	tests := []struct {
		tree NewTreeRequest
		pass bool
	}{
		{
			tree: NewTreeRequest{
				CollectionName:      "hello",
				Order:               10,
				NumLevel:            32,
				BatchBaseSize:       1024,
				Increment:           130,
				CompactionBatchSize: 1000,
			},
			pass: true,
		},
		{
			tree: NewTreeRequest{
				CollectionName:      "hello",
				Order:               2,
				NumLevel:            32,
				BatchBaseSize:       104,
				Increment:           105,
				CompactionBatchSize: 1000,
			},
			pass: false,
		},
	}

	for _, test := range tests {
		j, err := json.Marshal(test.tree)
		if err != nil {
			t.Fatal(err)
		}
		req := httptest.NewRequest(http.MethodPost, "/newtree", strings.NewReader(string(j)))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		resp := rec.Result()
		defer resp.Body.Close()

		if test.pass != (resp.StatusCode == http.StatusOK) {
			t.Fatalf("expected status OK; got %v", resp.Status)
		}
	}
}

func TestGetHandler(t *testing.T) {
	s, err := New()
	if err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	router := s.setupRouter(mux)

	u, err := s.Tree("users")
	if err != nil {
		t.Fatal(err)
	}

	var keySeq uint64 = 0
	key := []byte(utils.GenerateSeqRandomString(&keySeq, 16, 5, 4))
	err = u.Set(key, key)
	if err != nil {
		t.Fatalf("Insert failed: %s", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/get/users/123", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	resp := rec.Result()
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected status StatusNoContent; got %v", resp.Status)
	}

	req = httptest.NewRequest(http.MethodGet, "/get/users/"+string(key), nil)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	resp = rec.Result()
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status StatusOK; got %v", resp.Status)
	}
}
