package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/SonDuon/BACKEND_GOLANG_MUTI/internal/database"
	"github.com/SonDuon/BACKEND_GOLANG_MUTI/internal/models"
	"github.com/SonDuon/BACKEND_GOLANG_MUTI/internal/provider"
	"github.com/SonDuon/BACKEND_GOLANG_MUTI/internal/repository"
	"gorm.io/gorm"
)

type MovieService struct {
	repo     repository.MovieRepository
	provider provider.MovieProvider
}

const detailSyncTTL = 12 * time.Hour

// SearchMovies: Tìm kiếm phim từ Ophim API
func (s *MovieService) SearchMovies(ctx context.Context, query string, page int, limit int) ([]models.Movie, int64, error) {
	query = strings.TrimSpace(query)

	// 1) Ưu tiên tìm trong DB trước
	movies, total, err := s.repo.Search(ctx, query, page, limit)
	if err == nil && len(movies) > 0 {
		return movies, total, nil
	}
	if err != nil {
		log.Printf("⚠️ DB search failed for '%s': %v. Falling back to provider", query, err)
	} else {
		log.Printf("🔍 Không tìm thấy '%s' trong DB. Đang fetch từ provider...", query)
	}

	// 2) DB chưa có dữ liệu phù hợp -> gọi provider
	searchResult, err := s.provider.Search(ctx, &provider.SearchParams{
		Keyword: query,
		Page:    page,
		Limit:   limit,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("provider search error: %w", err)
	}

	// 3) Convert DTOs -> Models và auto-import từng phim vào DB
	movies = make([]models.Movie, 0, len(searchResult.Items))
	for _, dto := range searchResult.Items {
		movieModel := s.convertDTOToModel(&dto)

		// Auto-import vào DB (không fail request nếu lỗi)
		if err := s.repo.Create(ctx, movieModel); err != nil {
			log.Printf("⚠️ Auto-import failed for '%s': %v", dto.Slug, err)
		}

		movies = append(movies, *movieModel)
	}

	return movies, searchResult.Total, nil
}

func (s *MovieService) ListMovies(ctx context.Context, page int, limit int, typeFilter string, statusFilter string) ([]models.Movie, int64, error) {
	return s.repo.List(ctx, page, limit, typeFilter, statusFilter)
}

func NewMovieService(repo repository.MovieRepository, provider provider.MovieProvider) *MovieService {
	return &MovieService{
		repo:     repo,
		provider: provider,
	}
}

// ==========================================
// 🎬 1. DETAIL (Lazy Import Pattern)
// ==========================================

// GetMovieDetail: Kiểm tra DB -> Nếu không có thì gọi API -> Lưu DB -> Trả về
func (s *MovieService) GetMovieDetail(ctx context.Context, slug string) (*models.Movie, error) {
	// 1️⃣ Tìm trong Database trước
	movie, err := s.repo.GetBySlug(ctx, slug)
	if err == nil {
		if shouldRefreshMovieDetail(movie) {
			externalID := movie.ExternalID
			if strings.TrimSpace(externalID) == "" {
				externalID = slug
			}

			go s.backgroundSyncMovie(context.Background(), externalID)
		}

		return movie, nil
	}

	// Kiểm tra lỗi "Không tìm thấy bản ghi"
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("database error: %w", err)
	}

	// 2️⃣ Không có trong DB -> fetch đồng bộ từ provider và import
	return s.fetchAndImportFromProvider(ctx, slug)
}

func shouldRefreshMovieDetail(movie *models.Movie) bool {
	if movie == nil {
		return false
	}

	status := strings.ToLower(strings.TrimSpace(movie.Status))
	if status == "completed" {
		return false
	}

	if movie.LastSyncedAt.IsZero() {
		return true
	}

	return time.Since(movie.LastSyncedAt) >= detailSyncTTL
}

func (s *MovieService) fetchAndImportFromProvider(ctx context.Context, slug string) (*models.Movie, error) {
	log.Printf("🔍 Movie '%s' not found in DB. Fetching from provider...", slug)

	dto, err := s.provider.GetByExternalID(ctx, slug)
	if err != nil {
		return nil, fmt.Errorf("provider error: %w", err)
	}

	movieModel := s.convertDTOToModel(dto)
	if err := s.repo.Create(ctx, movieModel); err != nil {
		log.Printf("⚠️ Auto-import failed for '%s': %v", slug, err)
	}

	return movieModel, nil
}

func (s *MovieService) backgroundSyncMovie(ctx context.Context, externalID string) {
	log.Printf("🔄 Background sync started for '%s'", externalID)

	syncCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	dto, err := s.provider.GetByExternalID(syncCtx, externalID)
	if err != nil {
		log.Printf("⚠️ Background sync failed for '%s': %v", externalID, err)
		return
	}

	if strings.TrimSpace(dto.Status) == "" {
		log.Printf("ℹ️ Background sync skipped update for '%s' because provider status is empty", externalID)
		return
	}

	movieModel := s.convertDTOToModel(dto)
	if err := s.repo.Create(syncCtx, movieModel); err != nil {
		log.Printf("⚠️ Background sync persist failed for '%s': %v", externalID, err)
		return
	}

	log.Printf("✅ Background sync completed for '%s'", externalID)
}

// ==========================================
// 📋 2. LIST (Phân trang)
// ==========================================

// GetMovieList: Lấy danh sách phim từ Database
func (s *MovieService) GetMovieList(ctx context.Context, page, limit int) ([]*models.Movie, int64, error) {
	var movies []*models.Movie
	var total int64

	// Dùng database.DB trực tiếp cho nhanh trong giai đoạn test
	// (Production nên đưa logic này vào Repository)
	query := database.DB.Model(&models.Movie{})

	// Đếm tổng số phim
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Lấy dữ liệu với phân trang (Page 1 -> Offset 0)
	offset := (page - 1) * limit
	if err := query.
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&movies).Error; err != nil {
		return nil, 0, err
	}

	return movies, total, nil
}

// ==========================================
// ▶️ 3. WATCH (Lấy link xem)
// ==========================================

// GetWatchLinks: Trả về link streaming cho User
func (s *MovieService) GetWatchLinks(ctx context.Context, slug string, episodeSlug string) (*provider.StreamingDTO, error) {
	// 1️⃣ Lấy thông tin phim (Tận dụng hàm Detail để đảm bảo phim đã có trong DB/API)
	movie, err := s.GetMovieDetail(ctx, slug)
	if err != nil {
		return nil, err
	}

	episodeSlug = strings.TrimSpace(episodeSlug)
	if strings.EqualFold(episodeSlug, "full") {
		episodeSlug = ""
	}

	// 3️⃣ Xử lý theo Source
	if movie.Source == "self" {
		// 🏠 TỰ HOST: Lấy link từ Database (MediaSources)
		if !isSeriesType(movie.Type) {
			for _, ep := range movie.Episodes {
				if len(ep.MediaSources) == 0 {
					continue
				}

				src := ep.MediaSources[0]
				return &provider.StreamingDTO{
					MovieID:   movie.ID.String(),
					EpisodeID: ep.ID.String(),
					Sources: []provider.VideoSource{{
						Label: src.ServerName,
						URL:   src.SourceKey,
						Type:  src.SourceType,
					}},
				}, nil
			}
		}

		// Series-like + episode rỗng: trả danh sách nguồn theo từng tập
		if episodeSlug == "" {
			allSources := make([]provider.VideoSource, 0)
			for _, ep := range movie.Episodes {
				for _, src := range ep.MediaSources {
					allSources = append(allSources, provider.VideoSource{
						ID:      ep.Slug,
						Label:   fmt.Sprintf("%s - %s", ep.Title, src.ServerName),
						URL:     src.SourceKey,
						Type:    src.SourceType,
						Quality: src.Quality,
						Server:  src.ServerName,
					})
				}
			}

			if len(allSources) == 0 {
				return nil, fmt.Errorf("không có danh sách tập hoặc nguồn phát")
			}

			return &provider.StreamingDTO{
				MovieID: movie.ID.String(),
				Title:   movie.Title,
				Sources: allSources,
			}, nil
		}

		for _, ep := range movie.Episodes {
			if episodeSlugMatch(ep.Slug, episodeSlug) {
				if len(ep.MediaSources) > 0 {
					src := ep.MediaSources[0] // Lấy server đầu tiên
					return &provider.StreamingDTO{
						MovieID:   movie.ID.String(),
						EpisodeID: ep.ID.String(),
						Sources: []provider.VideoSource{
							{
								Label: src.ServerName,
								URL:   src.SourceKey, // Link MP4/M3U8 lưu trong DB
								Type:  src.SourceType,
							},
						},
					}, nil
				}
			}
		}
		return nil, fmt.Errorf("không tìm thấy tập '%s' hoặc nguồn phát", episodeSlug)
	}

	// 🌐 API BÊN THỨ 3 (Ophim1): Gọi API lấy link tươi (Fresh Link)
	// episode rỗng => provider trả danh sách tập; episode có giá trị => trả tập tương ứng
	// Link từ API thường có hạn (expire), nên mỗi lần bấm Play phải gọi lại API
	streaming, err := s.provider.GetStreamingLinks(ctx, movie.ExternalID, episodeSlug)
	if err != nil {
		return nil, fmt.Errorf("lỗi lấy link từ provider: %w", err)
	}

	return streaming, nil
}

// ==========================================
// 🛠️ HELPER FUNCTIONS
// ==========================================

// convertDTOToModel: Chuyển Provider DTO -> GORM Model
// internal/service/movie_service.go

// convertDTOToModel: Chuyển Provider DTO → GORM Model
// ⚠️ QUAN TRỌNG: Chỉ lưu metadata, KHÔNG lưu episodes từ external API
func (s *MovieService) convertDTOToModel(dto *provider.MovieDTO) *models.Movie {
	movie := &models.Movie{
		// 🔑 Hybrid Source Tracking
		Source:     dto.Source,
		ExternalID: dto.ExternalID,

		// 📋 Metadata cơ bản
		Title:         dto.Title,
		OriginalTitle: dto.OriginalTitle,
		Slug:          dto.Slug,
		Description:   dto.Overview,
		ThumbURL:      dto.ThumbURL,
		PosterURL:     dto.PosterURL,
		BackdropURL:   dto.BackdropURL,

		// 📊 Metadata cho search/filter
		Type:        dto.Type,
		Status:      dto.Status,
		ReleaseYear: dto.ReleaseYear,
		Rating:      dto.Rating,
		Duration:    dto.Runtime,
		LastSyncedAt: time.Now(),

		// 🎬 Episodes:
		// - External API: KHÔNG lưu (fetch tươi khi play)
		// - Self-hosted: Sẽ được set riêng qua Admin Handler
		Episodes: nil, // ← Quan trọng: không lưu episodes từ Ophim1
	}

	// Map categories (genres/countries) → Lưu vào DB để search/filter nhanh
	for _, genre := range dto.Genres {
		movie.Categories = append(movie.Categories, models.Category{
			Name: genre,
			Slug: slugify(genre),
			Type: "genre",
		})
	}

	return movie
}

// slugify: Chuyển "Hành Động" -> "hanh-dong"
func slugify(name string) string {
	s := strings.ToLower(name)
	s = strings.ReplaceAll(s, " ", "-")
	s = strings.ReplaceAll(s, "đ", "d")
	return s
}

// isSeriesType: Kiểm tra xem type có phải series-like không
// Series-like types: "series", "hoathinh", "phimbo", "tvshows", etc.
// Phim lẻ: "movie", "phim-le", "single", etc.
func isSeriesType(t string) bool {
	t = strings.ToLower(t)
	seriesTypes := map[string]bool{
		"series":   true,
		"hoathinh": true,
		"phimbo":   true,
		"tvshows":  true,
		"tv":       true,
	}
	return seriesTypes[t]
}

// episodeSlugMatch: hỗ trợ match episode=1 với slug dạng tap-1/episode-1
func episodeSlugMatch(episodeSlug string, selected string) bool {
	episodeSlug = strings.ToLower(strings.TrimSpace(episodeSlug))
	selected = strings.ToLower(strings.TrimSpace(selected))

	if episodeSlug == selected {
		return true
	}

	normalize := func(v string) string {
		v = strings.TrimPrefix(v, "tap-")
		v = strings.TrimPrefix(v, "tập-")
		v = strings.TrimPrefix(v, "episode-")
		v = strings.TrimPrefix(v, "ep-")
		return strings.TrimSpace(v)
	}

	return normalize(episodeSlug) == normalize(selected)
}
