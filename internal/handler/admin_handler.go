package handler

import (
	"net/http"

	"github.com/SonDuon/BACKEND_GOLANG_MUTI/internal/provider"
	"github.com/SonDuon/BACKEND_GOLANG_MUTI/internal/repository"
	"github.com/gin-gonic/gin"
)

// AdminHandler xử lý các request từ Admin Panel
type AdminHandler struct {
	repo     repository.MovieRepository // Để lưu phim vào DB
	provider provider.MovieProvider     // Để gọi API Ophim1
}

// NewAdminHandler: Constructor inject dependencies
func NewAdminHandler(
	repo repository.MovieRepository,
	provider provider.MovieProvider,
) *AdminHandler {
	return &AdminHandler{
		repo:     repo,
		provider: provider,
	}
}

// RegisterRoutes: Đăng ký các route admin vào Gin router
func (h *AdminHandler) RegisterRoutes(r *gin.RouterGroup) {
	admin := r.Group("/admin")
	{
     // 1. Health check các nguồn phim
        admin.GET("/providers/health", h.CheckProviders)

        // // 2. Admin upload phim tự-host (self-hosted)
        // admin.POST("/self-hosted/upload", h.UploadSelfHostedMovie)

        // // 3. Admin sửa thủ công metadata/link bị lỗi từ API ngoài
        // admin.PUT("/movies/:slug/override", h.OverrideMovieData)

        // // 4. Xóa cache (Redis/DB) khi cần sync lại
        // admin.POST("/cache/purge", h.PurgeCache)

        // // 5. Thống kê hệ thống (số phim lazy-import, request/day, v.v.)
        // admin.GET("/stats", h.GetSystemStats)
	}
}

// api/admin/providers/health
func (h *AdminHandler) CheckProviders(c *gin.Context) {
	// Nếu bạn có Provider Manager, gọi manager.HealthCheckAll()
	// Tạm thời return static cho ophim1
	ctx := c.Request.Context()

	available := h.provider.IsAvailable(ctx)

	c.JSON(http.StatusOK, gin.H{
		"providers": gin.H{
			"ophim1": gin.H{
				"available": available,
				"priority":  h.provider.Priority(),
			},
		},
	})
}
