package provider

import "context"

// ─────────────────────────────────────
// 🎯 MovieProvider Interface (Contract)
// ─────────────────────────────────────
// Mọi nguồn phim (ophim1, self-hosted, ...) PHẢI implement interface này

type MovieProvider interface {
	// Name: Trả về định danh duy nhất của provider
	// Dùng để logging, config, và chọn provider runtime
	Name() string

	// Priority: Mức độ ưu tiên (số càng lớn càng được ưu tiên)
	// Ví dụ: self-hosted = 10, ophim1 = 5, nguonc = 1
	Priority() int

	// IsAvailable: Health check - provider này có đang hoạt động không?
	// Gọi nhanh, timeout ngắn (< 3s) để không block request
	IsAvailable(ctx context.Context) bool

	// ─────────────────────────────────
	// 📚 Metadata Operations (Sync về DB)
	// ─────────────────────────────────

	// Search: Tìm phim theo keyword (dùng cho Admin panel + User search fallback)
	Search(ctx context.Context, params *SearchParams) (*SearchResult, error)

	// GetByExternalID: Lấy chi tiết phim theo ID gốc từ provider
	// Dùng khi sync metadata hoặc refresh thông tin phim
	GetByExternalID(ctx context.Context, externalID string) (*MovieDTO, error)

	// GetList: Lấy danh sách phim mới nhất/top/rating cao...
	// Dùng cho background sync job
	GetList(ctx context.Context, page, limit int, sortBy string) (*SearchResult, error)

	// ─────────────────────────────────
	// 🎬 Streaming Operations (On-demand)
	// ─────────────────────────────────

	// GetStreamingLinks: Lấy link xem phim THỰC
	// ⚠️ Chỉ gọi khi user bấm "Play" - KHÔNG lưu kết quả vào DB
	// Trả về multiple sources + subtitles + expiry time
	GetStreamingLinks(ctx context.Context, movieExternalID string, episodeExternalID string) (*StreamingDTO, error)

	// RefreshStreamingLinks: Force refresh link (dùng khi link cũ expired/error)
	RefreshStreamingLinks(ctx context.Context, movieExternalID string, episodeExternalID string) (*StreamingDTO, error)
}
