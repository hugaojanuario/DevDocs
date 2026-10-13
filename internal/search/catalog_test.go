package search

import (
	"context"
	"errors"
	"strings"
	"testing"
)

type catalogFetch struct {
	fail     bool
	redirect bool
}

func (f catalogFetch) Fetch(_ context.Context, u string, _ int64) ([]byte, string, error) {
	if f.fail {
		return nil, "", errors.New("source offline")
	}
	final := u
	if f.redirect {
		final = "https://evil.test/sitemap.xml"
	}
	if strings.Contains(u, "postgresql") {
		return []byte(`<html><a href="queries-with.html">WITH Queries (Common Table Expressions)</a><a href="https://evil.test/queries-with.html">WITH Queries</a><a href="https://www.postgresql.org/download/">Download</a></html>`), final, nil
	}
	if strings.Contains(u, "docker") {
		return []byte(`<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9"><url><loc>https://docs.docker.com/compose/gettingstarted/</loc></url><url><loc>https://evil.test/compose/</loc></url></urlset>`), final, nil
	}
	return []byte(`<html><p>Unrelated content</p></html>`), final, nil
}
func TestPublicCatalogFindsPostgresAndDockerWithoutPaidSearch(t *testing.T) {
	p := &Official{Fetcher: catalogFetch{}}
	for _, tt := range []struct{ q, tech, path string }{{"Como escrever uma CTE no PostgreSQL?", "postgresql", "queries-with.html"}, {"Como subir serviços com Docker Compose?", "docker", "compose/gettingstarted/"}} {
		results, err := p.Search(context.Background(), tt.q, []string{tt.tech})
		if err != nil || len(results) != 1 || !strings.Contains(results[0].URL, tt.path) {
			t.Fatal(results, err)
		}
	}
	if !contains(TopicTerms("carregar .env em Python", []string{"python"}), "dotenv") {
		t.Fatal("dotenv intent not normalized")
	}
}
func TestNoResultsAndUnavailableSourcesAreDifferent(t *testing.T) {
	p := &Official{Fetcher: catalogFetch{}}
	if _, err := p.Search(context.Background(), "unknown topic", []string{"docker"}); !errors.Is(err, ErrNotFound) {
		t.Fatal("empty index treated as outage", err)
	}
	p.Fetcher = catalogFetch{fail: true}
	if _, err := p.Search(context.Background(), "Docker Compose", []string{"docker"}); !errors.Is(err, ErrUnavailable) || errors.Is(err, ErrNotFound) {
		t.Fatal("outage treated as empty index", err)
	}
	p.Fetcher = catalogFetch{redirect: true}
	if results, _ := p.Search(context.Background(), "Docker Compose", []string{"docker"}); len(results) > 0 {
		t.Fatal("off-site sitemap accepted", results)
	}
}

func TestCatalogPrefersIntroductionForGenericComposeQuestion(t *testing.T) {
	terms := TopicTerms("Como subir serviços com Docker Compose?", []string{"docker"})
	intro, ok := catalogResult("https://docs.docker.com/compose/gettingstarted/", "", "docker", terms)
	advanced, _ := catalogResult("https://docs.docker.com/compose/how-tos/gpu-support/", "", "docker", terms)
	if !ok || intro.Score <= advanced.Score {
		t.Fatal(intro, advanced)
	}
}
