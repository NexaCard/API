package public

import (
	"strconv"

	"github.com/NexaCard/API/internal/dto"
	"github.com/NexaCard/API/internal/http/handlers/shared"
	"github.com/NexaCard/API/internal/http/response"

	"github.com/gin-gonic/gin"
)

// GetPublicBanners 获取前台 Banner 列表
func (h *Handler) GetPublicBanners(c *gin.Context) {
	position := c.DefaultQuery("position", "home_hero")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}

	banners, err := h.BannerService.ListPublic(position, limit)
	if err != nil {
		shared.RespondError(c, response.CodeInternal, "error.banner_fetch_failed", err)
		return
	}

	response.Success(c, dto.NewBannerRespList(banners))
}
