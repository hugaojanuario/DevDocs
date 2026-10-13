package extraction

import "testing"

func TestEvidenceAndIndentation(t *testing.T) {
	raw := []byte(`<html><head><title>Read CSV</title></head><body><nav>Ignore CSV</nav><main><h1>CSV</h1><p>Read CSV with this documented example for data files.</p><pre class="language-python"><code>if True:
    read_csv("data.csv")
</code></pre><script>read_csv secret</script></main></body></html>`)
	d, e := Extract(raw, []string{"csv"}, "source-1", "pandas")
	if e != nil || len(d.Examples) != 1 {
		t.Fatal(d, e)
	}
	if d.Examples[0].Code != "if True:\n    read_csv(\"data.csv\")" {
		t.Fatal("indentation changed")
	}
	if d.Examples[0].Explanations[0].Description != d.Summary {
		t.Fatal("not cited evidence")
	}
	empty, _ := Extract([]byte(`<main><p>An unrelated paragraph that contains no evidence.</p><pre>nothing()</pre></main>`), []string{"csv"}, "source-1", "python")
	if empty.Summary != "" || len(empty.Examples) > 0 {
		t.Fatal("invented evidence")
	}
}
