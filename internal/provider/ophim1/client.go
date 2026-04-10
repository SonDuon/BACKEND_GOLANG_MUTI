package ophim1

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/SonDuon/BACKEND_GOLANG_MUTI/internal/provider"
)

type client struct {
	baseURL string
	http    *http.Client
}

func newClient(baseURL string, timeout time.Duration) *client {
	return &client{
		baseURL: baseURL,
		http:    provider.DefaultHTTPClient(timeout),
	}
}

func (c *client) ping(ctx context.Context) bool {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodHead, c.baseURL, nil)
	if err != nil {
		return false
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode < 500
}

func (c *client) search(ctx context.Context, keyword string, page, limit int) (*searchResponse, error) {
	url := fmt.Sprintf("%s/v1/api/tim-kiem?keyword=%s",
		c.baseURL,
		keyword,
	)

	// Ophim1 có thể không support pagination trong search, ignore page/limit nếu cần

	var resp searchResponse
	if err := provider.FetchJSON(ctx, c.http, url, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *client) detail(ctx context.Context, slug string) (*detailResponse, error) {
	url := fmt.Sprintf("%s/v1/api/phim/%s",
		c.baseURL,
		slug,
	)

	var resp detailResponse

	if err := provider.FetchJSON(ctx, c.http, url, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *client) list(ctx context.Context, endpoint string, page, limit int) (*searchResponse, error) {
	url := fmt.Sprintf("%s/v1/api/%s?page=%d&limit=%d",
		c.baseURL, endpoint, page, limit)

	var resp searchResponse

	if err := provider.FetchJSON(ctx, c.http, url, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
