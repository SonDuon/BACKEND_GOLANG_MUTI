package models

import (
	"time"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Category đại diện cho bảng "categories" (Thể loại, Quốc gia, Năm)
type Category struct {
	ID        uuid.UUID      `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	Name      string         `gorm:"type:varchar(100);not null" json:"name"`
	Slug      string         `gorm:"type:varchar(100);uniqueIndex;not null" json:"slug"`
	Type      string         `gorm:"type:varchar(50)" json:"type"` // Ví dụ: "genre", "country", "year"
	Movies    []Movie        `gorm:"many2many:movie_categories;" json:"-"`
}

// Movie đại diện cho bảng "movies"
type Movie struct {
	ID          uuid.UUID      `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	Title       string         `gorm:"type:varchar(255);not null;index" json:"title"`
	Slug        string         `gorm:"type:varchar(255);uniqueIndex;not null" json:"slug"`
	Description string         `gorm:"type:text" json:"description"`
	PosterURL   string         `gorm:"type:varchar(500)" json:"poster_url"`
	BackdropURL string         `gorm:"type:varchar(500)" json:"backdrop_url"`
	Type        string         `gorm:"type:varchar(50);index" json:"type"` // "series" hoặc "single"
	Status      string         `gorm:"type:varchar(50)" json:"status"`     // "completed", "ongoing"
	Categories  []Category     `gorm:"many2many:movie_categories;" json:"categories"`
	Episodes    []Episode      `gorm:"foreignKey:MovieID" json:"episodes"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}