package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/SonDuon/BACKEND_GOLANG_MUTI/internal/service"
	"github.com/gin-gonic/gin"
)

type MovieHandler struct {
	service *service.MovieService
}

func NewMovieHandler(svc *service.MovieService) *MovieHandler {
	return &MovieHandler{service: svc}
}

func (h *MovieHandler) RegisterRoutes(r *gin.RouterGroup) {
	movies := r.Group("/movies")
	{
		// 🎬 Core Endpoints
		movies.GET("/:slug", h.GetMovie)
		movies.GET("/:slug/watch", h.GetWatch)
		movies.GET("", h.ListMovies)
		movies.GET("/", h.ListMovies) // Hỗ trợ trailing slash
		movies.GET("/search", h.SearchMovies)

		// 🆕 Category & Year Filtering
		movies.GET("/genre/:slug", h.GetMoviesByGenre)
		movies.GET("/country/:slug", h.GetMoviesByCountry)
		movies.GET("/year/:year", h.GetMoviesByYear)
	}
}

// Helper: Chuẩn hóa params phân trang
func parsePagination(c *gin.Context) (int, int) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "24"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 24
	}
	return page, limit
}

// Helper: Trả về response phân trang chuẩn
func respondPagination(c *gin.Context, data interface{}, page, limit int, total int64) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    data,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
			"pages": (total + int64(limit) - 1) / int64(limit),
		},
	})
}

// ==========================================
// 📋 EXISTING HANDLERS
// ==========================================

func (h *MovieHandler) GetMovie(c *gin.Context) {
	slug := c.Param("slug")
	ctx := c.Request.Context()

	movie, err := h.service.GetMovieDetail(ctx, slug)
	if err != nil {
		fmt.Printf("❌ [GetMovie] Error: %v\n", err)
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Movie not found",
			"slug":    slug,
			"details": err.Error(),
		})
		return
	}

	fmt.Printf("✅ [GetMovie] Success for slug: '%s'\n", slug)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    movie,
	})
}

func (h *MovieHandler) GetWatch(c *gin.Context) {
	slug := c.Param("slug")
	episodeSlug := c.Query("episode")
	ctx := c.Request.Context()

	links, err := h.service.GetWatchLinks(ctx, slug, episodeSlug)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    links,
	})
}

func (h *MovieHandler) ListMovies(c *gin.Context) {
	page, limit := parsePagination(c)
	typeFilter := c.Query("type")
	statusFilter := c.Query("status")

	movies, total, err := h.service.ListMovies(c.Request.Context(), page, limit, typeFilter, statusFilter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	respondPagination(c, movies, page, limit, total)
}

func (h *MovieHandler) SearchMovies(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Thiếu tham số 'q' (từ khóa tìm kiếm)"})
		return
	}

	page, limit := parsePagination(c)
	movies, total, err := h.service.SearchMovies(c.Request.Context(), query, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	respondPagination(c, movies, page, limit, total)
}

// ==========================================
// 🆕 CATEGORY & YEAR FILTERING HANDLERS
// ==========================================

// GetMoviesByGenre: GET /movies/genre/:slug?year=2023&status=ongoing&page=1&limit=24
func (h *MovieHandler) GetMoviesByGenre(c *gin.Context) {
	slug := c.Param("slug")
	page, limit := parsePagination(c)

	yearStr := c.Query("year")
	year := 0
	if yearStr != "" {
		y, err := strconv.Atoi(yearStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid 'year' parameter"})
			return
		}
		year = y
	}

	status := c.Query("status")
	ctx := c.Request.Context()

	movies, total, err := h.service.GetMoviesByCategory(ctx, "genre", slug, page, limit, year, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	respondPagination(c, movies, page, limit, total)
}

// GetMoviesByCountry: GET /movies/country/:slug?year=2023&status=completed&page=1&limit=24
func (h *MovieHandler) GetMoviesByCountry(c *gin.Context) {
	slug := c.Param("slug")
	page, limit := parsePagination(c)

	yearStr := c.Query("year")
	year := 0
	if yearStr != "" {
		y, err := strconv.Atoi(yearStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid 'year' parameter"})
			return
		}
		year = y
	}

	status := c.Query("status")
	ctx := c.Request.Context()

	movies, total, err := h.service.GetMoviesByCategory(ctx, "country", slug, page, limit, year, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	respondPagination(c, movies, page, limit, total)
}

// GetMoviesByYear: GET /movies/year/:year?genre=hanh-dong&status=ongoing&page=1&limit=24
func (h *MovieHandler) GetMoviesByYear(c *gin.Context) {
	yearStr := c.Param("year")
	year, err := strconv.Atoi(yearStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid 'year' path parameter"})
		return
	}

	page, limit := parsePagination(c)
	status := c.Query("status")
	genreSlug := c.Query("genre") // Hỗ trợ lọc kết hợp: Năm + Thể loại
	ctx := c.Request.Context()

	movies, total, err := h.service.GetMoviesByYear(ctx, year, page, limit, status, genreSlug)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	respondPagination(c, movies, page, limit, total)
}
