package repository

import (
	"fmt"

	"github.com/SonDuon/BACKEND_GOLANG_MUTI/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// findOrCreateCategory: Tìm category theo slug + type, nếu chưa có thì tạo mới
// Trả về ID của category (dùng để gán vào junction table movie_categories)
func findOrCreateCategory(db *gorm.DB, name, slug, categoryType string) (uuid.UUID, error) {
	var cat models.Category

	// 🔑 GORM FirstOrCreate: Tìm trước, nếu không thấy thì Insert với struct thứ 2
	err := db.Where("slug = ? AND type = ?", slug, categoryType).
		FirstOrCreate(&cat, models.Category{
			Name: name,
			Slug: slug,
			Type: categoryType,
		}).Error

	if err != nil {
		return uuid.Nil, fmt.Errorf("findOrCreateCategory failed: %w", err)
	}

	return cat.ID, nil
}
