package provider

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"
)

// ─────────────────────────────────────
// 🏢 Provider Manager
// ─────────────────────────────────────
// Quản lý lifecycle, selection, và fallback cho multiple providers

type Manager struct {
	providers []MovieProvider
	cache     sync.Map // map[string]*cachedStreaming
	cacheTTL  time.Duration
	mu        sync.RWMutex
}

type cachedStreaming struct {
	data      *StreamingDTO
	expiresAt time.Time
}

// NewManager: Khởi tạo manager với list providers
func NewManager(providers []MovieProvider, cacheTTL time.Duration) *Manager {
	// Sort by priority (cao trước)
	sort.Slice(providers, func(i, j int) bool {
		return providers[i].Priority() > providers[j].Priority()
	})

	return &Manager{
		providers: providers,
		cacheTTL:  cacheTTL,
	}
}

// GetProvider: Lấy provider theo tên (dùng cho admin import)
func (m *Manager) GetProvider(name string) (MovieProvider, error) {
	for _, p := range m.providers {
		if p.Name() == name {
			return p, nil
		}
	}
	return nil, fmt.Errorf("%w: %s", ErrProviderNotFound, name)
}

// GetBestProvider: Lấy provider khả dụng có priority cao nhất (dùng cho user play)
func (m *Manager) GetBestProvider(ctx context.Context) (MovieProvider, error) {
	for _, p := range m.providers {
		if p.IsAvailable(ctx) {
			return p, nil
		}
	}
	return nil, ErrProviderUnavailable
}

// GetStreamingWithFallback: Get link với fallback chain + cache
func (m *Manager) GetStreamingWithFallback(
	ctx context.Context,
	movieExternalID, episodeExternalID, preferredSource string,
) (*StreamingDTO, error) {

	// 1. Try cache first
	cacheKey := fmt.Sprintf("%s:%s", movieExternalID, episodeExternalID)
	if cached := m.getCached(cacheKey); cached != nil {
		return cached, nil
	}

	// 2. Try preferred source first (nếu user/admin chọn)
	if preferredSource != "" {
		if provider, err := m.GetProvider(preferredSource); err == nil {
			if resp, err := provider.GetStreamingLinks(ctx, movieExternalID, episodeExternalID); err == nil {
				m.setCache(cacheKey, resp)
				return resp, nil
			}
		}
	}

	// 3. Fallback chain theo priority
	var lastErr error
	for _, provider := range m.providers {
		if !provider.IsAvailable(ctx) {
			continue
		}

		resp, err := provider.GetStreamingLinks(ctx, movieExternalID, episodeExternalID)
		if err != nil {
			lastErr = err
			continue
		}

		// Success! Cache và return
		resp.Metadata = map[string]any{"actual_source": provider.Name()}
		m.setCache(cacheKey, resp)
		return resp, nil
	}

	return nil, fmt.Errorf("all providers failed: %w", lastErr)
}

// Cache helpers
func (m *Manager) getCached(key string) *StreamingDTO {
	if val, ok := m.cache.Load(key); ok {
		if cached, ok := val.(*cachedStreaming); ok {
			if time.Now().Before(cached.expiresAt) {
				return cached.data
			}
			m.cache.Delete(key) // expired
		}
	}
	return nil
}

func (m *Manager) setCache(key string, data *StreamingDTO) {
	m.cache.Store(key, &cachedStreaming{
		data:      data,
		expiresAt: time.Now().Add(m.cacheTTL),
	})
}

// HealthCheckAll: Dùng cho admin/monitoring endpoint
func (m *Manager) HealthCheckAll(ctx context.Context) map[string]bool {
	results := make(map[string]bool)
	for _, p := range m.providers {
		results[p.Name()] = p.IsAvailable(ctx)
	}
	return results
}

// ListProviders: Trả về danh sách provider đã đăng ký (cho debug/config)
func (m *Manager) ListProviders() []string {
	names := make([]string, len(m.providers))
	for i, p := range m.providers {
		names[i] = p.Name()
	}
	return names
}
