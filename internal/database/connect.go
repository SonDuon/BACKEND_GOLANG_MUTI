package database

import (
	"fmt"
	"log"
	"time"

	"github.com/SonDuon/BACKEND_GOLANG_MUTI/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// Config chứa thông tin kết nối
type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string // disable, require, verify-full
}

// ConnectDB khởi tạo kết nối và chạy AutoMigrate
func ConnectDB(cfg Config) (*gorm.DB, error) {
	// Build DSN (Data Source Name)
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=Asia/Ho_Chi_Minh",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
	)

	// Cấu hình logger cho GORM (hiển thị query khi debug)
	newLogger := logger.New(
		log.New(log.Writer(), "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Warn,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		},
	)

	// Mở kết nối
	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger:                 newLogger,
		SkipDefaultTransaction: true, // Optional: tăng performance
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect database: %w", err)
	}
	DB.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp";`)
	//  AUTO MIGRATE - Tạo/cập nhật table tự động
	err = DB.AutoMigrate(
		&models.User{},
		&models.Movie{},
		&models.Category{},
		&models.Episode{},
		&models.MediaSource{},
		&models.WatchHistory{},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to auto migrate: %w", err)
	}

	log.Println("✅ Database connected and migrated successfully!")
	return DB, nil
}

// CloseDB đóng kết nối khi ứng dụng dừng
func CloseDB() error {
	if DB != nil {
		sqlDB, err := DB.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}
