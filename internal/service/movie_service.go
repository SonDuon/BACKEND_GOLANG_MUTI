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

// NewMovieService khởi tạo service
func NewMovieService(repo repository.MovieRepository, provider provider.MovieProvider) *MovieService {
	return &MovieService{
		repo:     repo,
		provider: provider,
	}
}

// ==========================================
// 🔍 SEARCH
// ==========================================
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
		if err := s.repo.Create(ctx, movieModel); err != nil {
			log.Printf("⚠️ Auto-import failed for '%s': %v", dto.Slug, err)
		}
		movies = append(movies, *movieModel)
	}

	return movies, searchResult.Total, nil
}

// ==========================================
// 📋 LIST (Phân trang cơ bản)
// ==========================================
func (s *MovieService) ListMovies(ctx context.Context, page, limit int, typeFilter, statusFilter string) ([]models.Movie, int64, error) {
	return s.repo.List(ctx, page, limit, typeFilter, statusFilter)
}

// GetMovieList: Giữ lại tương thích với code cũ (dùng trực tiếp DB)
func (s *MovieService) GetMovieList(ctx context.Context, page, limit int) ([]*models.Movie, int64, error) {
	var movies []*models.Movie
	var total int64

	query := database.DB.Model(&models.Movie{})
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

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
// 🎬 DETAIL (Lazy Import Pattern)
// ==========================================
func (s *MovieService) GetMovieDetail(ctx context.Context, slug string) (*models.Movie, error) {
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

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("database error: %w", err)
	}

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
// ▶️ WATCH (Lấy link xem)
// ==========================================
func (s *MovieService) GetWatchLinks(ctx context.Context, slug string, episodeSlug string) (*provider.StreamingDTO, error) {
	movie, err := s.GetMovieDetail(ctx, slug)
	if err != nil {
		return nil, err
	}

	episodeSlug = strings.TrimSpace(episodeSlug)
	if strings.EqualFold(episodeSlug, "full") {
		episodeSlug = ""
	}

	if movie.Source == "self" {
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
		}
		return nil, fmt.Errorf("không tìm thấy tập '%s' hoặc nguồn phát", episodeSlug)
	}

	streaming, err := s.provider.GetStreamingLinks(ctx, movie.ExternalID, episodeSlug)
	if err != nil {
		return nil, fmt.Errorf("lỗi lấy link từ provider: %w", err)
	}

	return streaming, nil
}

// ==========================================
// 📂 CATEGORY-BASED FILTERING (MỚI)
// ==========================================

// GetMoviesByCategory: Lọc phim theo Genre hoặc Country
func (s *MovieService) GetMoviesByCategory(ctx context.Context, categoryType, slug string, page, limit int, year int, status string) ([]models.Movie, int64, error) {
	// 1. Validate category type
	if categoryType != "genre" && categoryType != "country" {
		return nil, 0, errors.New("invalid category type: must be 'genre' or 'country'")
	}

	// 2. Build filter struct
	filters := &repository.MovieFilter{
		Page:   page,
		Limit:  limit,
		Year:   year,
		Status: status,
	}

	// 3. Call repository
	movies, total, err := s.repo.ListByCategory(ctx, categoryType, strings.TrimSpace(slug), filters)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list movies by category: %w", err)
	}

	return movies, total, nil
}

// GetMoviesByYear: Lọc phim theo năm, hỗ trợ kết hợp genre
func (s *MovieService) GetMoviesByYear(ctx context.Context, year int, page, limit int, status, genreSlug string) ([]models.Movie, int64, error) {
	// 1. Validate year cơ bản
	if year < 1900 || year > time.Now().Year()+2 {
		return nil, 0, errors.New("invalid year parameter")
	}

	// 2. Build filter struct
	filters := &repository.MovieFilter{
		Page:      page,
		Limit:     limit,
		Status:    status,
		GenreSlug: strings.TrimSpace(genreSlug),
	}

	// 3. Call repository
	movies, total, err := s.repo.ListByYear(ctx, year, filters)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list movies by year: %w", err)
	}

	return movies, total, nil
}

// ==========================================
// 🛠️ HELPER FUNCTIONS
// ==========================================

func (s *MovieService) convertDTOToModel(dto *provider.MovieDTO) *models.Movie {
	movie := &models.Movie{
		Source:        dto.Source,
		ExternalID:    dto.ExternalID,
		Title:         dto.Title,
		OriginalTitle: dto.OriginalTitle,
		Slug:          dto.Slug,
		Description:   dto.Overview,
		ThumbURL:      dto.ThumbURL,
		PosterURL:     dto.PosterURL,
		BackdropURL:   dto.BackdropURL,
		Type:          dto.Type,
		Status:        dto.Status,
		ReleaseYear:   dto.ReleaseYear,
		Rating:        dto.Rating,
		Duration:      dto.Runtime,
		LastSyncedAt:  time.Now(),
		Episodes:      nil, // External API episodes không lưu DB, fetch tươi khi play
	}

	// Map categories (genres/countries)
	for _, genre := range dto.Genres {
		movie.Categories = append(movie.Categories, models.Category{
			Name: genre,
			Slug: slugify(genre),
			Type: "genre",
		})
	}

	return movie
}

func slugify(name string) string {
	s := strings.ToLower(name)
	s = strings.ReplaceAll(s, " ", "-")
	s = strings.ReplaceAll(s, "đ", "d")
	return s
}

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
