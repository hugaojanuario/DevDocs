//go:build integration

package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/hugojanuario/devdocs/internal/api"
	"github.com/hugojanuario/devdocs/internal/bookmarks"
	"github.com/hugojanuario/devdocs/internal/config"
	"github.com/hugojanuario/devdocs/internal/database"
	"github.com/hugojanuario/devdocs/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

type mockedSearch struct{ calls int }

func (e *mockedSearch) Search(context.Context, model.SearchRequest) (model.Answer, error) {
	e.calls++
	return model.Answer{Query: "integration fixture python", Title: "Fixture evidence", Technology: "python", Language: "pt-BR", Summary: "A documented fixture quote.", Examples: []model.Example{}, Observations: []string{}, References: []model.Reference{{ID: "source-1", Title: "Fixture", URL: "https://docs.python.org/3/", Domain: "docs.python.org", Official: true, Excerpt: "A documented fixture quote."}}, Evidence: []model.Evidence{{ReferenceID: "source-1", Field: "summary", Quote: "A documented fixture quote."}}, Mode: "deterministic", Status: "partial", Quality: "limited"}, nil
}
func TestPostgresExplicitPersistenceAndPhysicalDelete(t *testing.T) {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Fatal("TEST_DATABASE_URL required (dedicated *_test database)")
	}
	pc, err := pgxpool.ParseConfig(url)
	if err != nil {
		t.Fatal("invalid test database URL")
	}
	if !strings.HasSuffix(pc.ConnConfig.Database, "_test") {
		t.Fatal("requires a dedicated *_test database")
	}
	admin := pc.ConnConfig.Copy()
	name := admin.Database
	admin.Database = "postgres"
	ctx := context.Background()
	conn, err := pgx.ConnectConfig(ctx, admin)
	if err != nil {
		t.Fatal("test PostgreSQL unavailable")
	}
	var exists bool
	err = conn.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname=$1)", name).Scan(&exists)
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		if _, err = conn.Exec(ctx, "CREATE DATABASE "+pgx.Identifier{name}.Sanitize()); err != nil {
			t.Fatal(err)
		}
	}
	conn.Close(ctx)
	if err = database.Migrate(url, "../db/migrations"); err != nil {
		t.Fatal("test migration failed")
	}
	pool, err := pgxpool.NewWithConfig(ctx, pc)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	engine := &mockedSearch{}
	router := api.New(config.Config{CLIToken: "integration", WebPassword: "integration-password", Origin: "http://localhost:8080"}, bookmarks.New(pool), engine)
	req := func(method, path string, body any) *httptest.ResponseRecorder {
		b, _ := json.Marshal(body)
		r := httptest.NewRequest(method, path, bytes.NewReader(b))
		r.Header.Set("Authorization", "Bearer integration")
		r.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, r)
		return w
	}
	count := func() int {
		var n int
		if err = pool.QueryRow(ctx, "SELECT count(*) FROM saved_queries").Scan(&n); err != nil {
			t.Fatal(err)
		}
		return n
	}
	before := count()
	w := req("POST", "/api/v1/search", model.SearchRequest{Query: "integration fixture python"})
	if w.Code != http.StatusOK || count() != before {
		t.Fatal("search caused persistence", w.Code)
	}
	var answer model.Answer
	_ = json.Unmarshal(w.Body.Bytes(), &answer)
	w = req("POST", "/api/v1/saved", map[string]any{"response": answer})
	if w.Code != 201 || count() != before+1 || engine.calls != 1 {
		t.Fatal("explicit save failed", w.Code)
	}
	var saved bookmarks.Saved
	_ = json.Unmarshal(w.Body.Bytes(), &saved)
	t.Log("SQL count unchanged after search; exactly one new row after explicit save")
	w = req("GET", "/api/v1/saved?page=1&page_size=1&technology=python&q=fixture", nil)
	var page bookmarks.Page
	_ = json.Unmarshal(w.Body.Bytes(), &page)
	if w.Code != 200 || len(page.Items) != 1 || page.Total < 1 {
		t.Fatal("pagination/filter failed")
	}
	w = req("PATCH", "/api/v1/saved/"+saved.ID, map[string]string{"title": "Renamed fixture"})
	if w.Code != 200 {
		t.Fatal("rename", w.Code)
	}
	w = req("GET", "/api/v1/saved/"+saved.ID, nil)
	var got bookmarks.Saved
	_ = json.Unmarshal(w.Body.Bytes(), &got)
	if w.Code != 200 || got.Title != "Renamed fixture" {
		t.Fatal("get/rename failed")
	}
	w = req("DELETE", "/api/v1/saved/"+saved.ID, nil)
	if w.Code != 204 {
		t.Fatal("delete failed", w.Code)
	}
	var n int
	if err = pool.QueryRow(ctx, "SELECT count(*) FROM saved_queries WHERE id=$1", saved.ID).Scan(&n); err != nil || n != 0 {
		t.Fatal("row physically retained")
	}
	if count() != before {
		t.Fatal("row count mismatch")
	}
	t.Log("SELECT count(*) WHERE id = deleted UUID: 0; physical DELETE proven")
	if w = req("GET", "/api/v1/saved/"+saved.ID, nil); w.Code != 404 {
		t.Fatal("deleted query returned", w.Code)
	}
}
