package admin

import (
	"errors"

	"github.com/NexaCard/API/internal/http/handlers/shared"
	"github.com/NexaCard/API/internal/http/response"
	"github.com/NexaCard/API/internal/logger"
	"github.com/NexaCard/API/internal/service"

	"github.com/gin-gonic/gin"
)

// ResetUser2FA 管理员重置目标用户 2FA。
// 用户丢失 TOTP 设备和所有恢复码时由管理员协助解绑。
// DELETE /admin/users/:id/2fa
func (h *Handler) ResetUser2FA(c *gin.Context) {
	operatorID, ok := shared.GetAdminID(c)
	if !ok {
		return
	}
	userID, err := shared.ParseParamUint(c, "id")
	if err != nil {
		shared.RespondError(c, response.CodeBadRequest, "error.user_id_invalid", nil)
		return
	}
	target, err := h.UserTOTPService.AdminResetUser2FA(operatorID, userID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrNotFound):
			shared.RespondError(c, response.CodeNotFound, "error.user_not_found", nil)
		case errors.Is(err, service.ErrTOTPNotEnabled):
			shared.RespondError(c, response.CodeBadRequest, "error.totp_not_enabled", nil)
		default:
			shared.RespondError(c, response.CodeInternal, "error.internal_error", err)
		}
		return
	}

	// 留痕：高权限破坏性动作，写结构化日志，至少包含 operator / target / IP / RequestID
	requestID, _ := c.Get("request_id")
	rid, _ := requestID.(string)
	logger.Warnw("admin_reset_user_2fa",
		"operator_admin_id", operatorID,
		"target_user_id", target.ID,
		"target_email", target.Email,
		"client_ip", c.ClientIP(),
		"user_agent", c.Request.UserAgent(),
		"request_id", rid,
	)

	response.Success(c, nil)
}
