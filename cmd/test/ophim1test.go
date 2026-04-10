// cmd/test_ophim1/main.go
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/SonDuon/BACKEND_GOLANG_MUTI/internal/provider"
	"github.com/SonDuon/BACKEND_GOLANG_MUTI/internal/provider/ophim1"
	"github.com/SonDuon/BACKEND_GOLANG_MUTI/pkg/logger"
)

func main() {
	fmt.Println("🚀 Starting Ophim1 Adapter Test...")

	// 1. Khởi tạo Logger (dùng slog chuẩn Go 1.21+ hoặc logger của bạn)
	// _ = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	log := logger.New("development")
	// Nếu project bạn dùng custom logger, thay bằng: log := logger.New("development")

	// 2. Khởi tạo Adapter
	cfg := ophim1.DefaultConfig()
	cfg.Timeout = 10 * time.Second  // Test nhanh hơn
	adapter := ophim1.New(cfg, log) // ⚠️ Nếu New() yêu cầu *logger.Logger, hãy truyền instance logger của bạn

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// ─────────────────────────────────────
	// TEST 1: Health Check
	// ─────────────────────────────────────
	fmt.Println("\n📡 Test 1: IsAvailable()")
	if adapter.IsAvailable(ctx) {
		fmt.Println("✅ Ophim1 API is ONLINE")
	} else {
		fmt.Println("❌ Ophim1 API is OFFLINE or unreachable")
		return
	}

	// ─────────────────────────────────────
	// TEST 2: Search Movies
	// ─────────────────────────────────────
	fmt.Println("\n🔍 Test 2: Search(truc-ngoc')")
	result, err := adapter.Search(ctx, &provider.SearchParams{
		Keyword: "truc-ngoc",
		Limit:   3,
		Page:    1,
	})
	if err != nil {
		fmt.Printf("❌ Search failed: %v\n", err)
		return
	}
	fmt.Printf("✅ Found %d movies\n", len(result.Items))
	for i, m := range result.Items {
		fmt.Printf("   %d. %-40s | Slug: %s | Year: %d\n", i+1, m.Title, m.ExternalID, m.ReleaseYear)
	}

	// ─────────────────────────────────────
	// TEST 3: Get Streaming Links (Lấy link xem)
	// ─────────────────────────────────────
	if len(result.Items) > 0 {
		firstMovie := result.Items[0]
		fmt.Printf("\n🎬 Test 3: GetStreamingLinks('%s')\n", firstMovie.ExternalID)

		stream, err := adapter.GetStreamingLinks(ctx, firstMovie.ExternalID, "")
		if err != nil {
			fmt.Printf("❌ Get links failed: %v\n", err)
			return
		}

		fmt.Printf("✅ Title: %s\n", stream.Title)
		fmt.Printf("🔗 Found %d video sources:\n", len(stream.Sources))
		for _, src := range stream.Sources {
			fmt.Printf("   - [%s] %s | Quality: %s | Type: %s\n", src.Label, src.URL[:50]+"...", src.Quality, src.Type)
		}
		fmt.Printf("⏳ Expires at: %d (Unix timestamp)\n", stream.ExpiresAt)
	}

	fmt.Println("\n🎉 All tests passed! Adapter is working correctly.")
}
