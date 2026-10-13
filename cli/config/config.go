package config

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
)

type Config struct {
	Server                string `json:"server"`
	Token                 string `json:"token,omitempty"`
	TrustedPrivateNetwork bool   `json:"trusted_private_network,omitempty"`
}

func Path() (string, error) {
	if p := os.Getenv("DEVDOCS_CONFIG"); p != "" {
		return p, nil
	}
	d, e := os.UserConfigDir()
	if e != nil {
		return "", e
	}
	return filepath.Join(d, "devdocs", "config.json"), nil
}
func Load() (Config, error) {
	c := Config{Server: "http://localhost:8080"}
	p, e := Path()
	if e != nil {
		return c, e
	}
	b, e := os.ReadFile(p)
	if os.IsNotExist(e) {
		e = nil
	} else if e == nil {
		e = json.Unmarshal(b, &c)
	}
	if v := os.Getenv("DEVDOCS_SERVER"); v != "" {
		c.Server = v
	}
	if v := os.Getenv("DEVDOCS_TOKEN"); v != "" {
		c.Token = v
	}
	if os.Getenv("DEVDOCS_TRUST_PRIVATE_NETWORK") == "true" {
		c.TrustedPrivateNetwork = true
	}
	return c, e
}
func ValidateServer(server string, trusted bool) error {
	u, e := url.Parse(server)
	if e != nil || u.Host == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" || (u.Path != "" && u.Path != "/") || (u.Scheme != "http" && u.Scheme != "https") {
		return fmt.Errorf("servidor deve ser uma origem HTTP(S)")
	}
	if u.Scheme == "http" && u.Hostname() != "localhost" && u.Hostname() != "127.0.0.1" && u.Hostname() != "::1" && !trusted {
		return fmt.Errorf("use HTTPS ou --trusted-private-network numa rede privada confiável")
	}
	return nil
}
func Save(c Config) error {
	p, e := Path()
	if e != nil {
		return e
	}
	if e = os.MkdirAll(filepath.Dir(p), 0700); e != nil {
		return e
	}
	b, e := json.MarshalIndent(c, "", "  ")
	if e != nil {
		return e
	}
	f, e := os.CreateTemp(filepath.Dir(p), ".config-*")
	if e != nil {
		return e
	}
	defer os.Remove(f.Name())
	if e = f.Chmod(0600); e != nil {
		f.Close()
		return e
	}
	if _, e = f.Write(b); e != nil {
		f.Close()
		return e
	}
	if e = f.Close(); e != nil {
		return e
	}
	return os.Rename(f.Name(), p)
}
