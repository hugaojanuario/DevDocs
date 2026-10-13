package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPrivateConfig(t *testing.T) {
	p := filepath.Join(t.TempDir(), "devdocs", "config.json")
	t.Setenv("DEVDOCS_CONFIG", p)
	c := Config{Server: "https://devdocs.example", Token: "secret"}
	if e := Save(c); e != nil {
		t.Fatal(e)
	}
	info, e := os.Stat(p)
	if e != nil || info.Mode().Perm() != 0600 {
		t.Fatal("permissions")
	}
	got, e := Load()
	if e != nil || got.Token != c.Token {
		t.Fatal(got, e)
	}
	if ValidateServer("http://remote.example", false) == nil {
		t.Fatal("insecure remote")
	}
	if ValidateServer("http://remote.example", true) != nil {
		t.Fatal("trusted network")
	}
}
