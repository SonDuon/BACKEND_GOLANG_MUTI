package config

import (
	"database/sql"
	"log"
	"time"
	"github.com/SonDuon/BACKEND_GOLANG_MUTI/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB là biến global để sử dụng ở các package khác
var DB *gorm.DB
var SQLDB *sql.DB

func ConnectDB() {
	dsn := ""

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatal("❌ Lỗi kết nối DB:", err)
	}

	// Lấy *sql.DB để cấu hình connection pool
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal("❌ Không thể lấy sql.DB:", err)
	}

	// Cấu hình connection pool
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(25)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	// Kiểm tra kết nối
	if err := sqlDB.Ping(); err != nil {
		log.Fatal("❌ Không thể ping DB:", err)
	}

	SQLDB = sqlDB
	DB = db

	// Thực hiện AutoMigrate
	err = db.AutoMigrate(
		&models.Category{},
		&models.Movie{},
		&models.Episode{},
		&models.MediaSource{},
		&models.User{},
		&models.WatchHistory{},
	)
	if err != nil {
		log.Fatal("❌ Lỗi Migration:", err)
	}

	log.Println("✅ Database đã được đồng bộ hóa thành công!")
}
