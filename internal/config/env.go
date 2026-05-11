// File: internal/config/env.go
package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// LoadEnv nạp cấu hình từ file .env
func LoadEnv() {
	if os.Getenv("GIN_MODE") != "release" {
		err := godotenv.Load()
		if err != nil {
			log.Println("⚠️ Không tìm thấy file .env")
		} else {
			log.Println("✅ Đã nạp thành công file .env")
		}
	}
}
