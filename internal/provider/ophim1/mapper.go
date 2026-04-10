package ophim1

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/SonDuon/BACKEND_GOLANG_MUTI/internal/provider"
)

type mapper struct{}

func newMapper() *mapper { return &mapper{} }

// 🔄 Map search result → DTO chuẩn (dành cho API Search)
func (m *mapper) toSearchResultDTO(item searchItem, source string) *provider.MovieDTO {
	// Lấy rating từ TMDB/IMDB nếu có
	var rating float32
	if item.TMDB.VoteAverage > 0 {
		rating = float32(item.TMDB.VoteAverage)
	} else if item.IMDB.VoteAverage > 0 {
		rating = float32(item.IMDB.VoteAverage)
	}

	return &provider.MovieDTO{
		Source:        source,
		ExternalID:    item.Slug,
		Title:         item.Name,
		OriginalTitle: item.OriginName,
		Slug:          item.Slug,
		Overview:      "", // Search API không có overview
		PosterURL:     joinImageURL(item.PosterURL),
		BackdropURL:   "",
		Type:          mapType(item.Type),
		Status:        mapStatus(item.EpisodeCurrent),
		ReleaseYear:   item.Year,
		Rating:        rating,
		Genres:        m.extractGenres(item.Category),
		TotalEpisodes: parseEpisodeCount(item.EpisodeCurrent),
	}
}

// 🔄 Map danh sách search result → DTO chuẩn
func (m *mapper) toMovieDTOs(items []searchItem, source string) []provider.MovieDTO {
	res := make([]provider.MovieDTO, 0, len(items))
	for _, item := range items {
		if item.Slug == "" {
			continue
		}
		// ✅ Tái dùng toSearchResultDTO nhờ bridge function
		res = append(res, *m.toSearchResultDTO(item, source))
	}
	return res
}

// 🔄 Map chi tiết phim → DTO chuẩn
func (m *mapper) toMovieDTO(raw *detailResponse, source string) *provider.MovieDTO {
	item := raw.Data.Item 

	// Lấy rating từ TMDB hoặc IMDB
	var rating float32
	if item.TMDB.VoteAverage > 0 {
		rating = float32(item.TMDB.VoteAverage)
	} else if item.IMDB.VoteAverage > 0 {
		rating = float32(item.IMDB.VoteAverage)
	}

	return &provider.MovieDTO{
		Source:        source,
		ExternalID:    item.Slug,
		Title:         item.Name,
		OriginalTitle: item.OriginName,
		Slug:          item.Slug,
		Overview:      stripHTML(item.Content), // ✅ Remove HTML tags từ content
		PosterURL:     joinImageURL(item.PosterURL),
		BackdropURL:   joinImageURL(item.BackdropURL),
		Type:          mapType(item.Type),
		Status:        mapStatus(item.Status),
		ReleaseYear:   item.Year,
		Rating:        rating,
		Genres:        m.extractGenres(item.Category),
		TotalEpisodes: parseEpisodeCount(item.EpisodeCurrent),
	}
}

// stripHTML: Remove HTML tags từ content (overview có HTML)
func stripHTML(html string) string {
	re := regexp.MustCompile(`<[^>]*>`)
	return strings.TrimSpace(re.ReplaceAllString(html, ""))
}

// 🎬 Map link xem → VideoSource chuẩn
func (m *mapper) toVideoSources(raw *detailResponse, episodeSlug, source string) []provider.VideoSource {
	var sources []provider.VideoSource

	// ✅ Detail API: episodes nằm trong raw.Data.Item.Episodes
	episodes := raw.Data.Item.Episodes

	// Filter đúng tập nếu có episodeSlug
	if episodeSlug != "" {
		for _, ep := range episodes {
			// Check bằng server_name hoặc slug của episode
			if ep.ServerName == episodeSlug {
				episodes = []episodeItem{ep}
				break
			}
		}
	}

	// Parse từng episode
	for _, ep := range episodes {
		// ✅ server_data là array of object, không phải map
		for _, srv := range ep.ServerData {
			// Ưu tiên link_m3u8 (HLS), fallback sang link_embed
			videoURL := srv.LinkM3u8
			videoType := "hls"

			if videoURL == "" {
				videoURL = srv.LinkEmbed
				videoType = "embed"
			}

			if videoURL == "" {
				continue
			}

			sources = append(sources, provider.VideoSource{
				ID:        provider.Coalesce(srv.Slug, srv.Name),
				Label:     ep.ServerName, // "Vietsub #1"
				Quality:   detectQuality(videoURL),
				URL:       videoURL,
				Type:      videoType,
				Server:    ep.ServerName,
				IsDefault: ep.ServerName == "Vietsub #1" || ep.ServerName == "Server 1",
			})
		}
	}

	return sources
}

func (m *mapper) extractGenres(cats []categoryItem) []string {
	g := make([]string, 0, len(cats))
	for _, c := range cats {
		if c.Name != "" {
			g = append(g, c.Name)
		}
	}
	return g
}

// 🛠️ Helper mappers
func mapType(t string) string {
	t = strings.ToLower(t)
	if t == "phimbo" || t == "series" || t == "tv" {
		return "series"
	}
	return "movie"
}

// parseDuration: Chuyển "138 phút" → 8280 (giây)
func parseDuration(timeStr string) int {
	if timeStr == "" || timeStr == "Đang cập nhật" {
		return 0
	}

	// Extract số từ chuỗi "138 phút" hoặc "1H39M8S"
	var hours, minutes, seconds int

	// Case 1: "138 phút"
	if strings.Contains(timeStr, "phút") {
		fmt.Sscanf(timeStr, "%d", &minutes)
		return minutes * 60
	}

	// Case 2: "1H39M8S"
	if strings.Contains(timeStr, "H") {
		fmt.Sscanf(timeStr, "%dH%dM%dS", &hours, &minutes, &seconds)
		return hours*3600 + minutes*60 + seconds
	}

	// Case 3: Chỉ có số
	fmt.Sscanf(timeStr, "%d", &minutes)
	return minutes * 60
}

// Helper: Ghép URL ảnh với CDN domain
func joinImageURL(imagePath string) string {
	if imagePath == "" {
		return ""
	}
	// Ophim1 dùng CDN: https://img.ophim.live
	if strings.HasPrefix(imagePath, "http") {
		return imagePath
	}
	return "https://img.ophim.live/" + imagePath
}

// Helper: Parse số tập từ chuỗi như "Full", "Hoàn Tất (10/10)", "Đang cập nhật"
func parseEpisodeCount(episodeCurrent string) *int {
	if episodeCurrent == "" || episodeCurrent == "Full" || episodeCurrent == "Đang cập nhật" {
		return nil
	}

	// Extract số từ chuỗi "Hoàn Tất (10/10)" → 10
	// Implement regex hoặc string parsing đơn giản
	return nil // Tạm thời return nil
}

func mapStatus(s string) string {
	s = strings.ToLower(s)
	if s == "ongoing" || s == "đang chiếu" || s == "airing" {
		return "ongoing"
	}
	return "completed"
}

func mapSortToEndpoint(sort string) string {
	switch sort {
	case "newest":
		return "danh-sach/phim-moi-cap-nhat"
	case "popular":
		return "danh-sach/phim-hot"
	case "top":
		return "danh-sach/top-imdb"
	default:
		return "danh-sach/phim-moi-cap-nhat"
	}
}

func detectQuality(u string) string {
	if strings.Contains(u, "1080") || strings.Contains(u, "fhd") {
		return "1080p"
	}
	if strings.Contains(u, "720") || strings.Contains(u, "hd") {
		return "720p"
	}
	if strings.Contains(u, "4k") || strings.Contains(u, "uhd") {
		return "4K"
	}
	return "auto"
}

func detectType(u string) string {
	if strings.Contains(u, ".m3u8") {
		return "hls"
	}
	if strings.Contains(u, ".mp4") {
		return "mp4"
	}
	return "embed"
}
