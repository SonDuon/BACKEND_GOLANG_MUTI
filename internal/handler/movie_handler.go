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
		// 🎬 Endpoint để lấy chi tiết phim (sẽ tự động import nếu chưa có trong DB)
		movies.GET("/:slug", h.GetMovie)
		// 🎬 Endpoint để lấy link xem phim (có thể có query param episode=slug-episode để lấy link tập cụ thể)
		movies.GET("/:slug/watch", h.GetWatch)
		// 📋 List movies
		movies.GET("", h.ListMovies)
		movies.GET("/", h.ListMovies)
	}
}

// Handler cho detail endpoint
func (h *MovieHandler) GetMovie(c *gin.Context) {
	slug := c.Param("slug")

	// 🔥 Log để debug
	fmt.Printf("🎯 [GetMovie] Received request for slug: '%s'\n", slug)

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

// Handler cho watch endpoint
func (h *MovieHandler) GetWatch(c *gin.Context) {
	slug := c.Param("slug")
	episodeSlug := c.DefaultQuery("episode", "full")

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

// ListMovies: GET /api/v1/movies?page=1&limit=20
func (h *MovieHandler) ListMovies(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	ctx := c.Request.Context()
	movies, total, err := h.service.GetMovieList(ctx, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    movies,
		"total":   total,
		"page":    page,
		"limit":   limit,
	})
}
