package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ─────────────────────────────────────
// User Model
// ─────────────────────────────────────
type User struct {
	ID           uuid.UUID      `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	Email        string         `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	PasswordHash string         `gorm:"type:varchar(255);not null" json:"-"` // Never expose
	FullName     string         `gorm:"type:varchar(255)" json:"full_name,omitempty"`
	AvatarURL    string         `gorm:"type:varchar(500)" json:"avatar_url,omitempty"`
	Role         string         `gorm:"type:varchar(20);default:'user';check:role IN ('user','admin')" json:"role"`
	IsVerified   bool           `gorm:"default:false" json:"is_verified"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

func (User) TableName() string { return "users" }

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

// ─────────────────────────────────────
// WatchHistory Model
// ─────────────────────────────────────
type WatchHistory struct {
	ID            uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	UserID        uuid.UUID `gorm:"type:uuid;not null;index:idx_user_updated" json:"user_id"`
	ContentID     uuid.UUID `gorm:"type:uuid;not null;index:idx_user_updated" json:"content_id"` // movie_id or episode_id
	ContentType   string    `gorm:"type:varchar(20);not null;check:content_type IN ('movie','episode')" json:"content_type"`
	LastPosition  int       `gorm:"type:int;default:0" json:"last_position"` // seconds
	TotalDuration *int      `gorm:"type:int" json:"total_duration,omitempty"`
	IsCompleted   bool      `gorm:"default:false" json:"is_completed"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func (WatchHistory) TableName() string { return "watch_history" }

// Composite primary key + index for fast user history lookup
func (WatchHistory) TableNameWithIndex() string {
	return "watch_history"
}

func (w *WatchHistory) BeforeCreate(tx *gorm.DB) error {
	if w.ID == uuid.Nil {
		w.ID = uuid.New()
	}
	return nil
}

// ─────────────────────────────────────
// Favorite Model (Watchlist)
// ─────────────────────────────────────
type Favorite struct {
	ID        uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_user_movie" json:"user_id"`
	MovieID   uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_user_movie" json:"movie_id"`
	CreatedAt time.Time `json:"created_at"`
}

func (Favorite) TableName() string { return "favorites" }

func (f *Favorite) BeforeCreate(tx *gorm.DB) error {
	if f.ID == uuid.Nil {
		f.ID = uuid.New()
	}
	return nil
}
