package search

import (
	"bytes"
	"context"
	"encoding/xml"
	"net/url"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type catalog struct {
	technology, address string
	topics              []string
	page                bool
}

// Public publisher indexes/sitemaps, not search-engine result scraping.
// This covers collections without a Sphinx index using the same safe fetcher.
var catalogs = []catalog{
	{technology: "postgresql", address: "https://www.postgresql.org/docs/current/bookindex.html"},
	{technology: "docker", address: "https://docs.docker.com/sitemap.xml"},
	{technology: "go", address: "https://go.dev/doc/"},
	{technology: "spark", address: "https://spark.apache.org/docs/latest/index.html"},
	{technology: "dbt", address: "https://docs.getdbt.com/sitemap.xml"},
	{technology: "fastapi", address: "https://fastapi.tiangolo.com/"},
	{technology: "kubernetes", address: "https://kubernetes.io/sitemap.xml"},
	{technology: "python", address: "https://bbc2.github.io/python-dotenv/", topics: []string{"dotenv"}, page: true},
}

type sitemap struct {
	URLs []struct {
		Loc string `xml:"loc"`
	} `xml:"url"`
	Children []struct {
		Loc string `xml:"loc"`
	} `xml:"sitemap"`
}

func (p *Official) catalogSearch(ctx context.Context, q string, tech []string) (out []Result, attempted, available bool) {
	terms := TopicTerms(q, tech)
	for _, c := range catalogs {
		if !contains(tech, c.technology) {
			continue
		}
		if len(c.topics) > 0 {
			matched := false
			for _, term := range terms {
				if contains(c.topics, term) {
					matched = true
				}
			}
			if !matched {
				continue
			}
		}
		attempted = true
		raw, final, err := p.Fetcher.Fetch(ctx, c.address, 8<<20)
		if err != nil {
			continue
		}
		base, err := url.Parse(final)
		expected, _ := url.Parse(c.address)
		if err != nil || base.Scheme != "https" || base.Host != expected.Host {
			continue
		}
		if source, ok := Classify(final); !ok && !strings.HasSuffix(base.Path, "sitemap.xml") || ok && source.Technology != c.technology {
			continue
		}
		available = true
		if c.page {
			out = append(out, Result{URL: final, Score: 100})
			continue
		}
		var index sitemap
		if xml.Unmarshal(raw, &index) == nil && (len(index.URLs) > 0 || len(index.Children) > 0) {
			for _, child := range index.Children {
				u, err := url.Parse(child.Loc)
				if err != nil || u.Scheme != "https" || u.Host != base.Host || !strings.Contains(u.Path, "/en/") {
					continue
				}
				b, _, err := p.Fetcher.Fetch(ctx, u.String(), 8<<20)
				if err != nil {
					continue
				}
				var nested sitemap
				if xml.Unmarshal(b, &nested) == nil {
					index.URLs = append(index.URLs, nested.URLs...)
				}
				break // one English sitemap, never crawl localized copies or whole sites
			}
			for _, item := range index.URLs {
				if result, ok := catalogResult(item.Loc, "", c.technology, terms); ok {
					out = append(out, result)
				}
			}
			continue
		}
		doc, err := goquery.NewDocumentFromReader(bytes.NewReader(raw))
		if err != nil {
			continue
		}
		doc.Find("a[href]").Each(func(_ int, s *goquery.Selection) {
			target, err := base.Parse(s.AttrOr("href", ""))
			if err != nil {
				return
			}
			target.Fragment = ""
			if result, ok := catalogResult(target.String(), strings.Join(strings.Fields(s.Text()), " "), c.technology, terms); ok {
				out = append(out, result)
			}
		})
	}
	return
}

func catalogResult(address, title, technology string, terms []string) (Result, bool) {
	source, official := Classify(address)
	u, err := url.Parse(address)
	if err != nil || u.Scheme != "https" || !official || source.Technology != technology || u.RawQuery != "" {
		return Result{}, false
	}
	text := strings.ToLower(title + " " + u.Path)
	score := 0
	for _, term := range terms {
		if regexp.MustCompile(`(?:^|[^a-z0-9])` + regexp.QuoteMeta(term) + `(?:$|[^a-z0-9])`).MatchString(text) {
			score += 30
		}
		if term == "cte" && (strings.Contains(text, "common table expression") || strings.Contains(text, "queries-with")) {
			score += 80
		}
	}
	if score == 0 {
		return Result{}, false
	}
	for _, part := range []string{"gettingstarted", "getting-started", "tutorial", "quickstart", "introduction"} {
		if strings.Contains(text, part) {
			score += 60
			break
		}
	}
	// Favor the current English documentation over version archives and releases.
	if strings.Contains(text, "release") || strings.Contains(text, "deprecated") {
		score -= 20
	}
	u.Fragment = ""
	return Result{Title: title, URL: u.String(), Score: score}, true
}
