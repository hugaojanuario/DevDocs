package ai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/hugojanuario/devdocs/internal/model"
)

func evidence() model.Answer {
	return model.Answer{Query: "Como ler CSV com Pandas?", Language: "pt-BR", Summary: "Read CSV into a DataFrame.", Mode: "deterministic", Status: "completed", Quality: "extracted", Examples: []model.Example{{Language: "python", Code: `pd.read_csv("data.csv")`, SourceCode: `>>> pd.read_csv("data.csv")`, Setup: "import pandas as pd", ReferenceID: "source-1", Explanations: []model.Explanation{}}}, References: []model.Reference{{ID: "source-1", Excerpt: "Read CSV into a DataFrame."}}, Evidence: []model.Evidence{{ReferenceID: "source-1", Field: "summary", Quote: "Read CSV into a DataFrame."}}, Observations: []string{"A versão solicitada não foi confirmada."}}
}
func TestGroqUsesBoundedStructuredEvidenceAndPreservesCode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.Header.Get("Authorization") != "Bearer private-test-key" {
			t.Error("bad request auth")
		}
		var req map[string]any
		if json.NewDecoder(r.Body).Decode(&req) != nil {
			t.Error("invalid request")
		}
		if req["max_completion_tokens"] != float64(2048) || req["model"] != "openai/gpt-oss-20b" {
			t.Error("missing limits")
		}
		body, _ := json.Marshal(req)
		if strings.Contains(string(body), "private-test-key") {
			t.Error("credential sent inside evidence")
		}
		rf := req["response_format"].(map[string]any)
		if rf["type"] != "json_schema" || rf["json_schema"].(map[string]any)["strict"] != true {
			t.Error("missing strict contract")
		}
		json.NewEncoder(w).Encode(map[string]any{"choices": []any{map[string]any{"finish_reason": "stop", "message": map[string]string{"content": `{"title":"Ler CSV com Pandas","summary":"Leia o CSV em um DataFrame.","description":"A função carrega o arquivo indicado.","explanations":[{"index":0,"snippet":"pd.read_csv","description":"Lê o arquivo CSV."}]}`}}}})
	}))
	defer server.Close()
	g := NewGroq("private-test-key", "openai/gpt-oss-20b")
	g.endpoint = server.URL
	original := evidence()
	a, err := g.Answer(context.Background(), original)
	if err != nil || a.Mode != "ai_assisted" || a.Summary == original.Summary || a.Examples[0].Code != original.Examples[0].Code || !reflect.DeepEqual(a.References, original.References) || a.Examples[0].SourceCode != original.Examples[0].SourceCode {
		t.Fatal(a, err)
	}
	if len(a.Examples[0].Explanations) != 1 || len(a.Observations) != 2 || len(a.Evidence) != 2 {
		t.Fatal("lost evidence or version note", a)
	}
}
func TestGroqRejectsUnsupportedOutputAndNeverLeaksProviderBody(t *testing.T) {
	tests := []struct {
		status          int
		content, reason string
	}{
		{429, "private-test-key", "stop"},
		{200, `{"title":"Test","summary":"Text","description":"","explanations":[],"code":"invented()"}`, "stop"},
		{200, `{"title":"Test","summary":"Text","description":"","explanations":[{"index":0,"snippet":"invented()","description":"Wrong"}]}`, "stop"},
		{200, `{"title":"Test","summary":"Text","description":"","explanations":[]}`, "length"},
		{200, "not JSON", "stop"},
	}
	for _, tt := range tests {
		t.Run(tt.content, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.status)
				json.NewEncoder(w).Encode(map[string]any{"choices": []any{map[string]any{"finish_reason": tt.reason, "message": map[string]string{"content": tt.content}}}})
			}))
			defer server.Close()
			g := NewGroq("private-test-key", "openai/gpt-oss-20b")
			g.endpoint = server.URL
			_, err := g.Answer(context.Background(), evidence())
			if err != ErrUnavailable {
				t.Fatal("invalid output accepted", err)
			}
		})
	}
}
