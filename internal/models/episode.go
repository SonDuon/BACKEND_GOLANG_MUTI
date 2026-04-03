package models

import (
	"time"
	"github.com/google/uuid"
)

// Episode đại diện cho bảng "episodes" (Tập phim)
type Episode struct {
	ID           uuid.UUID     `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	MovieID      uuid.UUID     `gorm:"type:uuid;not null;index" json:"movie_id"`
	Name         string        `gorm:"type:varchar(100);not null" json:"name"` // VD: "Tập 1", "Tập 2"
	Slug         string        `gorm:"type:varchar(100);not null;index" json:"slug"`
	MediaSources []MediaSource `gorm:"foreignKey:EpisodeID" json:"media_sources"`
	CreatedAt    time.Time     `json:"created_at"`
}

// MediaSource đại diện cho bảng "media_sources" (Bản đồ chỉ đường API)
type MediaSource struct {
	ID         uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	EpisodeID  uuid.UUID `gorm:"type:uuid;not null;index" json:"episode_id"`
	ServerName string    `gorm:"type:varchar(100);not null" json:"server_name"` // VD: "Ophim #1", "Nguonc #1", "Server VIP"
	SourceType string    `gorm:"type:varchar(50);not null" json:"source_type"`   // VD: "external_api", "self_hosted"
	SourceKey  string    `gorm:"type:varchar(500);not null" json:"source_key"`  // Slug để gọi API (truc-ngoc) hoặc Link file MP4 tự host
	CreatedAt  time.Time `json:"created_at"`
}