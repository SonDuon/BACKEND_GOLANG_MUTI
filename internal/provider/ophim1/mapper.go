package ophim1

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/SonDuon/BACKEND_GOLANG_MUTI/internal/provider"
)

type mapper struct{}

func newMapper() *mapper { return &mapper{} }

// toSearchResultDTO: Map Search API (nhiều phim, ít thông tin) -> DTO
func (m *mapper) toSearchResultDTO(item searchItem, source string) *provider.MovieDTO {
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
		Overview:      "", // Search API không trả về mô tả chi tiết
		ThumbURL:      joinImageURL(item.ThumbURL),
		PosterURL:     joinImageURL(item.PosterURL),
		Type:          mapType(item.Type),
		Status:        mapSearchStatus(item.Status, item.EpisodeCurrent),
		ReleaseYear:   item.Year,
		Rating:        rating,
		Genres:        m.extractGenres(item.Category),
		TotalEpisodes: parseEpisodeCount(item.EpisodeCurrent),
	}
}

// toMovieDTOs: Map danh sách search result -> List DTO
func (m *mapper) toMovieDTOs(items []searchItem, source string) []provider.MovieDTO {
	res := make([]provider.MovieDTO, 0, len(items))
	for _, i := range items {
		if i.Slug == "" {
			continue
		}
		res = append(res, *m.toSearchResultDTO(i, source))
	}
	return res
}

// toMovieDTO: Map Detail API (1 phim, đầy đủ thông tin) -> DTO
func (m *mapper) toMovieDTO(raw *detailResponse, source string) *provider.MovieDTO {
	// ⚠️ QUAN TRỌNG: Detail API trả về data nằm trong raw.Data.Item
	item := raw.Data.Item

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
		Overview:      stripHTML(item.Content),
		ThumbURL:      joinImageURL(item.ThumbURL),
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

// toVideoSources: Map Detail API -> Danh sách link phim

func (m *mapper) toVideoSources(raw *detailResponse, episodeSlug, source string) []provider.VideoSource {
	var sources []provider.VideoSource

	episodes := raw.Data.Item.Episodes

	// 🔑 Normalize episodeSlug: "tap-1" → "1", "episode-2" → "2"
	normalizedSlug := normalizeEpisodeSlug(episodeSlug)

	// Ophim1 structure: mỗi "episode" thực ra là SERVER, server_data chứa tất cả tập
	for _, ep := range episodes {
		for _, srv := range ep.ServerData {
			// 🎯 Filter đúng tập nếu có episodeSlug
			if episodeSlug != "" {
				// Match với: exact slug, exact name, hoặc normalized version
				if srv.Slug != episodeSlug && srv.Name != episodeSlug &&
					srv.Slug != normalizedSlug && srv.Name != normalizedSlug {
					continue // Skip tập không khớp
				}
			}

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
				Label:     fmt.Sprintf("%s - Tập %s", ep.ServerName, srv.Name),
				Quality:   detectQuality(videoURL),
				URL:       videoURL,
				Type:      videoType,
				Server:    ep.ServerName,
				IsDefault: srv.Slug == "1" || srv.Name == "1", // Default tập 1
			})
		}
	}

	return sources
}

// normalizeEpisodeSlug: Chuyển "tap-1" → "1", "episode-2" → "2"
func normalizeEpisodeSlug(slug string) string {
	if slug == "" {
		return ""
	}
	s := strings.ToLower(slug)
	// Remove common prefixes
	s = strings.TrimPrefix(s, "tap-")
	s = strings.TrimPrefix(s, "tập-")
	s = strings.TrimPrefix(s, "episode-")
	s = strings.TrimPrefix(s, "ep-")
	s = strings.TrimPrefix(s, "tập")
	s = strings.TrimSpace(s)
	return s
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

// ─────────────────────────────────────
// 🛠️ Helper Functions
// ─────────────────────────────────────

func mapType(t string) string {
	// Lưu type nguyên bản từ API
	return strings.ToLower(t)
}

func mapStatus(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	if s == "" {
		return "ongoing"
	}

	if strings.Contains(s, "ongoing") || strings.Contains(s, "đang") || strings.Contains(s, "airing") {
		return "ongoing"
	}

	if strings.Contains(s, "completed") || strings.Contains(s, "hoàn") || strings.Contains(s, "full") {
		return "completed"
	}

	return "ongoing"
}

func mapSearchStatus(status string, episodeCurrent string) string {
	if strings.TrimSpace(status) != "" {
		return mapStatus(status)
	}
	return mapStatus(episodeCurrent)
}

func joinImageURL(imagePath string) string {
	if imagePath == "" {
		return ""
	}
	if strings.HasPrefix(imagePath, "http") {
		return imagePath
	}
	// CDN của Ophim1
	return "https://img.ophim.live/uploads/movies/" + imagePath
}

func parseEpisodeCount(episodeCurrent string) *int {
	if episodeCurrent == "" || strings.ToLower(episodeCurrent) == "full" || strings.Contains(episodeCurrent, "Đang cập nhật") {
		return nil
	}
	// Xử lý chuỗi "Hoàn Tất (10/10)" -> lấy số 10
	re := regexp.MustCompile(`\d+`)
	matches := re.FindAllString(episodeCurrent, -1)
	if len(matches) > 0 {
		val, err := strconv.Atoi(matches[0])
		if err == nil {
			return &val
		}
	}
	return nil
}

// parseDuration: Chuyển "138 phút" hoặc "1H39M8S" -> giây (int)
func parseDuration(duration string) int {
	if duration == "" || strings.Contains(duration, "Cập nhật") {
		return 0
	}

	// Xử lý "1H39M8S"
	if strings.Contains(duration, "H") {
		re := regexp.MustCompile(`(\d+)H(\d+)M(\d+)S`)
		matches := re.FindStringSubmatch(duration)
		if len(matches) == 4 {
			h, _ := strconv.Atoi(matches[1])
			m, _ := strconv.Atoi(matches[2])
			s, _ := strconv.Atoi(matches[3])
			return h*3600 + m*60 + s
		}
	}

	// Xử lý "138 phút"
	re := regexp.MustCompile(`(\d+)`)
	matches := re.FindStringSubmatch(duration)
	if len(matches) == 2 {
		m, _ := strconv.Atoi(matches[1])
		return m * 60
	}

	return 0
}

// stripHTML: Loại bỏ thẻ HTML khỏi mô tả phim
func stripHTML(html string) string {
	re := regexp.MustCompile(`<[^>]*>`)
	return strings.TrimSpace(re.ReplaceAllString(html, ""))
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
func sortToEndpoint(sort string) string {
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
