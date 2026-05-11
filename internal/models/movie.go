package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ─────────────────────────────────────
// Category Model (Unified: genre/country/year)
// ─────────────────────────────────────
type Category struct {
	ID        uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	Name      string    `gorm:"type:varchar(100);not null" json:"name"`
	Slug      string    `gorm:"type:varchar(100);uniqueIndex:idx_category_slug_type;not null" json:"slug"`
	Type      string    `gorm:"type:varchar(50);not null;index:idx_category_slug_type" json:"type"` // genre/country/year
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (Category) TableName() string { return "categories" }

func (c *Category) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

// ─────────────────────────────────────
// Movie Model
// ─────────────────────────────────────
type Movie struct {
	ID            uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	Title         string    `gorm:"type:varchar(255);not null;index" json:"title"`
	OriginalTitle string    `gorm:"type:varchar(255);index" json:"original_title,omitempty"`
	Slug          string    `gorm:"type:varchar(255);uniqueIndex;not null" json:"slug"`
	Description   string    `gorm:"type:text" json:"description"`
	ThumbURL      string    `gorm:"type:varchar(500)" json:"thumb_url,omitempty"`
	PosterURL     string    `gorm:"type:varchar(500)" json:"poster_url"`
	BackdropURL   string    `gorm:"type:varchar(500)" json:"backdrop_url,omitempty"`

	// 🔑 Hybrid Source Tracking (QUAN TRỌNG)
	Source     string `gorm:"type:varchar(50);not null;index:idx_source_external;default:'self'" json:"source"`
	ExternalID string `gorm:"type:varchar(100);not null;index:idx_source_external" json:"external_id"`

	// 📊 Metadata for Search/Filter
	Type        string  `gorm:"type:varchar(50);not null;index" json:"type"` // Lưu type nguyên bản: movie, series, hoathinh, phimbo, tvshows, etc.
	Status      string  `gorm:"type:varchar(50);index;check:status IN ('completed','ongoing')" json:"status"`
	ReleaseYear int     `gorm:"type:int;index" json:"release_year"`
	Rating      float32 `gorm:"type:decimal(3,2);default:0.00;index" json:"rating"`
	VoteCount   int     `gorm:"default:0" json:"vote_count"`
	Duration    int     `gorm:"type:int" json:"duration"` // seconds

	// 📺 Series Info
	TotalEpisodes *int `gorm:"type:int" json:"total_episodes,omitempty"`

	// 🔗 Many-to-Many Relations (FindOrCreate-friendly)
	Categories []Category `gorm:"many2many:movie_categories;constraint:OnDelete:CASCADE" json:"categories,omitempty"`

	// 🔗 One-to-Many Relations
	Episodes []Episode `gorm:"foreignKey:MovieID;constraint:OnDelete:CASCADE" json:"episodes,omitempty"`

	// ⏰ Timestamps
	LastSyncedAt time.Time      `gorm:"index" json:"last_synced_at"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Movie) TableName() string { return "movies" }

func (m *Movie) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}

// ─────────────────────────────────────
// MovieCategory Junction Table (Explicit for FindOrCreate)
// ─────────────────────────────────────
type MovieCategory struct {
	MovieID    uuid.UUID `gorm:"type:uuid;not null;primaryKey" json:"movie_id"`
	CategoryID uuid.UUID `gorm:"type:uuid;not null;primaryKey" json:"category_id"`
}

func (MovieCategory) TableName() string { return "movie_categories" }
