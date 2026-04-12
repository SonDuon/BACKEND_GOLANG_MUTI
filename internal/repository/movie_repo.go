package repository

import (
	"context"
	"fmt"

	"github.com/SonDuon/BACKEND_GOLANG_MUTI/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MovieRepository interface {
	Create(ctx context.Context, movie *models.Movie) error
	GetBySlug(ctx context.Context, slug string) (*models.Movie, error)
	GetByExternalID(ctx context.Context, source, externalID string) (*models.Movie, error)
}

type movieRepo struct {
	db *gorm.DB
}

func NewMovieRepository(db *gorm.DB) MovieRepository {
	return &movieRepo{db: db}
}

// Create: Import movie + xử lý duplicate categories bằng FindOrCreate
func (r *movieRepo) Create(ctx context.Context, movie *models.Movie) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {

		// 1️⃣ Kiểm tra phim đã tồn tại chưa (tránh import trùng)
		var existing models.Movie
		if err := tx.Where("source = ? AND external_id = ?", movie.Source, movie.ExternalID).
			First(&existing).Error; err == nil {
			// Đã tồn tại → Update thay vì Create
			movie.ID = existing.ID
			return tx.Model(&existing).Updates(movie).Error
		}

		// 2️⃣ FindOrCreate Categories (Genres/Countries)
		var categoryIDs []uuid.UUID
		for _, cat := range movie.Categories {
			if cat.Slug == "" || cat.Type == "" {
				continue
			}
			id, err := findOrCreateCategory(tx, cat.Name, cat.Slug, cat.Type)
			if err != nil {
				return err
			}
			categoryIDs = append(categoryIDs, id)
		}

		// 3️⃣ Tách relations ra khỏi struct movie để GORM chỉ insert bảng movies
		movie.Categories = nil
		if err := tx.Create(movie).Error; err != nil {
			return fmt.Errorf("create movie: %w", err)
		}

		// 4️⃣ Gán categories vào movie qua junction table (dùng Association Mode)
		if len(categoryIDs) > 0 {
			var categories []models.Category
			tx.Where("id IN ?", categoryIDs).Find(&categories)

			// Cast to []interface{} for Append
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

// GetBySlug & GetByExternalID: Query thông thường + Preload relations
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
