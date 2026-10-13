package api

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"strings"
	"sync"
	"time"
)

type session struct {
	CSRF    string
	Expires time.Time
}
type auth struct {
	token    [32]byte
	password []byte
	origin   string
	secure   bool
	mu       sync.Mutex
	sessions map[[32]byte]session
}

func newAuth(token, password, origin string, secure bool) *auth {
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return &auth{token: sha256.Sum256([]byte(token)), password: hash, origin: origin, secure: secure, sessions: map[[32]byte]session{}}
}
func secret() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		panic("randomness unavailable")
	}
	return hex.EncodeToString(b)
}
func (a *auth) lookup(r *http.Request) (session, bool) {
	c, e := r.Cookie("devdocs_session")
	if e != nil {
		return session{}, false
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	now := time.Now()
	for k, v := range a.sessions {
		if now.After(v.Expires) {
			delete(a.sessions, k)
		}
	}
	s, ok := a.sessions[sha256.Sum256([]byte(c.Value))]
	return s, ok
}
func (a *auth) middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := r.Header.Get("Authorization")
		if h != "" {
			if !strings.HasPrefix(h, "Bearer ") {
				fail(w, r, 401, "unauthorized", "Autenticação necessária")
				return
			}
			provided := sha256.Sum256([]byte(strings.TrimPrefix(h, "Bearer ")))
			if subtle.ConstantTimeCompare(a.token[:], provided[:]) != 1 {
				fail(w, r, 401, "unauthorized", "Autenticação necessária")
				return
			}
			next.ServeHTTP(w, r)
			return
		}
		s, ok := a.lookup(r)
		if !ok {
			fail(w, r, 401, "unauthorized", "Autenticação necessária")
			return
		}
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			if r.Header.Get("Origin") != a.origin || subtle.ConstantTimeCompare([]byte(r.Header.Get("X-CSRF-Token")), []byte(s.CSRF)) != 1 {
				fail(w, r, 403, "csrf_failed", "Origem ou CSRF inválido")
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}
func (a *auth) cookie(w http.ResponseWriter, value string, maxAge int) {
	http.SetCookie(w, &http.Cookie{Name: "devdocs_session", Value: value, Path: "/api/v1", MaxAge: maxAge, HttpOnly: true, Secure: a.secure, SameSite: http.SameSiteStrictMode})
}
func (a *auth) login(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Origin") != a.origin {
		fail(w, r, 403, "origin_rejected", "Origem inválida")
		return
	}
	var body struct {
		Password string `json:"password"`
	}
	if !decode(w, r, &body) {
		return
	}
	if len(body.Password) == 0 || len(body.Password) > 1024 {
		fail(w, r, 400, "invalid_input", "Senha inválida")
		return
	}
	if bcrypt.CompareHashAndPassword(a.password, []byte(body.Password)) != nil {
		fail(w, r, 401, "unauthorized", "Credenciais inválidas")
		return
	}
	token := secret()
	s := session{CSRF: secret(), Expires: time.Now().Add(8 * time.Hour)}
	a.mu.Lock()
	for k, v := range a.sessions {
		if time.Now().After(v.Expires) {
			delete(a.sessions, k)
		}
	}
	if c, e := r.Cookie("devdocs_session"); e == nil {
		delete(a.sessions, sha256.Sum256([]byte(c.Value)))
	}
	if len(a.sessions) >= 128 {
		a.mu.Unlock()
		fail(w, r, 503, "session_limit", "Limite de sessões atingido")
		return
	}
	a.sessions[sha256.Sum256([]byte(token))] = s
	a.mu.Unlock()
	a.cookie(w, token, 28800)
	jsonResponse(w, 200, map[string]any{"csrf_token": s.CSRF, "expires_at": s.Expires})
}
func (a *auth) current(w http.ResponseWriter, r *http.Request) {
	s, ok := a.lookup(r)
	if !ok {
		fail(w, r, 401, "unauthorized", "Sessão web necessária")
		return
	}
	jsonResponse(w, 200, map[string]any{"csrf_token": s.CSRF, "expires_at": s.Expires})
}
func (a *auth) logout(w http.ResponseWriter, r *http.Request) {
	if c, e := r.Cookie("devdocs_session"); e == nil {
		a.mu.Lock()
		delete(a.sessions, sha256.Sum256([]byte(c.Value)))
		a.mu.Unlock()
	}
	a.cookie(w, "", -1)
	w.WriteHeader(204)
}
