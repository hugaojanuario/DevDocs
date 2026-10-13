package commands

import (
	"bytes"
	"encoding/json"
	"github.com/hugojanuario/devdocs/internal/model"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
)

func TestRemoteSearchAndExplicitSave(t *testing.T) {
	searches, saves := 0, 0
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer private-test-token" {
			t.Error("missing remote auth")
		}
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/v1/search":
			searches++
			json.NewEncoder(w).Encode(model.Answer{Title: "Remote answer", Examples: []model.Example{{Language: "python", Code: "if True:\n    pass"}}})
		case "/api/v1/saved":
			saves++
			w.WriteHeader(201)
			w.Write([]byte(`{"id":"123e4567-e89b-12d3-a456-426614174000"}`))
		default:
			t.Error("unexpected route")
		}
	}))
	defer s.Close()
	t.Setenv("DEVDOCS_CONFIG", filepath.Join(t.TempDir(), "config.json"))
	t.Setenv("DEVDOCS_SERVER", s.URL)
	t.Setenv("DEVDOCS_TOKEN", "private-test-token")
	for _, save := range []bool{false, true} {
		root := New()
		var out, errout bytes.Buffer
		root.SetOut(&out)
		root.SetErr(&errout)
		args := []string{"search", "test remote", "--format", "json"}
		if save {
			args = append(args, "--save")
		}
		root.SetArgs(args)
		if e := root.Execute(); e != nil {
			t.Fatal(e)
		}
		var a model.Answer
		if e := json.Unmarshal(out.Bytes(), &a); e != nil || a.Title != "Remote answer" {
			t.Fatal(out.String(), e)
		}
	}
	if searches != 2 || saves != 1 {
		t.Fatal(searches, saves)
	}
}
