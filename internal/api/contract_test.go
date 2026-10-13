package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers/legacy"
	"github.com/hugojanuario/devdocs/internal/model"
	"net/http/httptest"
	"testing"
)

func TestResponsesMatchPublishedContract(t *testing.T) {
	ctx := context.Background()
	doc, err := openapi3.NewLoader().LoadFromFile("../../api/openapi.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if err = doc.Validate(ctx); err != nil {
		t.Fatal(err)
	}
	doc.Servers = openapi3.Servers{&openapi3.Server{URL: "http://localhost/api/v1"}}
	routes, err := legacy.NewRouter(doc)
	if err != nil {
		t.Fatal(err)
	}
	api, _, _ := testRouter()
	cases := []struct {
		method, path string
		body         any
		auth         bool
	}{{"GET", "/api/v1/health", nil, false}, {"GET", "/api/v1/ready", nil, false}, {"GET", "/api/v1/technologies", nil, true}, {"GET", "/api/v1/sources", nil, true}, {"GET", "/api/v1/saved", nil, true}, {"GET", "/api/v1/saved", nil, false}, {"POST", "/api/v1/search", model.SearchRequest{Query: "test python"}, true}, {"POST", "/api/v1/saved", map[string]any{"response": testAnswer()}, true}}
	cases = append(cases, struct {
		method, path string
		body         any
		auth         bool
	}{"POST", "/api/v1/saved", map[string]any{"response": normalizedAnswer()}, true})
	cases = append(cases, struct {
		method, path string
		body         any
		auth         bool
	}{"GET", "/api/v1/capabilities", nil, true})
	for _, c := range cases {
		t.Run(fmt.Sprintf("%s %s auth=%t", c.method, c.path, c.auth), func(t *testing.T) {
			b, _ := json.Marshal(c.body)
			req := httptest.NewRequest(c.method, "http://localhost"+c.path, bytes.NewReader(b))
			req.Header.Set("Content-Type", "application/json")
			if c.auth {
				req.Header.Set("Authorization", "Bearer token")
			}
			rec := httptest.NewRecorder()
			api.ServeHTTP(rec, req)
			route, params, err := routes.FindRoute(req)
			if err != nil {
				t.Fatal(err)
			}
			input := &openapi3filter.ResponseValidationInput{RequestValidationInput: &openapi3filter.RequestValidationInput{Request: req, PathParams: params, Route: route}, Status: rec.Code, Header: rec.Header(), Options: &openapi3filter.Options{IncludeResponseStatus: true}}
			input.SetBodyBytes(rec.Body.Bytes())
			if err = openapi3filter.ValidateResponse(ctx, input); err != nil {
				t.Fatal(err)
			}
		})
	}
}
