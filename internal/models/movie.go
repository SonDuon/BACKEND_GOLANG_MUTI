package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Category đại diện cho bảng "categories" (Thể loại, Quốc gia, Năm)
type Category struct {
	ID     uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	Name   string    `gorm:"type:varchar(100);not null" json:"name"`
	Slug   string    `gorm:"type:varchar(100);uniqueIndex;not null" json:"slug"`
	Type   string    `gorm:"type:varchar(50)" json:"type"` // Ví dụ: "genre", "country", "year"
	Movies []Movie   `gorm:"many2many:movie_categories;" json:"-"`
}

// Movie đại diện cho bảng "movies"
type Movie struct {
	ID            uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	Title         string    `gorm:"type:varchar(255);not null;index" json:"title"`
	OriginalTitle string    `gorm:"type:varchar(255);index" json:"original_title,omitempty"` // Thêm để search tốt hơn
	Slug          string    `gorm:"type:varchar(255);uniqueIndex;not null" json:"slug"`
	Description   string    `gorm:"type:text" json:"description"`
	PosterURL     string    `gorm:"type:varchar(500)" json:"poster_url"`
	BackdropURL   string    `gorm:"type:varchar(500)" json:"backdrop_url,omitempty"`

	// 🔑 QUAN TRỌNG: Hybrid Source Tracking
	Source     string `gorm:"type:varchar(50);not null;index;default:'self'" json:"source"` // "ophim1", "self", "nguonc"
	ExternalID string `gorm:"type:varchar(100);index" json:"external_id"`                   // ID gốc từ API (slug, tmdb_id...)

	// 📊 Metadata bổ sung cho filter/search
	Type        string  `gorm:"type:varchar(50);index" json:"type"`                 // "series" hoặc "single"
	Status      string  `gorm:"type:varchar(50);index" json:"status"`               // "completed", "ongoing"
	ReleaseYear int     `gorm:"type:int;index" json:"release_year"`                 // Thêm để filter theo năm
	Rating      float32 `gorm:"type:decimal(3,2);default:0.00;index" json:"rating"` // Thêm để sort popular

	// 🔗 Relations
	Categories []Category `gorm:"many2many:movie_categories;" json:"categories"`
	Episodes   []Episode  `gorm:"foreignKey:MovieID;constraint:OnDelete:CASCADE" json:"episodes,omitempty"`

	// ⏰ Timestamps
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
