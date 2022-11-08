package snow

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
)

type Client struct {
	BaseURL         string
	Username        string
	Password        string
	TLSClientConfig *tls.Config
}

func (c *Client) Request() (*resty.Request, error) {
	rc := resty.New()
	rc.DisableWarn = true
	rc.SetBasicAuth(c.Username, c.Password)
	rc.SetTLSClientConfig(c.TLSClientConfig)
	rc.SetTimeout(10 * time.Second)
	rc.SetBaseURL(c.BaseURL)

	return rc.R(), nil
}

func (c *Client) Fetch(ctx context.Context, endpoint string, results interface{}) (interface{}, error) {
	r, err := c.Request()

	if err != nil {
		return nil, err
	}

	r.SetContext(ctx)

	resp, err := r.SetResult(results).Get(endpoint)

	if err != nil {
		return nil, err
	}

	if resp.IsError() {
		return nil, fmt.Errorf("%s: %s", resp.Status(), resp.String())
	}

	return resp.Result(), nil
}

func NewClient(baseURL string, username string, password string) *Client {
	return &Client{
		BaseURL:         baseURL,
		Username:        username,
		Password:        password,
		TLSClientConfig: &tls.Config{},
	}
}
