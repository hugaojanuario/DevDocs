package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/hugojanuario/devdocs/internal/bookmarks"
	"github.com/hugojanuario/devdocs/internal/config"
	"github.com/hugojanuario/devdocs/internal/model"
	"github.com/hugojanuario/devdocs/internal/search"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type fakeStore struct{ inserts, deletes int }

func (s *fakeStore) Save(_ context.Context, a model.Answer, title string) (bookmarks.Saved, error) {
	s.inserts++
	return bookmarks.Saved{ID: "123e4567-e89b-12d3-a456-426614174000", Title: title, Response: a}, nil
}
func (s *fakeStore) Get(context.Context, string) (bookmarks.Saved, error) {
	return bookmarks.Saved{}, nil
}
func (s *fakeStore) List(_ context.Context, p, n int, q, t, order string) (bookmarks.Page, error) {
	return bookmarks.Page{Items: []bookmarks.Summary{}, Page: p, PageSize: n}, nil
}
func (s *fakeStore) Rename(context.Context, string, string) (bookmarks.Saved, error) {
	return bookmarks.Saved{}, nil
}
func (s *fakeStore) Delete(context.Context, string) (bool, error) { s.deletes++; return true, nil }
func (s *fakeStore) Ready(context.Context) error                  { return nil }

type fakeEngine struct {
	err   error
	calls int
}

func (e *fakeEngine) Search(context.Context, model.SearchRequest) (model.Answer, error) {
	e.calls++
	return testAnswer(), e.err
}
func testAnswer() model.Answer {
	return model.Answer{Query: "test python", Title: "Test", Technology: "python", Language: "pt-BR", Mode: "deterministic", Status: "completed", Quality: "extracted", Summary: "Evidence", References: []model.Reference{{ID: "source-1", Title: "Python", URL: "https://docs.python.org/3/", Domain: "docs.python.org", Official: true, Excerpt: "Evidence"}}, Examples: []model.Example{}, Observations: []string{}, Evidence: []model.Evidence{{ReferenceID: "source-1", Field: "summary", Quote: "Evidence"}}}
}
func testRouter() (http.Handler, *fakeStore, *fakeEngine) {
	s := &fakeStore{}
	e := &fakeEngine{}
	r := New(config.Config{CLIToken: "token", WebPassword: "long-test-password", Origin: "http://localhost:8080"}, s, e)
	return r, s, e
}
func request(r http.Handler, method, path string, body any, headers map[string]string) *httptest.ResponseRecorder {
	var b []byte
	if body != nil {
		b, _ = json.Marshal(body)
	}
	req := httptest.NewRequest(method, path, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

var bearer = map[string]string{"Authorization": "Bearer token"}

func TestExplicitPersistence(t *testing.T) {
	r, s, e := testRouter()
	w := request(r, "POST", "/api/v1/search", model.SearchRequest{Query: "test python"}, bearer)
	if w.Code != 200 || s.inserts != 0 {
		t.Fatalf("search status %d inserts %d", w.Code, s.inserts)
	}
	w = request(r, "POST", "/api/v1/saved", map[string]any{"response": testAnswer()}, bearer)
	if w.Code != 201 || s.inserts != 1 || e.calls != 1 {
		t.Fatalf("save status %d inserts %d calls %d", w.Code, s.inserts, e.calls)
	}
	w = request(r, "DELETE", "/api/v1/saved/123e4567-e89b-12d3-a456-426614174000", nil, bearer)
	if w.Code != 204 || s.deletes != 1 {
		t.Fatal(w.Code)
	}
}
func TestAuthenticationAndCSRF(t *testing.T) {
	r, _, _ := testRouter()
	if w := request(r, "GET", "/api/v1/saved", nil, nil); w.Code != 401 {
		t.Fatal(w.Code)
	}
	login := request(r, "POST", "/api/v1/auth/session", map[string]string{"password": "long-test-password"}, map[string]string{"Origin": "http://localhost:8080"})
	if login.Code != 200 {
		t.Fatal(login.Code)
	}
	cookie := login.Result().Cookies()[0]
	if !cookie.HttpOnly || cookie.SameSite != http.SameSiteStrictMode {
		t.Fatal("unsafe cookie")
	}
	var session map[string]string
	_ = json.Unmarshal(login.Body.Bytes(), &session)
	h := map[string]string{"Cookie": cookie.String(), "Origin": "http://localhost:8080"}
	if w := request(r, "POST", "/api/v1/search", model.SearchRequest{Query: "test python"}, h); w.Code != 403 {
		t.Fatal(w.Code)
	}
	h["X-CSRF-Token"] = session["csrf_token"]
	if w := request(r, "POST", "/api/v1/search", model.SearchRequest{Query: "test python"}, h); w.Code != 200 {
		t.Fatal(w.Code)
	}
	w := request(r, "DELETE", "/api/v1/auth/session", nil, h)
	if w.Code != 204 {
		t.Fatal(w.Code)
	}
	if w = request(r, "GET", "/api/v1/saved", nil, h); w.Code != 401 {
		t.Fatal(w.Code)
	}
	if w = request(r, "GET", "/api/v1/sources", nil, map[string]string{"Authorization": "Bearer wrong", "Origin": "https://evil.example"}); w.Code != 403 {
		t.Fatal(w.Code)
	}
}
func TestValidationAndProviderErrors(t *testing.T) {
	r, _, e := testRouter()
	for _, path := range []string{"/api/v1/saved?page=0", "/api/v1/saved?page_size=101", "/api/v1/saved/invalid"} {
		if w := request(r, "GET", path, nil, bearer); w.Code != 400 {
			t.Fatal(path, w.Code)
		}
	}
	for err, status := range map[error]int{search.ErrProvider: 502, search.ErrUnavailable: 503, errors.New("unexpected"): 503} {
		e.err = err
		if w := request(r, "POST", "/api/v1/search", model.SearchRequest{Query: "test python"}, bearer); w.Code != status {
			t.Fatal(w.Code)
		}
	}
	if w := request(r, "POST", "/api/v1/search", map[string]any{"query": "test", "extra": true}, bearer); w.Code != 400 {
		t.Fatal(w.Code)
	}
	a := testAnswer()
	a.Evidence[0].Quote = "fabricated"
	if w := request(r, "POST", "/api/v1/saved", map[string]any{"response": a}, bearer); w.Code != 400 {
		t.Fatal(w.Code)
	}
}
func TestRateLimit(t *testing.T) {
	r, _, _ := testRouter()
	for i := 0; i < 21; i++ {
		w := request(r, "POST", "/api/v1/search", model.SearchRequest{Query: "test python"}, bearer)
		if i == 20 && w.Code != 429 {
			t.Fatal(w.Code)
		}
	}
}

func TestPayloadLimitAndOfficialClaim(t *testing.T) {
	r, _, _ := testRouter()
	req := httptest.NewRequest("POST", "/api/v1/search", strings.NewReader(`{"query":"test python"}`+strings.Repeat(" ", 1<<20)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer token")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != 413 {
		t.Fatal("payload limit", w.Code)
	}
	a := testAnswer()
	a.References[0].URL = "https://complementary.example/tutorial"
	a.References[0].Domain = "complementary.example"
	if validAnswer(a) {
		t.Fatal("forged official claim accepted")
	}
}
