package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/hugojanuario/devdocs/cli/config"
	"io"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	Server, Token string
	HTTP          *http.Client
}

func New(c config.Config) (*Client, error) {
	if err := config.ValidateServer(c.Server, c.TrustedPrivateNetwork); err != nil {
		return nil, err
	}
	return &Client{Server: strings.TrimRight(c.Server, "/"), Token: c.Token, HTTP: &http.Client{Timeout: 65 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}}, nil
}
func (c *Client) Do(ctx context.Context, method, path string, body, out any) error {
	var raw []byte
	var e error
	if body != nil {
		raw, e = json.Marshal(body)
		if e != nil {
			return e
		}
	}
	r, e := http.NewRequestWithContext(ctx, method, c.Server+"/api/v1"+path, bytes.NewReader(raw))
	if e != nil {
		return fmt.Errorf("requisição inválida")
	}
	if c.Token != "" {
		r.Header.Set("Authorization", "Bearer "+c.Token)
	}
	if body != nil {
		r.Header.Set("Content-Type", "application/json")
	}
	res, e := c.HTTP.Do(r)
	if e != nil {
		return fmt.Errorf("API indisponível")
	}
	defer res.Body.Close()
	b, e := io.ReadAll(io.LimitReader(res.Body, (2<<20)+1))
	if e != nil || len(b) > 2<<20 {
		return fmt.Errorf("resposta inválida ou muito grande")
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		var envelope struct {
			Error struct{ Code, Message string }
		}
		_ = json.Unmarshal(b, &envelope)
		if envelope.Error.Code == "" {
			return fmt.Errorf("API retornou HTTP %d", res.StatusCode)
		}
		return fmt.Errorf("%s: %s (HTTP %d)", envelope.Error.Code, envelope.Error.Message, res.StatusCode)
	}
	if out != nil && res.StatusCode != 204 {
		if e = json.Unmarshal(b, out); e != nil {
			return fmt.Errorf("JSON inválido da API")
		}
	}
	return nil
}
