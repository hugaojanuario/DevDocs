package config

import "testing"

func TestSecretsAndRemoteTransport(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://devdocs:local@localhost/devdocs")
	t.Setenv("CLI_TOKEN", "                                ")
	t.Setenv("WEB_PASSWORD", "a-valid-web-password")
	if _, err := Load(); err == nil {
		t.Fatal("empty token after normalization accepted")
	}
	t.Setenv("CLI_TOKEN", "a-private-cli-token-longer-than-32-characters")
	t.Setenv("WEB_ORIGIN", "http://remote.example")
	t.Setenv("TRUST_PRIVATE_NETWORK", "false")
	if _, err := Load(); err == nil {
		t.Fatal("remote HTTP allowed")
	}
	t.Setenv("WEB_ORIGIN", "https://docs.example")
	if c, err := Load(); err != nil || !c.SecureCookie {
		t.Fatal(c.SecureCookie, err)
	}
	t.Setenv("CACHE_ENTRIES", "1000")
	if _, err := Load(); err == nil {
		t.Fatal("unbounded cache")
	}
}
