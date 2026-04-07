// File: internal/config/env.go
package config

import (
	"log"

	"github.com/joho/godotenv"
)

// LoadEnv nạp cấu hình từ file .env
func LoadEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Println("⚠️ Không tìm thấy file .env, dùng biến môi trường mặc định của OS")
	} else {
		log.Println("✅ Đã nạp thành công file cấu hình .env")
	}
}

