package search

import "testing"

func TestBasicHelloDoesNotSwallowSpecificRequests(t *testing.T) {
	for _, q := range []string{"Como fazer um hello world em Python?", "Hello world", "Como imprimir olá mundo em Python?", "how to write hello world in Python"} {
		if !BasicPythonHello(q, []string{"python"}) {
			t.Fatal("missed basic intent", q)
		}
	}
	for _, q := range []string{"Hello world em Python com Tkinter", "Hello world em Python asyncio", "Hello world Python Docker", "Hello world Python 2.7", "Python print CSV"} {
		if BasicPythonHello(q, []string{"python"}) {
			t.Fatal("swallowed specific intent", q)
		}
	}
	if BasicPythonHello("hello world", []string{"go"}) {
		t.Fatal("swallowed another technology")
	}
}
