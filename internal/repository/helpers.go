package repository

import (
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// findOrCreateCategory: Tìm category theo slug + type, nếu chưa có thì tạo mới (concurrent-safe)
func findOrCreateCategory(db *gorm.DB, name, slug, categoryType string) (uuid.UUID, error) {
	var idStr string
	newUUID := uuid.New()

	// Sử dụng Upsert nguyên tử trên Postgres
	err := db.Raw(`
		INSERT INTO categories (id, name, slug, type, created_at, updated_at)
		VALUES (?, ?, ?, ?, NOW(), NOW())
		ON CONFLICT (slug, type)
		DO UPDATE SET name = EXCLUDED.name
		RETURNING id
	`, newUUID, name, slug, categoryType).Scan(&idStr).Error

	if err != nil {
		return uuid.Nil, fmt.Errorf("findOrCreateCategory failed to execute upsert: %w", err)
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		return uuid.Nil, fmt.Errorf("findOrCreateCategory failed to parse returned uuid: %w", err)
	}

	return id, nil
}
