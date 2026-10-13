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
	if len(d.Examples[0].Explanations) != 0 || d.Examples[0].SourceCode != d.Examples[0].Code {
		t.Fatal("duplicate summary or missing source evidence")
	}
	empty, _ := Extract([]byte(`<main><p>An unrelated paragraph that contains no evidence.</p><pre>nothing()</pre></main>`), []string{"csv"}, "source-1", "python")
	if empty.Summary != "" || len(empty.Examples) > 0 {
		t.Fatal("invented evidence")
	}
}

func TestVersionFromDocumentTitle(t *testing.T) {
	d, e := Extract([]byte(`<html><head><title>csv — Python 3.14.5 documentation</title></head><body><main><h1>csv</h1><p>CSV reader documented for reading files.</p></main></body></html>`), []string{"csv"}, "source-1", "python")
	if e != nil || d.Version != "3.14.5" {
		t.Fatal(d.Version, e)
	}
}
