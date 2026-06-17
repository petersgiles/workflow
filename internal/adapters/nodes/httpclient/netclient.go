package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
)

type NetClient struct {
	c *http.Client
}

func NewNetClient() *NetClient { return &NetClient{c: &http.Client{}} }

func (n *NetClient) PostJSON(ctx context.Context, url string, body any) (int, []byte, error) {
	b, _ := json.Marshal(body)
	req, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	resp, err := n.c.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()
	bb, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, bb, nil
}
