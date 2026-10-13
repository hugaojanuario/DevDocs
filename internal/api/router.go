package api

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/hugojanuario/devdocs/internal/bookmarks"
	"github.com/hugojanuario/devdocs/internal/config"
	"github.com/hugojanuario/devdocs/internal/model"
	"github.com/hugojanuario/devdocs/internal/search"
	"github.com/hugojanuario/devdocs/internal/service"
	"github.com/jackc/pgx/v5"
	"io"
	"mime"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

type key int

const requestKey key = 0

func jsonResponse(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
func fail(w http.ResponseWriter, r *http.Request, status int, code, message string) {
	requestID, _ := r.Context().Value(requestKey).(string)
	jsonResponse(w, status, map[string]any{"error": map[string]string{"code": code, "message": message, "request_id": requestID}})
}
func decode(w http.ResponseWriter, r *http.Request, out any) bool {
	media, _, _ := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if media != "application/json" {
		fail(w, r, 400, "invalid_content_type", "Use application/json")
		return false
	}
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	err := d.Decode(out)
	if err == nil {
		extra := d.Decode(new(any))
		if extra == nil {
			err = errors.New("trailing body")
		} else if extra != io.EOF {
			err = extra
		}
	}
	if err != nil {
		var max *http.MaxBytesError
		if errors.As(err, &max) {
			fail(w, r, 413, "payload_too_large", "Limite de 1 MiB")
		} else {
			fail(w, r, 400, "invalid_input", "JSON inválido")
		}
		return false
	}
	return true
}

type bucket struct {
	mu       sync.Mutex
	tokens   float64
	last     time.Time
	capacity float64
}

func rate(capacity float64) func(http.Handler) http.Handler {
	b := &bucket{tokens: capacity, last: time.Now(), capacity: capacity}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			b.mu.Lock()
			now := time.Now()
			b.tokens += now.Sub(b.last).Minutes() * b.capacity
			if b.tokens > b.capacity {
				b.tokens = b.capacity
			}
			b.last = now
			ok := b.tokens >= 1
			if ok {
				b.tokens--
			}
			b.mu.Unlock()
			if !ok {
				w.Header().Set("Retry-After", "60")
				fail(w, r, 429, "rate_limited", "Limite de requisições")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
func New(c config.Config, store bookmarks.Store, engine service.Searcher) http.Handler {
	a := newAuth(c.CLIToken, c.WebPassword, c.Origin, c.SecureCookie)
	r := chi.NewRouter()
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			rid := uuid.NewString()
			req = req.WithContext(context.WithValue(req.Context(), requestKey, rid))
			w.Header().Set("X-Request-ID", rid)
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("Cache-Control", "no-store")
			defer func() {
				if recover() != nil {
					fail(w, req, 500, "internal_error", "Falha interna")
				}
			}()
			origin := req.Header.Get("Origin")
			if origin != "" && origin != c.Origin {
				fail(w, req, 403, "origin_rejected", "Origem inválida")
				return
			}
			if origin == c.Origin {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				w.Header().Set("Vary", "Origin")
			}
			if req.Method == "OPTIONS" {
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-CSRF-Token")
				w.WriteHeader(204)
				return
			}
			next.ServeHTTP(w, req)
		})
	})
	r.NotFound(func(w http.ResponseWriter, r *http.Request) { fail(w, r, 404, "not_found", "Endpoint não encontrado") })
	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		fail(w, r, 405, "method_not_allowed", "Método não permitido")
	})
	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/health", func(w http.ResponseWriter, r *http.Request) { jsonResponse(w, 200, map[string]string{"status": "ok"}) })
		r.Get("/ready", func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
			defer cancel()
			if store.Ready(ctx) != nil {
				fail(w, r, 503, "not_ready", "Banco indisponível")
				return
			}
			jsonResponse(w, 200, map[string]string{"status": "ready"})
		})
		r.With(rate(5)).Post("/auth/session", a.login)
		r.Group(func(r chi.Router) {
			r.Use(a.middleware)
			r.Use(rate(120))
			r.Get("/auth/session", a.current)
			r.Delete("/auth/session", a.logout)
			r.Get("/capabilities", func(w http.ResponseWriter, r *http.Request) {
				provider, modelName := "", ""
				if c.GroqKey != "" {
					provider, modelName = "groq", c.GroqModel
				}
				jsonResponse(w, 200, map[string]any{"ai_available": c.GroqKey != "", "ai_provider": provider, "ai_model": modelName})
			})
			r.With(rate(20)).Post("/search", func(w http.ResponseWriter, r *http.Request) {
				var body model.SearchRequest
				if !decode(w, r, &body) {
					return
				}
				if !validSearch(&body) {
					fail(w, r, 400, "invalid_input", "Consulta ou opções inválidas")
					return
				}
				if body.AIEnabled && c.GroqKey == "" {
					fail(w, r, 503, "ai_unavailable", "IA opcional não configurada")
					return
				}
				answer, err := engine.Search(r.Context(), body)
				if err != nil {
					if errors.Is(err, service.ErrAINotConfigured) {
						fail(w, r, 503, "ai_unavailable", "IA opcional não configurada")
					} else if errors.Is(err, search.ErrProvider) {
						fail(w, r, 502, "provider_failed", "Provedor de busca falhou")
					} else {
						fail(w, r, 503, "search_unavailable", "Nenhuma evidência disponível; configure um provedor ou tente outra tecnologia")
					}
					return
				}
				jsonResponse(w, 200, answer)
			})
			r.Get("/technologies", func(w http.ResponseWriter, r *http.Request) { jsonResponse(w, 200, search.Technologies) })
			r.Get("/sources", func(w http.ResponseWriter, r *http.Request) { jsonResponse(w, 200, search.Sources) })
			r.Post("/saved", func(w http.ResponseWriter, r *http.Request) {
				var body struct {
					Response model.Answer `json:"response"`
					Title    *string      `json:"title"`
				}
				if !decode(w, r, &body) {
					return
				}
				if !validAnswer(body.Response) {
					fail(w, r, 400, "invalid_input", "Resposta e evidências inválidas")
					return
				}
				title := body.Response.Title
				if body.Title != nil {
					title = strings.TrimSpace(*body.Title)
				}
				if !text(title, 1, 255) {
					fail(w, r, 400, "invalid_input", "Título inválido")
					return
				}
				saved, err := store.Save(r.Context(), body.Response, title)
				if err != nil {
					dbError(w, r, err)
					return
				}
				w.Header().Set("Location", "/api/v1/saved/"+saved.ID)
				jsonResponse(w, 201, saved)
			})
			r.Get("/saved", func(w http.ResponseWriter, r *http.Request) {
				page, size := 1, 20
				var err error
				if v := r.URL.Query().Get("page"); v != "" {
					page, err = strconv.Atoi(v)
					if err != nil {
						page = 0
					}
				}
				if v := r.URL.Query().Get("page_size"); v != "" {
					size, err = strconv.Atoi(v)
					if err != nil {
						size = 0
					}
				}
				q, tech := r.URL.Query().Get("q"), r.URL.Query().Get("technology")
				order := r.URL.Query().Get("order")
				if order == "" {
					order = "recent"
				}
				if order != "recent" && order != "oldest" && order != "title" {
					fail(w, r, 400, "invalid_input", "Ordenação inválida")
					return
				}
				if page < 1 || page > 100000 || size < 1 || size > 100 || !text(q, 0, 600) || !text(tech, 0, 100) {
					fail(w, r, 400, "invalid_input", "Paginação ou filtros inválidos")
					return
				}
				p, err := store.List(r.Context(), page, size, q, tech, order)
				if err != nil {
					dbError(w, r, err)
					return
				}
				jsonResponse(w, 200, p)
			})
			r.Route("/saved/{id}", func(r chi.Router) {
				r.Use(func(next http.Handler) http.Handler {
					return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						if _, err := uuid.Parse(chi.URLParam(r, "id")); err != nil {
							fail(w, r, 400, "invalid_id", "UUID inválido")
							return
						}
						next.ServeHTTP(w, r)
					})
				})
				r.Get("/", func(w http.ResponseWriter, r *http.Request) {
					v, err := store.Get(r.Context(), chi.URLParam(r, "id"))
					if err != nil {
						dbError(w, r, err)
						return
					}
					jsonResponse(w, 200, v)
				})
				r.Patch("/", func(w http.ResponseWriter, r *http.Request) {
					var body struct {
						Title string `json:"title"`
					}
					if !decode(w, r, &body) {
						return
					}
					body.Title = strings.TrimSpace(body.Title)
					if !text(body.Title, 1, 255) {
						fail(w, r, 400, "invalid_input", "Título inválido")
						return
					}
					v, err := store.Rename(r.Context(), chi.URLParam(r, "id"), body.Title)
					if err != nil {
						dbError(w, r, err)
						return
					}
					jsonResponse(w, 200, v)
				})
				r.Delete("/", func(w http.ResponseWriter, r *http.Request) {
					ok, err := store.Delete(r.Context(), chi.URLParam(r, "id"))
					if err != nil {
						dbError(w, r, err)
						return
					}
					if !ok {
						fail(w, r, 404, "not_found", "Consulta não encontrada")
						return
					}
					if cache, ok := engine.(interface{ Clear() }); ok {
						cache.Clear()
					}
					w.WriteHeader(204)
				})
			})
		})
	})
	return r
}
func dbError(w http.ResponseWriter, r *http.Request, err error) {
	if errors.Is(err, pgx.ErrNoRows) {
		fail(w, r, 404, "not_found", "Consulta não encontrada")
	} else {
		fail(w, r, 503, "database_unavailable", "Banco indisponível")
	}
}
