package service

import (
	"context"

	"github.com/hugojanuario/devdocs/internal/extraction"
	"github.com/hugojanuario/devdocs/internal/model"
	"github.com/hugojanuario/devdocs/internal/search"
)

// This small deterministic recipe still requires live official evidence. It
// cannot produce a canned example if the documented statement disappears.
func (e *Engine) pythonHello(ctx context.Context, r model.SearchRequest) (model.Answer, error) {
	const printURL = "https://docs.python.org/3/builtins/functions.html"
	const exampleURL = "https://docs.python.org/3/whatsnew/3.5.html"
	printRaw, printFinal, err := e.Fetcher.Fetch(ctx, printURL, 2<<20)
	if err != nil || printFinal != printURL {
		return model.Answer{}, search.ErrUnavailable
	}
	exampleRaw, exampleFinal, err := e.Fetcher.Fetch(ctx, exampleURL, 2<<20)
	if err != nil || exampleFinal != exampleURL {
		return model.Answer{}, search.ErrUnavailable
	}
	summary, err := extraction.PythonPrintSummary(printRaw)
	if err != nil || summary == "" {
		return model.Answer{}, search.ErrUnavailable
	}
	code, err := extraction.PythonHello(exampleRaw)
	if err != nil || code == "" {
		return model.Answer{}, search.ErrUnavailable
	}
	title := "Hello world em Python"
	instruction := "Salve o código em hello.py e execute: python hello.py."
	if r.Language == "en" {
		title = "Hello world in Python"
		instruction = "Save the code in hello.py and run: python hello.py."
	}
	return model.Answer{
		Query: r.Query, Language: r.Language, Title: title, Technology: "python",
		Summary: summary, Description: "", Mode: "deterministic", Status: "completed", Quality: "extracted",
		Examples:     []model.Example{{Language: "python", Code: code, ReferenceID: "source-2", Explanations: []model.Explanation{}}},
		Observations: []string{instruction},
		References: []model.Reference{
			{ID: "source-1", Title: "print — Built-in Functions", URL: printURL + "#print", Domain: "docs.python.org", Official: true, Technology: "python", Excerpt: summary},
			{ID: "source-2", Title: "Python documentation — print example", URL: exampleURL + "#pep-475-retry-system-calls-failing-with-eintr", Domain: "docs.python.org", Official: true, Technology: "python", Excerpt: code},
		},
		Evidence: []model.Evidence{{ReferenceID: "source-1", Field: "summary", Quote: summary}, {ReferenceID: "source-2", Field: "example", Quote: code}},
	}, nil
}
