// File: cmd/api/main.go
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/SonDuon/BACKEND_GOLANG_MUTI/internal/config"
	"github.com/SonDuon/BACKEND_GOLANG_MUTI/internal/database"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
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
	// 3. Khởi động server API
	app := fiber.New()
	app.Use(logger.New())

	app.Get("/api/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "success",
			"message": "Trái tim Backend đang đập rất khỏe!",
		})
	})
	port := "3000"
	fmt.Printf("🚀 Server đang lắng nghe tại: http://localhost:%s\n", port)
	err = app.Listen(":" + port)
	if err != nil {
		log.Fatal("❌ Server sập: ", err)
	}
}
