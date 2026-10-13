package search

// BasicPythonHello recognizes only a plain introductory request. Extra topics
// such as Tkinter, asyncio or a web server keep the normal documentation search.
func BasicPythonHello(query string, technologies []string) bool {
	if len(technologies) != 1 || technologies[0] != "python" {
		return false
	}
	allowed := map[string]bool{}
	for _, term := range []string{"hello", "world", "olá", "ola", "mundo", "python", "fazer", "write", "make", "create", "escrever", "imprimir", "mostrar", "print", "exemplo", "example", "simples", "simple", "básico", "basico", "basic", "terminal", "console", "programa", "program"} {
		allowed[term] = true
	}
	seen := map[string]bool{}
	for _, term := range Terms(query) {
		if !allowed[term] {
			return false
		}
		seen[term] = true
	}
	return (seen["hello"] && seen["world"]) || ((seen["olá"] || seen["ola"]) && seen["mundo"])
}
