package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

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
		return movie, nil 
	}

	// Kiểm tra lỗi "Không tìm thấy bản ghi"
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("database error: %w", err)
	}

	// 2️⃣ Không có trong DB -> Gọi Provider (Ophim1)
	log.Printf("🔍 Movie '%s' not found in DB. Fetching from provider...", slug)
	dto, err := s.provider.GetByExternalID(ctx, slug)
	if err != nil {
		return nil, fmt.Errorf("provider error: %w", err)
	}

	// 3️⃣ Chuyển đổi DTO -> Model để lưu
	movieModel := s.convertDTOToModel(dto)

	// 4️⃣ Tự động Import vào DB (Dùng repo.Create đã có FindOrCreate)
	if err := s.repo.Create(ctx, movieModel); err != nil {
		// ⚠️ Log lỗi nhưng KHÔNG fail request. User vẫn nhận được data từ API.
		log.Printf("⚠️ Auto-import failed for '%s': %v", slug, err)
	}

	// 5️⃣ Trả về phim vừa fetch/import
	return movieModel, nil
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

	// 2️⃣ Xử lý theo Source
	if movie.Source == "self" {
		// 🏠 TỰ HOST: Lấy link từ Database (MediaSources)
		for _, ep := range movie.Episodes {
			if ep.Slug == episodeSlug {
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
		PosterURL:     dto.PosterURL,
		BackdropURL:   dto.BackdropURL,
		
		// 📊 Metadata cho search/filter
		Type:          dto.Type,
		Status:        dto.Status,
		ReleaseYear:   dto.ReleaseYear,
		Rating:        dto.Rating,
		Duration:      dto.Runtime,
		
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
