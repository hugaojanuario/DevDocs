package api

import (
	"testing"

	"github.com/hugojanuario/devdocs/internal/model"
)

func normalizedAnswer() model.Answer {
	a := testAnswer()
	a.Examples = []model.Example{{Language: "python", Code: `pd.read_csv("data.csv")`, SourceCode: ">>> pd.read_csv(\"data.csv\")\n   Name Value\n0  foo 1", Setup: "import pandas as pd", SetupReferenceID: "source-2", Output: "   Name Value\n0  foo 1", Explanations: []model.Explanation{}, ReferenceID: "source-1"}}
	a.References = append(a.References, model.Reference{ID: "source-2", Title: "Pandas guide", URL: "https://pandas.pydata.org/docs/user_guide/io.html", Domain: "pandas.pydata.org", Official: true, Excerpt: "import pandas as pd"})
	a.Evidence = append(a.Evidence, model.Evidence{ReferenceID: "source-1", Field: "example", Quote: a.Examples[0].SourceCode}, model.Evidence{ReferenceID: "source-2", Field: "setup", Quote: a.Examples[0].Setup})
	return a
}

func TestNormalizedEvidenceCannotBeChangedWhenSaving(t *testing.T) {
	a := normalizedAnswer()
	if !validAnswer(a) {
		t.Fatal("normalized answer rejected")
	}
	for _, mutation := range []func(*model.Answer){
		func(a *model.Answer) { a.Examples[0].Code = "invented()" },
		func(a *model.Answer) { a.Examples[0].Output = "invented output" },
		func(a *model.Answer) { a.Examples[0].Setup = "import invented" },
		func(a *model.Answer) { a.Examples[0].SetupReferenceID = "source-9" },
	} {
		a = normalizedAnswer()
		mutation(&a)
		if validAnswer(a) {
			t.Fatal("unsupported transformation accepted", a)
		}
	}
	r, _, _ := testRouter()
	if w := request(r, "POST", "/api/v1/saved", map[string]any{"response": normalizedAnswer()}, bearer); w.Code != 201 {
		t.Fatal(w.Code, w.Body.String())
	}
}

func TestAIParaphraseKeepsLiteralSupportingEvidence(t *testing.T) {
	a := normalizedAnswer()
	a.Mode = "ai_assisted"
	a.Summary = "Resumo traduzido da evidência."
	a.Description = "Descrição sintetizada do código."
	a.Evidence = append(a.Evidence, model.Evidence{ReferenceID: "source-1", Field: "description", Quote: a.Examples[0].SourceCode})
	if !validAnswer(a) {
		t.Fatal("AI prose with literal supporting evidence rejected")
	}
	a.Evidence[len(a.Evidence)-1].Quote = "unsupported"
	if validAnswer(a) {
		t.Fatal("unsupported synthesis evidence accepted")
	}
}
