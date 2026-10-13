package extraction

import (
	"bytes"
	"github.com/PuerkitoBio/goquery"
	"github.com/hugojanuario/devdocs/internal/model"
	"strings"
	"unicode/utf8"
)

type Document struct {
	Title, Summary, Description, Version string
	Examples                             []model.Example
}

func clean(s string) string { return strings.Join(strings.Fields(s), " ") }
func trim(s string, n int) string {
	r := []rune(s)
	if len(r) > n {
		return string(r[:n])
	}
	return s
}
func relevance(text string, terms []string) int {
	n := 0
	t := strings.ToLower(text)
	for _, w := range terms {
		if strings.Contains(t, w) {
			n++
		}
	}
	return n
}
func Extract(raw []byte, terms []string, reference, tech string) (Document, error) {
	doc, e := goquery.NewDocumentFromReader(bytes.NewReader(raw))
	if e != nil {
		return Document{}, e
	}
	title := clean(doc.Find("h1").First().Text())
	if title == "" {
		title = clean(doc.Find("title").Text())
	}
	version := doc.Find(`meta[name="version"],meta[name="docsearch:version"]`).First().AttrOr("content", "")
	doc.Find("script,style,nav,header,footer,aside,.sidebar,.toc,.wy-nav-side,.headerlink,.copybtn,.linenos,.line-numbers,.navigation").Remove()
	content := doc.Find("main,article,[role=main],.document,.rst-content").First()
	if content.Length() == 0 {
		content = doc.Find("body")
	}
	paragraphs := []string{}
	content.Find("p").Each(func(_ int, s *goquery.Selection) {
		t := clean(s.Text())
		if utf8.RuneCountInString(t) >= 30 && len(t) <= 12000 {
			paragraphs = append(paragraphs, t)
		}
	})
	d := Document{Title: trim(title, 255), Version: version, Examples: []model.Example{}}
	best, second := -1, -1
	bs, ss := 0, 0
	for i, p := range paragraphs {
		score := relevance(p, terms)
		if score > bs {
			second, ss = best, bs
			best, bs = i, score
		} else if score > ss {
			second, ss = i, score
		}
	}
	if best >= 0 {
		d.Summary = trim(paragraphs[best], 2000)
	}
	if second >= 0 {
		d.Description = trim(paragraphs[second], 3000)
	}
	type block struct {
		example model.Example
		score   int
	}
	blocks := []block{}
	content.Find("pre").Each(func(_ int, s *goquery.Selection) {
		code := strings.Trim(s.Text(), "\r\n")
		if len(code) == 0 || len(code) > 16000 || relevance(code, terms) == 0 {
			return
		}
		language := s.AttrOr("data-language", "")
		classes := s.AttrOr("class", "") + " " + s.Find("code").AttrOr("class", "") + " " + s.Parent().AttrOr("class", "")
		for _, c := range strings.Fields(classes) {
			if strings.HasPrefix(c, "language-") {
				language = strings.TrimPrefix(c, "language-")
			}
			if strings.HasPrefix(c, "highlight-") {
				language = strings.TrimPrefix(c, "highlight-")
			}
		}
		if language == "" {
			switch tech {
			case "python", "pandas", "airflow", "sqlalchemy", "fastapi", "spark":
				language = "python"
			case "postgresql":
				language = "sql"
			case "go":
				language = "go"
			case "docker":
				language = "shell"
			case "kubernetes":
				language = "yaml"
			default:
				language = "text"
			}
		}
		explanations := []model.Explanation{}
		p := s.Parent()
		preceding := s.PrevAllFiltered("p").First()
		if preceding.Length() == 0 {
			preceding = p.PrevAllFiltered("p").First()
		}
		quote := clean(preceding.Text())
		// This is a source quote, never an invented line-by-line interpretation.
		if len(quote) >= 20 && len(quote) <= 2000 {
			explanations = append(explanations, model.Explanation{Snippet: code, Description: quote, ReferenceID: reference})
		}
		blocks = append(blocks, block{model.Example{Language: language, Code: code, Explanations: explanations, ReferenceID: reference}, relevance(code, terms)})
	})
	for len(blocks) > 0 && len(d.Examples) < 3 {
		best := 0
		for i := range blocks {
			if blocks[i].score > blocks[best].score {
				best = i
			}
		}
		d.Examples = append(d.Examples, blocks[best].example)
		blocks = append(blocks[:best], blocks[best+1:]...)
	}
	return d, nil
}
