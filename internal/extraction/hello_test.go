package extraction

import "testing"

func TestHelloRequiresStandaloneLiteralEvidence(t *testing.T) {
	for _, code := range []string{`import asyncio
print("Hello World")`, `def hello():
    print("Hello World")`, `label = "Hello World"`, `print("Goodbye")`, `>>> print("Hello World")`} {
		got, err := PythonHello([]byte("<pre>" + code + "</pre>"))
		if err != nil || got != "" {
			t.Fatalf("accepted unrelated or dependent example %q: %q %v", code, got, err)
		}
	}
	code, err := PythonHello([]byte(`<pre>print("Hello World")
</pre>`))
	if err != nil || code != `print("Hello World")` {
		t.Fatal(code, err)
	}
}

func TestPrintSummaryScopedToFunction(t *testing.T) {
	raw := []byte(`<main><p>Unrelated Python information.</p><dl><dt id="print">print</dt><dd><p>Print objects to the text stream file.</p><p>Additional options.</p></dd></dl></main>`)
	summary, err := PythonPrintSummary(raw)
	if err != nil || summary != "Print objects to the text stream file." {
		t.Fatal(summary, err)
	}
}
