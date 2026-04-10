package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// ─────────────────────────────────────
// 🛠️ Shared Utilities for Providers
// ─────────────────────────────────────

// DefaultHTTPClient: HTTP client config tối ưu cho API calls
func DefaultHTTPClient(timeout time.Duration) *http.Client {
	return &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
			TLSHandshakeTimeout: 10 * time.Second,
		},
	}
}

// FetchJSON: Helper fetch + parse JSON từ URL (giảm boilerplate code)
func FetchJSON[T any](ctx context.Context, client *http.Client, url string, result *T) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	// Default headers
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "MovieApp-Backend/1.0")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("%w: status=%d, body=%s", ErrAPIResponse, resp.StatusCode, string(body))
	}

	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return fmt.Errorf("decode JSON: %w", err)
	}

	return nil
}

// Ptr: Helper tạo pointer từ value (dùng cho optional fields trong DTO)
func Ptr[T any](v T) *T {
	return &v
}

// Coalesce: Trả về giá trị đầu tiên khác zero/empty (dùng cho fallback values)
func Coalesce[T comparable](values ...T) T {
	var zero T
	for _, v := range values {
		if v != zero {
			return v
		}
	}
	return zero
}

// SanitizeSlug: Chuẩn hóa slug cho consistent across providers
func SanitizeSlug(slug string) string {
	// Implement: lowercase, remove special chars, replace spaces with dash
	// Ví dụ: "Ma Trận 2" → "ma-tran-2"
	return slug // TODO: implement actual sanitization
}
