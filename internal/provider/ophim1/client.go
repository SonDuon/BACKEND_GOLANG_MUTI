package ophim1

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/SonDuon/BACKEND_GOLANG_MUTI/internal/provider"
)

// ✅ Singleton HTTP Client - Reuse connection pool
var httpClient *http.Client

func init() {
	httpClient = &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 100, // Tăng từ 10 → 100 cho high concurrency
			IdleConnTimeout:     90 * time.Second,
			TLSHandshakeTimeout: 10 * time.Second,
			DisableKeepAlives:   false, // Giữ kết nối alive để reuse
			ForceAttemptHTTP2:   true,  // Sử dụng HTTP/2 nếu server hỗ trợ
		},
	}
}

type client struct {
	baseURL string
}

func newClient(baseURL string, timeout time.Duration) *client {
	return &client{
		baseURL: strings.TrimRight(baseURL, "/"),
	}
}

// ✅ Hàm search đã fix
func (c *client) search(ctx context.Context, keyword string, page, limit int) (*searchResponse, error) {
	// 1. URL Encode keyword để xử lý khoảng trắng & ký tự đặc biệt (VD: "ma trận" → "ma%20tr%E1%BA%ADn")
	encodedKeyword := url.QueryEscape(keyword)

	// 2. Build URL chuẩn
	searchURL := fmt.Sprintf("%s/v1/api/tim-kiem?keyword=%s&page=%d&limit=%d",
		c.baseURL, encodedKeyword, page, limit)

	var resp searchResponse
	if err := provider.FetchJSON(ctx, httpClient, searchURL, &resp); err != nil {
		return nil, fmt.Errorf("search ophim1: %w", err)
	}

	return &resp, nil
}

func (c *client) ping(ctx context.Context) bool {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodHead, c.baseURL, nil)
	if err != nil {
		return false
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode < 500
}

func (c *client) detail(ctx context.Context, slug string) (*detailResponse, error) {
	url := fmt.Sprintf("%s/v1/api/phim/%s",
		c.baseURL,
		slug,
	)

	var resp detailResponse

	if err := provider.FetchJSON(ctx, httpClient, url, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *client) list(ctx context.Context, endpoint string, page, limit int) (*searchResponse, error) {
	url := fmt.Sprintf("%s/v1/api/%s?page=%d&limit=%d",
		c.baseURL, endpoint, page, limit)

	var resp searchResponse

	if err := provider.FetchJSON(ctx, httpClient, url, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
