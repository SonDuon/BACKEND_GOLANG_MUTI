package ophim1

import (
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
		PosterURL:     joinImageURL(item.PosterURL),
		Type:          mapType(item.Type),
		Status:        mapStatus(item.EpisodeCurrent),
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
		Overview:      stripHTML(item.Content), // Loại bỏ thẻ <p>
		PosterURL:     joinImageURL(item.PosterURL),
		BackdropURL:   joinImageURL(item.BackdropURL), // Nếu có
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
	// ⚠️ Lấy episodes từ item
	episodes := raw.Data.Item.Episodes

	// Filter đúng tập nếu có episodeSlug (thường là "Full" hoặc "tap-1")
	if episodeSlug != "" {
		for _, ep := range episodes {
			if ep.ServerName == episodeSlug {
				episodes = []episodeItem{ep}
				break
			}
		}
	}

	for _, ep := range episodes {
		// server_data là mảng các server, không phải map
		for _, srvData := range ep.ServerData {
			// Ưu tiên link_m3u8 (HLS), fallback sang link_embed
			videoURL := srvData.LinkM3u8
			videoType := "hls"

			if videoURL == "" {
				videoURL = srvData.LinkEmbed
				videoType = "embed"
			}

			if videoURL == "" {
				continue
			}

			sources = append(sources, provider.VideoSource{
				ID:        provider.Coalesce(srvData.Slug, srvData.Name),
				Label:     ep.ServerName, // Ví dụ: "Vietsub #1"
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

// ─────────────────────────────────────
// 🛠️ Helper Functions
// ─────────────────────────────────────

func mapType(t string) string {
	t = strings.ToLower(t)
	if t == "phimbo" || t == "series" || t == "tv" {
		return "series"
	}
	return "movie"
}

func mapStatus(s string) string {
	s = strings.ToLower(s)
	if strings.Contains(s, "ongoing") || strings.Contains(s, "đang") || strings.Contains(s, "airing") {
		return "ongoing"
	}
	return "completed"
}

func joinImageURL(imagePath string) string {
	if imagePath == "" {
		return ""
	}
	if strings.HasPrefix(imagePath, "http") {
		return imagePath
	}
	// CDN của Ophim1
	return "https://img.ophim.live/" + imagePath
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
