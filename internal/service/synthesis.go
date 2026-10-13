package service

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/hugojanuario/devdocs/internal/model"
)

var ErrAINotConfigured = errors.New("AI provider not configured")

func (e *Engine) finish(ctx context.Context, r model.SearchRequest, key [32]byte, generation uint64, a model.Answer) (model.Answer, error) {
	if r.AIEnabled {
		// Providers work on an isolated copy: even an invalid/failed synthesis
		// cannot corrupt the deterministic fallback or its in-memory evidence.
		raw, _ := json.Marshal(a)
		var candidate model.Answer
		_ = json.Unmarshal(raw, &candidate)
		result, err := e.Synthesis.Answer(ctx, candidate)
		if err == nil {
			a = result
		} else {
			message := "A síntese por IA está indisponível. Exibindo os trechos da documentação."
			if r.Language == "en" {
				message = "AI synthesis is unavailable. Showing documentation excerpts."
			}
			a.Observations = append(a.Observations, message)
			// A transient Groq error should not suppress another attempt for the
			// entire search-cache TTL.
			return a, nil
		}
	}
	return e.remember(key, generation, a)
}
