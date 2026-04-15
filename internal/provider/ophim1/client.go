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

type client struct {
	baseURL string
	http    *http.Client
}

func newClient(baseURL string, timeout time.Duration) *client {
	return &client{
		// ✅ Loại bỏ dấu "/" thừa ở cuối URL nếu có
		baseURL: strings.TrimRight(baseURL, "/"),
		http: &http.Client{
			Timeout: timeout,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
			},
		},
	}
}

// ✅ Hàm search đã fix
func (c *client) search(ctx context.Context, keyword string, page, limit int) (*searchResponse, error) {
	// 1. URL Encode keyword để xử lý khoảng trắng & ký tự đặc biệt (VD: "ma trận" → "ma%20tr%E1%BA%ADn")
	encodedKeyword := url.QueryEscape(keyword)

	// 2. Build URL chuẩn
	searchURL := fmt.Sprintf("%s/v1/api/tim-kiem?keyword=%s&page=%d&limit=%d",
		c.baseURL, encodedKeyword, page, limit)

	// 🐛 Debug: In ra URL thực tế đang gọi (xoá sau khi OK)
	fmt.Printf("🔍 Calling Ophim1 Search: %s\n", searchURL)

	var resp searchResponse
	if err := provider.FetchJSON(ctx, c.http, searchURL, &resp); err != nil {
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

	resp, err := c.http.Do(req)
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
