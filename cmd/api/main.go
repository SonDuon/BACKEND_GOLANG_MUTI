package main

import (
	"fmt"
	"log"
	"os"

	"github.com/SonDuon/BACKEND_GOLANG_MUTI/internal/config"
	"github.com/SonDuon/BACKEND_GOLANG_MUTI/internal/database"
	"github.com/gin-gonic/gin"
)

func main() {
	// 1. Nạp các biến từ file .env vào hệ thống
	config.LoadEnv()

	// 2. Kết nối đến Database
	cfg := database.Config{
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		DBName:   os.Getenv("DB_NAME"),
		SSLMode:  "disable",
	}

	_, err := database.ConnectDB(cfg)
	if err != nil {
		log.Fatal("❌ Hệ thống sập do không thể kết nối Database: ", err)
	}
	defer database.CloseDB()

	// 3. Khởi động server API với Gin
	// gin.DebugMode hiện query SQL ra console (tiện khi dev)
	gin.SetMode(gin.DebugMode)

	// gin.Default() đã bao gồm middleware Logger & Recovery (bắt panic)
	r := gin.Default()

	// Route Health Check
	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "success",
			"message": "Trái tim Backend đang đập rất khỏe!",
		})
	})

	// 🌐 Cổng mặc định cho Go API (tránh conflict với Next.js port 3000)
	port := ":8080"
	fmt.Printf("🚀 Server Gin đang chạy: http://localhost%s\n", port)

	if err := r.Run(port); err != nil {
		log.Fatal("❌ Server Lỗi: ", err)
	}
}
