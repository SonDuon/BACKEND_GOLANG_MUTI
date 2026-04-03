package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	Email        string    `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	PasswordHash string    `gorm:"type:varchar(255);not null" json:"-"` // Dấu "-" để không bao giờ trả password ra JSON
	FullName     string    `gorm:"type:varchar(255)" json:"full_name"`
	CreatedAt    time.Time `json:"created_at"`
}

type WatchHistory struct {
	ID          uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	UserID      uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	MovieID     uuid.UUID `gorm:"type:uuid;not null;index" json:"movie_id"`
	EpisodeID   uuid.UUID `gorm:"type:uuid;not null" json:"episode_id"`
	CurrentTime int       `gorm:"type:int;default:0" json:"current_time"` // Lưu số giây đang xem dở
	UpdatedAt   time.Time `json:"updated_at"`
}
