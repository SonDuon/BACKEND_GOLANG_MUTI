package provider

// ─────────────────────────────────────
// 📦 Data Transfer Objects (DTOs)
// ─────────────────────────────────────
// Các struct này dùng để giao tiếp giữa các layer
// KHÔNG chứa logic business, chỉ chứa dữ liệu

// MovieDTO: Cấu trúc phim CHUẨN hóa từ mọi nguồn
type MovieDTO struct {
	Source     string `json:"source"`      // "ophim1", "self", "nguonc"
	ExternalID string `json:"external_id"` // ID gốc từ API để gọi lại

	// 🎬 Hiển thị
	Title         string `json:"title"`
	OriginalTitle string `json:"original_title,omitempty"`
	Slug          string `json:"slug,omitempty"`     // URL-friendly name
	Overview      string `json:"overview,omitempty"` // Mô tả ngắn
	PosterURL     string `json:"poster_url,omitempty"`
	BackdropURL   string `json:"backdrop_url,omitempty"`
	TrailerURL    string `json:"trailer_url,omitempty"`

	// 📊 Metadata
	Type        string  `json:"type"`             // "movie" | "series"
	Status      string  `json:"status,omitempty"` // "ongoing" | "completed"
	ReleaseYear int     `json:"release_year,omitempty"`
	Rating      float32 `json:"rating,omitempty"`
	VoteCount   int     `json:"vote_count,omitempty"`
	Runtime     int     `json:"runtime,omitempty"` // Phút

	// 🏷️ Phân loại
	Genres    []string `json:"genres,omitempty"`
	Countries []string `json:"countries,omitempty"`

	// 📺 Series info
	SeasonsCount  *int `json:"seasons_count,omitempty"`
	TotalEpisodes *int `json:"total_episodes,omitempty"`

	// ⏰ Timestamp (optional, dùng cho sync)
	LastSyncedAt int64 `json:"last_synced_at,omitempty"`
}

// StreamingDTO: Cấu trúc trả về khi user bấm "Play"
type StreamingDTO struct {
	MovieID   string         `json:"movie_id"`
	EpisodeID string         `json:"episode_id,omitempty"`
	Title     string         `json:"title"`
	Sources   []VideoSource  `json:"sources"`
	Subtitles []Subtitle     `json:"subtitles,omitempty"`
	Thumbnail string         `json:"thumbnail,omitempty"`
	Duration  int            `json:"duration,omitempty"` // seconds
	ExpiresAt int64          `json:"expires_at"`         // URL signing expiry
	Metadata  map[string]any `json:"metadata,omitempty"` // Extra info từ provider
}

// VideoSource: Một nguồn video cụ thể
type VideoSource struct {
	ID        string `json:"id"`
	Label     string `json:"label"`      // "Server VIP", "Google Drive"
	Quality   string `json:"quality"`    // "1080p", "720p", "4K", "auto"
	URL       string `json:"url"`        // m3u8, mp4, hoặc embed URL
	Type      string `json:"type"`       // "hls", "mp4", "embed", "dash"
	Server    string `json:"server"`     // "server1", "cdn", "google_drive"
	IsDefault bool   `json:"is_default"` // Default source để auto-play
}

// Subtitle: Track phụ đề
type Subtitle struct {
	Language  string `json:"language"` // "vi", "en", "ja"
	Label     string `json:"label"`    // "Tiếng Việt", "English"
	URL       string `json:"url"`
	IsDefault bool   `json:"is_default"`
}

// SearchParams: Params cho hàm Search (dùng chung cho mọi provider)
type SearchParams struct {
	Keyword string
	Page    int
	Limit   int
	Genre   string
	Year    int
	Type    string // "movie" | "series"
}

// SearchResult: Wrapper cho kết quả search có pagination
type SearchResult struct {
	Items   []MovieDTO `json:"items"`
	Total   int64      `json:"total"`
	Page    int        `json:"page"`
	Limit   int        `json:"limit"`
	HasMore bool       `json:"has_more"`
}
