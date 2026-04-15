package main

import (
	"fmt"
	"log"
	"os"

	"github.com/SonDuon/BACKEND_GOLANG_MUTI/internal/config"
	"github.com/SonDuon/BACKEND_GOLANG_MUTI/internal/database"
	"github.com/SonDuon/BACKEND_GOLANG_MUTI/internal/handler"
	"github.com/SonDuon/BACKEND_GOLANG_MUTI/internal/provider/ophim1"
	"github.com/SonDuon/BACKEND_GOLANG_MUTI/internal/repository"
	"github.com/SonDuon/BACKEND_GOLANG_MUTI/internal/service"
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

	// 4. Khởi tạo Repository & Provider
	movieRepo := repository.NewMovieRepository(database.DB)
	ophimAdapter := ophim1.New(ophim1.DefaultConfig(), nil)

	// 5. Khởi tạo Handler
	movieSvc := service.NewMovieService(movieRepo, ophimAdapter)
	movieHandler := handler.NewMovieHandler(movieSvc)

	// 6. Setup Gin Router
	r := gin.Default()
	api := r.Group("/api/v1")

	// Register admin routes
	adminHandler := handler.NewAdminHandler(movieRepo, ophimAdapter)
	adminHandler.RegisterRoutes(api)
	// Register movie routes
	movieHandler.RegisterRoutes(api)

	// Health check route
	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok" })
	})

	// 7. Run server
	port := ":8080"
	fmt.Printf("🚀 Server Gin đang chạy: http://localhost%s\n", port)
	if err := r.Run(port); err != nil {
		log.Fatal("❌ Server Lỗi: ", err)
	}
}
