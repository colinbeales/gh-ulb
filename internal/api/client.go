package api

import (
	"bytes"
	"encoding/json"
	"io"
	"os"

	"github.com/cli/go-gh/v2/pkg/api"
)

type Client struct {
	rest *api.RESTClient
}

func NewClient() (*Client, error) {
	opts := api.ClientOptions{
		Headers: map[string]string{
			"X-GitHub-Api-Version": "2022-11-28",
		},
	}
	if host := os.Getenv("GH_HOST"); host != "" {
		opts.Host = host
	}
	rest, err := api.NewRESTClient(opts)
	if err != nil {
		return nil, err
	}
	return &Client{rest: rest}, nil
}

func (c *Client) Get(path string, resp interface{}) error {
	return c.rest.Get(path, resp)
}

func (c *Client) Post(path string, body interface{}, resp interface{}) error {
	r, err := toReader(body)
	if err != nil {
		return err
	}
	return c.rest.Post(path, r, resp)
}

func (c *Client) Patch(path string, body interface{}, resp interface{}) error {
	r, err := toReader(body)
	if err != nil {
		return err
	}
	return c.rest.Patch(path, r, resp)
}

func (c *Client) Delete(path string, resp interface{}) error {
	return c.rest.Delete(path, resp)
}

func toReader(body interface{}) (io.Reader, error) {
	b, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(b), nil
}
