package admin

import (
	"errors"

	"github.com/NexaCard/API/internal/constants"
	"github.com/NexaCard/API/internal/http/handlers/shared"
	"github.com/NexaCard/API/internal/http/response"
	"github.com/NexaCard/API/internal/i18n"
	"github.com/NexaCard/API/internal/service"

	"github.com/gin-gonic/gin"
)

// LoginRequest 登录请求
type LoginRequest struct {
	Username       string                       `json:"username" binding:"required"`
	Password       string                       `json:"password" binding:"required"`
	CaptchaPayload shared.CaptchaPayloadRequest `json:"captcha_payload"`
}

// AdminLogin 管理员登录
func (h *Handler) AdminLogin(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		shared.RespondBindError(c, err)
		return
	}

	if h.CaptchaService != nil {
		if captchaErr := h.CaptchaService.Verify(constants.CaptchaSceneLogin, req.CaptchaPayload.ToServicePayload(), c.ClientIP()); captchaErr != nil {
			switch {
			case errors.Is(captchaErr, service.ErrCaptchaRequired):
				shared.RespondError(c, response.CodeBadRequest, "error.captcha_required", nil)
				return
			case errors.Is(captchaErr, service.ErrCaptchaInvalid):
				shared.RespondError(c, response.CodeBadRequest, "error.captcha_invalid", nil)
				return
			case errors.Is(captchaErr, service.ErrCaptchaConfigInvalid):
				shared.RespondError(c, response.CodeInternal, "error.captcha_config_invalid", captchaErr)
				return
			default:
				shared.RespondError(c, response.CodeInternal, "error.captcha_verify_failed", captchaErr)
				return
			}
		}
	}

	loginRes, err := h.AuthService.Login(req.Username, req.Password)
	if err != nil {
		failReason := constants.AdminLoginFailInvalidCredentials
		if !errors.Is(err, service.ErrInvalidCredentials) {
			failReason = constants.AdminLoginFailInternal
		}
		h.writeAdminLoginLog(c, 0, req.Username, constants.AdminLoginEventLoginPassword, constants.AdminLoginStatusFailed, failReason, nil)
		if errors.Is(err, service.ErrInvalidCredentials) {
			shared.RespondError(c, response.CodeUnauthorized, "error.admin_login_invalid", nil)
			return
		}
		shared.RespondError(c, response.CodeInternal, "error.login_failed", err)
		return
	}

	if loginRes.RequiresTOTP {
		h.writeAdminLoginLog(c, loginRes.Admin.ID, loginRes.Admin.Username, constants.AdminLoginEventLoginPassword, constants.AdminLoginStatusSuccess, "", nil)
		response.Success(c, gin.H{
			"requires_totp":        true,
			"challenge_token":      loginRes.ChallengeToken,
			"challenge_expires_at": loginRes.ChallengeExpiresAt.Format("2006-01-02T15:04:05Z07:00"),
		})
		return
	}

	h.writeAdminLoginLog(c, loginRes.Admin.ID, loginRes.Admin.Username, constants.AdminLoginEventLoginPassword, constants.AdminLoginStatusSuccess, "", nil)
	response.Success(c, gin.H{
		"requires_totp": false,
		"token":         loginRes.Token,
		"user": gin.H{
			"id":       loginRes.Admin.ID,
			"username": loginRes.Admin.Username,
		},
		"expires_at": loginRes.ExpiresAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

// UpdatePasswordRequest 修改密码请求
type UpdatePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

// UpdateAdminPassword 修改管理员密码
func (h *Handler) UpdateAdminPassword(c *gin.Context) {
	// 获取当前登录用户 ID
	id, ok := shared.GetAdminID(c)
	if !ok {
		return
	}

	var req UpdatePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		shared.RespondBindError(c, err)
		return
	}

	if err := h.AuthService.ChangePassword(id, req.OldPassword, req.NewPassword); err != nil {
		if errors.Is(err, service.ErrInvalidPassword) {
			shared.RespondError(c, response.CodeBadRequest, "error.password_old_invalid", nil)
			return
		}
		if errors.Is(err, service.ErrWeakPassword) {
			locale := i18n.ResolveLocale(c)
			if perr, ok := err.(interface {
				Key() string
				Args() []interface{}
			}); ok {
				msg := i18n.Sprintf(locale, perr.Key(), perr.Args()...)
				shared.RespondErrorWithMsg(c, response.CodeBadRequest, msg, nil)
				return
			}
			shared.RespondError(c, response.CodeBadRequest, "error.password_weak", nil)
			return
		}
		if errors.Is(err, service.ErrNotFound) {
			shared.RespondError(c, response.CodeNotFound, "error.user_not_found", nil)
			return
		}
		shared.RespondError(c, response.CodeInternal, "error.save_failed", err)
		return
	}

	response.Success(c, nil)
}
