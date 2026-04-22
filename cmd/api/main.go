package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/SonDuon/BACKEND_GOLANG_MUTI/internal/config"
	"github.com/SonDuon/BACKEND_GOLANG_MUTI/internal/database"
	"github.com/SonDuon/BACKEND_GOLANG_MUTI/internal/handler"
	"github.com/SonDuon/BACKEND_GOLANG_MUTI/internal/provider/ophim1"
	"github.com/SonDuon/BACKEND_GOLANG_MUTI/internal/repository"
	"github.com/SonDuon/BACKEND_GOLANG_MUTI/internal/service"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// --- 1. CONFIG & ENVIRONMENT ---
	config.LoadEnv()

	// --- 2. DATABASE CONNECTION ---
	dbCfg := database.Config{
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		DBName:   os.Getenv("DB_NAME"),
		SSLMode:  "disable",
	}

	db, err := database.ConnectDB(dbCfg)
	if err != nil {
		log.Fatal("❌ Hệ thống sập do không thể kết nối Database: ", err)
	}
	defer database.CloseDB()

	// --- 3. PROVIDERS & REPOSITORIES ---
	ophimAdapter := ophim1.New(ophim1.DefaultConfig(), nil)
	movieRepo := repository.NewMovieRepository(db)

	// --- 4. SERVICES ---
	movieSvc := service.NewMovieService(movieRepo, ophimAdapter)

	// --- 5. HANDLERS ---
	movieHandler := handler.NewMovieHandler(movieSvc)
	adminHandler := handler.NewAdminHandler(movieRepo, ophimAdapter)

	// --- 6. ROUTER SETUP ---
	gin.SetMode(gin.ReleaseMode) // Default to release, can override via ENV if needed
	if os.Getenv("GIN_MODE") == "debug" {
		gin.SetMode(gin.DebugMode)
	}

	r := gin.Default()

	// Middleware: CORS
	r.Use(cors.New(cors.Config{
		AllowOriginFunc:  func(origin string) bool { return true },
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length", "Authorization"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// --- 7. ROUTES REGISTRATION ---
	// Health check
	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "time": time.Now().Format(time.RFC3339)})
	})

	// API V1 Group
	v1 := r.Group("/api/v1")
	{
		adminHandler.RegisterRoutes(v1)
		movieHandler.RegisterRoutes(v1)
	}

	// --- 8. START SERVER ---
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("🚀 Server đang chạy tại: http://localhost:%s\n", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal("❌ Server lỗi khi khởi động: ", err)
	}
}
