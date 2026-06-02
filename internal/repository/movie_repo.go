package repository

import (
	"context"
	"fmt"
	"sync"

	"github.com/SonDuon/BACKEND_GOLANG_MUTI/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// MovieFilter cấu hình truy vấn phân trang & lọc
type MovieFilter struct {
	Page      int
	Limit     int
	SortBy    string // format: "field:asc" hoặc "field:desc"
	Year      int
	Status    string
	GenreSlug string // Dùng cho ListByYear kết hợp lọc thể loại
}

func (f *MovieFilter) GetOffset() int {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.Limit < 1 {
		f.Limit = 20
	}
	return (f.Page - 1) * f.Limit
}

type MovieRepository interface {
	Create(ctx context.Context, movie *models.Movie) error
	GetBySlug(ctx context.Context, slug string) (*models.Movie, error)
	GetByExternalID(ctx context.Context, source, externalID string) (*models.Movie, error)
	List(ctx context.Context, page, limit int, typeFilter, statusFilter string) ([]models.Movie, int64, error)
	Search(ctx context.Context, query string, page, limit int) ([]models.Movie, int64, error)

	// 🆕 Category-Based Filtering
	ListByCategory(ctx context.Context, categoryType, slug string, filters *MovieFilter) ([]models.Movie, int64, error)
	ListByYear(ctx context.Context, year int, filters *MovieFilter) ([]models.Movie, int64, error)
	GetOrCreateCategory(ctx context.Context, name, slug, categoryType string) (*models.Category, error)
}

type movieRepo struct {
	db *gorm.DB
	mu sync.Mutex
}

func NewMovieRepository(db *gorm.DB) MovieRepository {
	return &movieRepo{db: db}
}

// ------------------------------------------------------------------
// CORE CRUD & SYNC METHODS
// ------------------------------------------------------------------

func (r *movieRepo) Create(ctx context.Context, movie *models.Movie) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1️⃣ Kiểm tra trùng (source + external_id) hoặc slug
		var existing models.Movie
		if err := tx.Where("(source = ? AND external_id = ?) OR slug = ?", movie.Source, movie.ExternalID, movie.Slug).First(&existing).Error; err == nil {
			movie.ID = existing.ID
			return tx.Model(&existing).Updates(movie).Error
		}

		// 2️⃣ FindOrCreate Categories
		var categoryIDs []uuid.UUID
		for _, cat := range movie.Categories {
			if cat.Slug == "" || cat.Type == "" {
				continue
			}
			createdCat, err := r.getOrCreateCategoryTx(tx, cat.Name, cat.Slug, cat.Type)
			if err != nil {
				return err
			}
			categoryIDs = append(categoryIDs, createdCat.ID)
		}

		// 3️⃣ Tách relations để GORM chỉ insert bảng movies
		movie.Categories = nil
		if err := tx.Create(movie).Error; err != nil {
			return fmt.Errorf("create movie: %w", err)
		}

		// 4️⃣ Gán categories qua junction table
		if len(categoryIDs) > 0 {
			var categories []models.Category
			tx.Where("id IN ?", categoryIDs).Find(&categories)

			cats := make([]interface{}, len(categories))
			for i, cat := range categories {
				cats[i] = cat
			}

			if err := tx.Model(movie).Association("Categories").Append(cats...); err != nil {
				return fmt.Errorf("associate categories: %w", err)
			}
		}
		return nil
	})
}

func (r *movieRepo) GetBySlug(ctx context.Context, slug string) (*models.Movie, error) {
	var movie models.Movie
	err := r.db.WithContext(ctx).
		Preload("Categories").
		Preload("Episodes.MediaSources").
		Where("slug = ?", slug).
		First(&movie).Error
	if err != nil {
		return nil, err
	}
	return &movie, nil
}

func (r *movieRepo) GetByExternalID(ctx context.Context, source, externalID string) (*models.Movie, error) {
	var movie models.Movie
	err := r.db.WithContext(ctx).
		Preload("Categories").
		Preload("Episodes.MediaSources").
		Where("source = ? AND external_id = ?", source, externalID).
		First(&movie).Error
	if err != nil {
		return nil, err
	}
	return &movie, nil
}

func (r *movieRepo) List(ctx context.Context, page, limit int, typeFilter, statusFilter string) ([]models.Movie, int64, error) {
	var movies []models.Movie
	var total int64
	query := r.db.WithContext(ctx).Model(&models.Movie{})

	if typeFilter != "" {
		query = query.Where("type = ?", typeFilter)
	}
	if statusFilter != "" {
		query = query.Where("status = ?", statusFilter)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := query.
		Preload("Categories").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&movies).Error

	return movies, total, err
}

func (r *movieRepo) Search(ctx context.Context, query string, page, limit int) ([]models.Movie, int64, error) {
	var movies []models.Movie
	var total int64
	queryStr := "%" + query + "%"
	db := r.db.WithContext(ctx).Model(&models.Movie{}).
		Where("title ILIKE ? OR original_title ILIKE ? OR description ILIKE ?", queryStr, queryStr, queryStr).
		Preload("Categories")

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := db.Limit(limit).Offset(offset).Find(&movies).Error
	return movies, total, err
}

// ------------------------------------------------------------------
// 🆕 CATEGORY-BASED FILTERING IMPLEMENTATIONS
// ------------------------------------------------------------------

func (r *movieRepo) ListByCategory(ctx context.Context, categoryType, slug string, filters *MovieFilter) ([]models.Movie, int64, error) {
	var movies []models.Movie
	var total int64

	query := r.db.WithContext(ctx).Model(&models.Movie{}).
		Joins("JOIN movie_categories mc ON mc.movie_id = movies.id").
		Joins("JOIN categories c ON c.id = mc.category_id AND c.type = ? AND c.slug = ?", categoryType, slug)

	// Áp dụng bộ lọc phụ
	if filters.Status != "" {
		query = query.Where("movies.status = ?", filters.Status)
	}
	if filters.Year != 0 {
		query = query.Where("movies.release_year = ?", filters.Year)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count movies by category: %w", err)
	}

	// Default sort: release_year DESC
	sortClause := "movies.release_year DESC"
	if filters.SortBy != "" {
		sortClause = "movies." + filters.SortBy
	}

	offset := filters.GetOffset()
	err := query.
		Preload("Categories").
		Order(sortClause).
		Limit(filters.Limit).
		Offset(offset).
		Find(&movies).Error

	return movies, total, err
}

func (r *movieRepo) ListByYear(ctx context.Context, year int, filters *MovieFilter) ([]models.Movie, int64, error) {
	var movies []models.Movie
	var total int64

	query := r.db.WithContext(ctx).Model(&models.Movie{}).
		Where("movies.release_year = ?", year)

	if filters.GenreSlug != "" {
		query = query.
			Joins("JOIN movie_categories mc ON mc.movie_id = movies.id").
			Joins("JOIN categories c ON c.id = mc.category_id AND c.type = 'genre' AND c.slug = ?", filters.GenreSlug)
	}
	if filters.Status != "" {
		query = query.Where("movies.status = ?", filters.Status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count movies by year: %w", err)
	}

	offset := filters.GetOffset()
	err := query.
		Preload("Categories").
		Order("movies.release_year DESC").
		Limit(filters.Limit).
		Offset(offset).
		Find(&movies).Error

	return movies, total, err
}

// GetOrCreateCategory: Public method for auto-sync
func (r *movieRepo) GetOrCreateCategory(ctx context.Context, name, slug, categoryType string) (*models.Category, error) {
	return r.getOrCreateCategoryTx(r.db.WithContext(ctx), name, slug, categoryType)
}

// getOrCreateCategoryTx: Transaction-safe internal helper
func (r *movieRepo) getOrCreateCategoryTx(tx *gorm.DB, name, slug, categoryType string) (*models.Category, error) {
	id, err := findOrCreateCategory(tx, name, slug, categoryType)
	if err != nil {
		return nil, err
	}
	return &models.Category{
		ID:   id,
		Name: name,
		Slug: slug,
		Type: categoryType,
	}, nil
}
