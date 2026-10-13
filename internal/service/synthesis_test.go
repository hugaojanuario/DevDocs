package service

import (
	"context"
	"errors"
	"github.com/hugojanuario/devdocs/internal/model"
	"testing"
	"time"
)

type failingSynthesis struct{ calls int }

func (s *failingSynthesis) Answer(_ context.Context, a model.Answer) (model.Answer, error) {
	s.calls++
	a.Examples[0].Code = "corrupted"
	return a, errors.New("upstream error")
}
func TestAIFailureKeepsEvidenceAndDoesNotCacheTransientFailure(t *testing.T) {
	s := &failingSynthesis{}
	e := New(&provider{}, fetch{}, time.Minute, 4)
	e.Synthesis = s
	for i := 0; i < 2; i++ {
		a, err := e.Search(context.Background(), model.SearchRequest{Query: "Python CSV", AIEnabled: true})
		if err != nil || a.Mode != "deterministic" || a.Examples[0].Code == "corrupted" || len(a.Observations) != 2 {
			t.Fatal(a, err)
		}
	}
	if s.calls != 2 {
		t.Fatal("failure cached", s.calls)
	}
}
