package api

import (
	"github.com/hugojanuario/devdocs/internal/model"
	"github.com/hugojanuario/devdocs/internal/security"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

func text(s string, min, max int) bool {
	n := utf8.RuneCountInString(s)
	if !utf8.ValidString(s) || n < min || n > max {
		return false
	}
	for _, r := range s {
		if unicode.IsControl(r) && r != '\n' && r != '\r' && r != '\t' {
			return false
		}
	}
	return true
}
func validSearch(r *model.SearchRequest) bool {
	r.Query = strings.Join(strings.Fields(r.Query), " ")
	if r.Language == "" {
		r.Language = "pt-BR"
	}
	if !text(r.Query, 3, 600) || (r.Language != "pt-BR" && r.Language != "en") || len(r.PreferredSources) > 2 {
		return false
	}
	seen := map[string]bool{}
	for _, s := range r.PreferredSources {
		if (s != "official" && s != "complementary") || seen[s] {
			return false
		}
		seen[s] = true
	}
	return true
}

var refID = regexp.MustCompile(`^source-[0-9]+$`)

func validAnswer(a model.Answer) bool {
	if !text(a.Query, 3, 600) || !text(a.Title, 1, 255) || !text(a.Technology, 0, 100) || !text(a.Version, 0, 100) || !text(a.Summary, 0, 12000) || !text(a.Description, 0, 12000) || (a.Mode != "deterministic" && a.Mode != "ai_assisted") || (a.Status != "completed" && a.Status != "partial") || (a.Quality != "extracted" && a.Quality != "limited") || (a.Language != "pt-BR" && a.Language != "en") || len(a.References) == 0 || len(a.References) > 16 || len(a.Examples) > 24 || len(a.Observations) > 32 || a.Examples == nil || a.Evidence == nil || a.Observations == nil {
		return false
	}
	refs := map[string]model.Reference{}
	for _, r := range a.References {
		u, e := security.ValidateURL(r.URL)
		if e != nil || !refID.MatchString(r.ID) || r.Domain != u.Hostname() || !text(r.Title, 1, 255) || !text(r.Excerpt, 0, 12000) || !text(r.Technology, 0, 100) || !text(r.Version, 0, 100) {
			return false
		}
		if _, ok := refs[r.ID]; ok {
			return false
		}
		refs[r.ID] = r
	}
	for _, ex := range a.Examples {
		if _, ok := refs[ex.ReferenceID]; !ok {
			return false
		}
		if !text(ex.Code, 1, 16000) || !text(ex.Language, 1, 50) || ex.Explanations == nil {
			return false
		}
		for _, xp := range ex.Explanations {
			if _, ok := refs[xp.ReferenceID]; !ok {
				return false
			}
			if xp.ReferenceID != ex.ReferenceID || !strings.Contains(ex.Code, xp.Snippet) || !text(xp.Snippet, 1, 16000) || !text(xp.Description, 1, 2000) {
				return false
			}
		}
	}
	for _, o := range a.Observations {
		if !text(o, 1, 4000) {
			return false
		}
	}
	if len(a.Evidence) > 64 {
		return false
	}
	hasSummary, hasDescription := a.Summary == "", a.Description == ""
	hasExamples := map[int]bool{}
	for _, ev := range a.Evidence {
		r, ok := refs[ev.ReferenceID]
		if !ok || !text(ev.Quote, 1, 16000) {
			return false
		}
		switch ev.Field {
		case "summary":
			if ev.Quote != a.Summary || ev.Quote != r.Excerpt {
				return false
			}
			hasSummary = true
		case "description":
			if ev.Quote != a.Description {
				return false
			}
			hasDescription = true
		case "example":
			matched := false
			for i, ex := range a.Examples {
				if ex.ReferenceID == ev.ReferenceID && ex.Code == ev.Quote {
					hasExamples[i] = true
					matched = true
				}
			}
			if !matched {
				return false
			}
		default:
			return false
		}
	}
	return hasSummary && hasDescription && len(hasExamples) == len(a.Examples)
}
