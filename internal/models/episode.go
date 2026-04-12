package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ─────────────────────────────────────
// Episode Model
// ─────────────────────────────────────
type Episode struct {
	ID          uuid.UUID  `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	MovieID     uuid.UUID  `gorm:"type:uuid;not null;index" json:"movie_id"`
	Title       string     `gorm:"type:varchar(255);not null" json:"title"`
	Slug        string     `gorm:"type:varchar(255);not null;uniqueIndex:idx_movie_slug" json:"slug"`
	EpisodeNum  int        `gorm:"column:episode_number;type:int;not null" json:"episode_number"`
	Season      int        `gorm:"column:season_number;type:int;default:1" json:"season_number"`
	Description string     `gorm:"type:text" json:"description,omitempty"`
	Thumbnail   string     `gorm:"type:varchar(500)" json:"thumbnail,omitempty"`
	Duration    int        `gorm:"type:int" json:"duration"` // seconds
	AirDate     *time.Time `gorm:"type:date" json:"air_date,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`

	// 🔗 One-to-Many: Media Sources
	MediaSources []MediaSource `gorm:"foreignKey:EpisodeID;constraint:OnDelete:CASCADE" json:"media_sources,omitempty"`
}

func (Episode) TableName() string { return "episodes" }

func (e *Episode) BeforeCreate(tx *gorm.DB) error {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	return nil
}

// ─────────────────────────────────────
// MediaSource Model (Multi-source video links)
// ─────────────────────────────────────
type MediaSource struct {
	ID         uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	EpisodeID  uuid.UUID `gorm:"type:uuid;not null;index" json:"episode_id"`
	ServerName string    `gorm:"type:varchar(100);not null" json:"server_name"` // "Vietsub #1", "Server VIP"
	SourceType string    `gorm:"type:varchar(50);not null;check:source_type IN ('external_api','self_hosted')" json:"source_type"`
	SourceKey  string    `gorm:"type:varchar(1000);not null" json:"source_key"` // slug for API call OR signed URL for self-hosted
	Quality    string    `gorm:"type:varchar(20)" json:"quality"`               // "1080p", "720p", "auto"
	IsDefault  bool      `gorm:"default:false" json:"is_default"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func (MediaSource) TableName() string { return "media_sources" }

func (m *MediaSource) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}
