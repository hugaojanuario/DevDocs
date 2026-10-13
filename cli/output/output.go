package output

import (
	"encoding/json"
	"fmt"
	"github.com/hugojanuario/devdocs/internal/model"
	"io"
	"strings"
)

func JSON(w io.Writer, v any) error {
	e := json.NewEncoder(w)
	e.SetIndent("", "  ")
	return e.Encode(v)
}

// Prevent terminal-control injection from third-party documentation.
func safe(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r == '\n' || r == '\t' || r >= 32 && r != 127 && !(r >= 128 && r <= 159) {
			b.WriteRune(r)
		}
	}
	return b.String()
}
func Render(w io.Writer, a model.Answer, format string, color bool) error {
	if format == "json" {
		return JSON(w, a)
	}
	if format != "text" && format != "markdown" {
		return fmt.Errorf("formato deve ser text, markdown ou json")
	}
	heading := func(s string) {
		if format == "markdown" {
			fmt.Fprintf(w, "## %s\n\n", safe(s))
		} else if color {
			fmt.Fprintf(w, "\033[1;36m%s\033[0m\n", safe(s))
		} else {
			fmt.Fprintln(w, safe(s))
		}
	}
	heading(a.Title)
	fmt.Fprintf(w, "%s | %s | %s\n\n", safe(a.Technology), safe(a.Mode), safe(a.Status))
	heading("O que é?")
	fmt.Fprintln(w, safe(a.Summary))
	heading("Como funciona?")
	fmt.Fprintln(w, safe(a.Description))
	for _, ex := range a.Examples {
		heading("Exemplo (" + ex.Language + ")")
		if format == "markdown" {
			fence := "```"
			for strings.Contains(ex.Code, fence) {
				fence += "`"
			}
			fmt.Fprintf(w, "%s%s\n%s\n%s\n\n", fence, safe(ex.Language), safe(ex.Code), fence)
		} else {
			fmt.Fprintln(w, safe(ex.Code))
		}
		for _, xp := range ex.Explanations {
			fmt.Fprintf(w, "%s [%s]\n", safe(xp.Description), safe(xp.ReferenceID))
		}
	}
	heading("Observações")
	for _, o := range a.Observations {
		fmt.Fprintln(w, safe(o))
	}
	heading("Fontes")
	for _, r := range a.References {
		kind := "complementar"
		if r.Official {
			kind = "oficial"
		}
		fmt.Fprintf(w, "[%s] %s (%s)\n%s\n", safe(r.ID), safe(r.Title), kind, safe(r.URL))
	}
	return nil
}
