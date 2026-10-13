package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	GroqKey, GroqModel                                                                     string
	Addr, DatabaseURL, CLIToken, WebPassword, Origin, BraveKey, SearchProvider, Migrations string
	SecureCookie                                                                           bool
	CacheTTL                                                                               time.Duration
	CacheEntries                                                                           int
}

func Load() (Config, error) {
	c := Config{Addr: env("API_ADDR", ":8081"), DatabaseURL: os.Getenv("DATABASE_URL"), CLIToken: os.Getenv("CLI_TOKEN"), WebPassword: os.Getenv("WEB_PASSWORD"), Origin: env("WEB_ORIGIN", "http://localhost:8080"), BraveKey: os.Getenv("BRAVE_API_KEY"), SearchProvider: env("SEARCH_PROVIDER", "official"), Migrations: env("MIGRATIONS_DIR", "db/migrations"), CacheTTL: 5 * time.Minute, CacheEntries: 32}
	c.CLIToken = strings.TrimSpace(c.CLIToken)
	c.GroqKey = strings.TrimSpace(os.Getenv("GROQ_API_KEY"))
	c.GroqModel = env("GROQ_MODEL", "openai/gpt-oss-20b")
	if strings.ContainsAny(c.GroqKey, "\r\n\t ") || len(c.GroqModel) > 100 || strings.ContainsAny(c.GroqModel, "\r\n\t ") {
		return c, fmt.Errorf("invalid Groq configuration")
	}
	if strings.ContainsAny(c.CLIToken, "\r\n\t ") {
		return c, fmt.Errorf("CLI_TOKEN cannot contain whitespace")
	}
	if c.DatabaseURL == "" || len(c.CLIToken) < 32 || (len(c.WebPassword) < 16 || len(c.WebPassword) > 72) {
		return c, fmt.Errorf("DATABASE_URL, CLI_TOKEN (32+ chars) and WEB_PASSWORD (16+ chars) required")
	}
	if c.CLIToken == c.WebPassword {
		return c, fmt.Errorf("CLI token and web password must differ")
	}
	u, e := url.Parse(c.Origin)
	if e != nil || u.Host == "" || u.Path != "" || u.RawQuery != "" || (u.Scheme != "http" && u.Scheme != "https") {
		return c, fmt.Errorf("WEB_ORIGIN must be a bare http(s) origin")
	}
	c.SecureCookie = u.Scheme == "https"
	if !c.SecureCookie && u.Hostname() != "localhost" && u.Hostname() != "127.0.0.1" && os.Getenv("TRUST_PRIVATE_NETWORK") != "true" {
		return c, fmt.Errorf("remote access requires HTTPS or TRUST_PRIVATE_NETWORK=true on a trusted private network")
	}
	if c.SearchProvider != "official" && c.SearchProvider != "brave" {
		return c, fmt.Errorf("unknown SEARCH_PROVIDER")
	}
	if c.SearchProvider == "brave" && c.BraveKey == "" {
		return c, fmt.Errorf("BRAVE_API_KEY required for brave")
	}
	if v := os.Getenv("CACHE_TTL"); v != "" {
		d, e := time.ParseDuration(v)
		if e != nil || d < 0 || d > 15*time.Minute {
			return c, fmt.Errorf("CACHE_TTL must be 0..15m")
		}
		c.CacheTTL = d
	}
	if v := os.Getenv("CACHE_ENTRIES"); v != "" {
		n, e := strconv.Atoi(v)
		if e != nil || n < 0 || n > 128 {
			return c, fmt.Errorf("CACHE_ENTRIES must be 0..128")
		}
		c.CacheEntries = n
	}
	c.CLIToken = strings.TrimSpace(c.CLIToken)
	return c, nil
}
func env(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
