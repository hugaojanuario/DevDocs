package output

import (
	"bytes"
	"github.com/hugojanuario/devdocs/internal/model"
	"strings"
	"testing"
)

func TestFormatsPreserveCode(t *testing.T) {
	a := model.Answer{Title: "Test", Examples: []model.Example{{Language: "python", Code: "if True:\n    print('```')"}}}
	for _, f := range []string{"text", "markdown", "json"} {
		var b bytes.Buffer
		if e := Render(&b, a, f, false); e != nil {
			t.Fatal(e)
		}
		if f != "json" && !strings.Contains(b.String(), a.Examples[0].Code) {
			t.Fatal("indent lost")
		}
		if f == "markdown" && !strings.Contains(b.String(), "````python") {
			t.Fatal("unsafe fence")
		}
		if strings.Contains(b.String(), "\033") {
			t.Fatal("redirected output colors")
		}
	}
	if safe("\033[2Jtest") != "[2Jtest" {
		t.Fatal("escape allowed")
	}
}

func TestContextAndOutputStaySeparate(t *testing.T) {
	a := model.Answer{Examples: []model.Example{{Language: "python", Setup: "import pandas as pd", Code: `pd.read_csv("data.csv")`, Output: "Name Value\nfoo 1"}}}
	var b bytes.Buffer
	if err := Render(&b, a, "markdown", false); err != nil {
		t.Fatal(err)
	}
	text := b.String()
	if !strings.Contains(text, "```python\nimport pandas as pd\n\npd.read_csv(\"data.csv\")\n```") || !strings.Contains(text, "## Saída documentada\n\n```text\nName Value\nfoo 1\n```") || strings.Contains(text, "Como funciona?") {
		t.Fatal(text)
	}
}
