package public

import (
	"context"
	"errors"

	"github.com/NexaCard/API/internal/cache"
	"github.com/NexaCard/API/internal/constants"
	"github.com/NexaCard/API/internal/dto"
	"github.com/NexaCard/API/internal/http/handlers/shared"
	"github.com/NexaCard/API/internal/http/response"
	"github.com/NexaCard/API/internal/service"

	"github.com/gin-gonic/gin"
)

// GetUser2FAStatus 当前用户 2FA 状态
func (h *Handler) GetUser2FAStatus(c *gin.Context) {
	uid, ok := shared.GetUserID(c)
	if !ok {
		return
	}
	st, err := h.UserTOTPService.GetStatus(uid)
	if err != nil {
		shared.RespondError(c, response.CodeInternal, "error.internal_error", err)
		return
	}
	response.Success(c, st)
}

// SetupUser2FA 开始绑定，返回 secret + otpauth url
func (h *Handler) SetupUser2FA(c *gin.Context) {
	uid, ok := shared.GetUserID(c)
	if !ok {
		return
	}
	res, err := h.UserTOTPService.Setup(uid)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrTOTPAlreadyEnabled):
			shared.RespondError(c, response.CodeBadRequest, "error.totp_already_enabled", nil)
		case errors.Is(err, service.ErrNotFound):
			shared.RespondError(c, response.CodeNotFound, "error.user_not_found", nil)
		default:
			shared.RespondError(c, response.CodeInternal, "error.internal_error", err)
		}
		return
	}
	response.Success(c, res)
}

// EnableUser2FARequest 启用请求
type EnableUser2FARequest struct {
	Code string `json:"code" binding:"required"`
}

// EnableUser2FA 完成绑定
func (h *Handler) EnableUser2FA(c *gin.Context) {
	uid, ok := shared.GetUserID(c)
	if !ok {
		return
	}
	var req EnableUser2FARequest
	if err := c.ShouldBindJSON(&req); err != nil {
		shared.RespondBindError(c, err)
		return
	}
	res, err := h.UserTOTPService.Enable(uid, req.Code)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrTOTPAlreadyEnabled):
			shared.RespondError(c, response.CodeBadRequest, "error.totp_already_enabled", nil)
		case errors.Is(err, service.ErrTOTPPendingExpired):
			shared.RespondError(c, response.CodeBadRequest, "error.totp_pending_expired", nil)
		case errors.Is(err, service.ErrTOTPCodeInvalid):
			shared.RespondError(c, response.CodeBadRequest, "error.totp_code_invalid", nil)
		case errors.Is(err, service.ErrTOTPTooManyAttempts):
			shared.RespondError(c, response.CodeBadRequest, "error.totp_too_many_attempts", nil)
		default:
			shared.RespondError(c, response.CodeInternal, "error.internal_error", err)
		}
		return
	}
	// 启用 2FA 时 TokenVersion++，原 token 已失效；签发新 token 让当前 session 继续可用
	loginRes, signErr := h.UserAuthService.CompleteLoginAfter2FA(uid, false)
	if signErr == nil {
		res.Token = loginRes.Token
		res.ExpiresAt = loginRes.ExpiresAt
	}
	response.Success(c, res)
}

// DisableUser2FARequest 关闭请求
type DisableUser2FARequest struct {
	Code         string `json:"code"`
	RecoveryCode string `json:"recovery_code"`
}

// DisableUser2FA 关闭 2FA
func (h *Handler) DisableUser2FA(c *gin.Context) {
	uid, ok := shared.GetUserID(c)
	if !ok {
		return
	}
	var req DisableUser2FARequest
	if err := c.ShouldBindJSON(&req); err != nil {
		shared.RespondBindError(c, err)
		return
	}
	if req.Code == "" && req.RecoveryCode == "" {
		shared.RespondError(c, response.CodeBadRequest, "error.totp_code_required", nil)
		return
	}
	isRecovery := req.RecoveryCode != ""
	codeArg := req.Code
	if isRecovery {
		codeArg = req.RecoveryCode
	}
	if err := h.UserTOTPService.Disable(uid, codeArg, isRecovery); err != nil {
		switch {
		case errors.Is(err, service.ErrTOTPNotEnabled):
			shared.RespondError(c, response.CodeBadRequest, "error.totp_not_enabled", nil)
		case errors.Is(err, service.ErrTOTPCodeInvalid):
			shared.RespondError(c, response.CodeBadRequest, "error.totp_code_invalid", nil)
		case errors.Is(err, service.ErrTOTPRecoveryInvalid):
			shared.RespondError(c, response.CodeBadRequest, "error.recovery_code_invalid", nil)
		default:
			shared.RespondError(c, response.CodeInternal, "error.internal_error", err)
		}
		return
	}
	response.Success(c, nil)
}

// RegenerateUser2FARecoveryCodesRequest 重新生成请求
type RegenerateUser2FARecoveryCodesRequest struct {
	Code string `json:"code" binding:"required"`
}

// RegenerateUser2FARecoveryCodes 重新生成恢复码
func (h *Handler) RegenerateUser2FARecoveryCodes(c *gin.Context) {
	uid, ok := shared.GetUserID(c)
	if !ok {
		return
	}
	var req RegenerateUser2FARecoveryCodesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		shared.RespondBindError(c, err)
		return
	}
	codes, err := h.UserTOTPService.RegenerateRecoveryCodes(uid, req.Code)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrTOTPNotEnabled):
			shared.RespondError(c, response.CodeBadRequest, "error.totp_not_enabled", nil)
		case errors.Is(err, service.ErrTOTPCodeInvalid):
			shared.RespondError(c, response.CodeBadRequest, "error.totp_code_invalid", nil)
		default:
			shared.RespondError(c, response.CodeInternal, "error.internal_error", err)
		}
		return
	}
	response.Success(c, gin.H{"recovery_codes": codes})
}

// VerifyUser2FARequest 第二步登录请求
type VerifyUser2FARequest struct {
	ChallengeToken string `json:"challenge_token" binding:"required"`
	Code           string `json:"code"`
	RecoveryCode   string `json:"recovery_code"`
}

// VerifyUser2FA 完成两步登录
func (h *Handler) VerifyUser2FA(c *gin.Context) {
	var req VerifyUser2FARequest
	if err := c.ShouldBindJSON(&req); err != nil {
		shared.RespondBindError(c, err)
		return
	}
	if req.Code == "" && req.RecoveryCode == "" {
		shared.RespondError(c, response.CodeBadRequest, "error.totp_code_required", nil)
		return
	}
	claims, err := h.UserAuthService.ParseUserChallengeToken(req.ChallengeToken)
	if err != nil {
		h.recordUserLogin(c, "", 0, constants.LoginLogStatusFailed, constants.LoginLogFailReasonChallengeInvalid, constants.LoginLogSourceWeb)
		shared.RespondError(c, response.CodeUnauthorized, "error.totp_challenge_invalid", nil)
		return
	}
	ctx := context.Background()
	rdb := cache.Client()
	if rdb != nil {
		if v, _ := rdb.Exists(ctx, service.UserChallengeRevokedKey(claims.JTI)).Result(); v == 1 {
			h.recordUserLogin(c, "", claims.UserID, constants.LoginLogStatusFailed, constants.LoginLogFailReasonChallengeInvalid, constants.LoginLogSourceWeb)
			shared.RespondError(c, response.CodeUnauthorized, "error.totp_challenge_invalid", nil)
			return
		}
	}

	verifyErr := h.verifyUserChallengeAttempt(claims.UserID, req.Code, req.RecoveryCode)
	email := ""
	if u, _ := h.UserRepo.GetByID(claims.UserID); u != nil {
		email = u.Email
	}
	if verifyErr != nil {
		failCnt := h.bumpUserChallengeFails(ctx, claims.JTI)
		var failReason string
		switch {
		case errors.Is(verifyErr, service.ErrTOTPRecoveryInvalid):
			failReason = constants.LoginLogFailReasonInvalidRecoveryCode
		case errors.Is(verifyErr, service.ErrTOTPCodeInvalid):
			failReason = constants.LoginLogFailReasonInvalidTOTPCode
		default:
			failReason = constants.LoginLogFailReasonInternalError
		}
		h.recordUserLogin(c, email, claims.UserID, constants.LoginLogStatusFailed, failReason, constants.LoginLogSourceWeb)
		if failCnt >= service.UserChallengeMaxFailures {
			h.revokeUserChallenge(ctx, claims.JTI)
			shared.RespondError(c, response.CodeUnauthorized, "error.totp_too_many_attempts", nil)
			return
		}
		switch {
		case errors.Is(verifyErr, service.ErrTOTPCodeInvalid):
			shared.RespondError(c, response.CodeUnauthorized, "error.totp_code_invalid", nil)
		case errors.Is(verifyErr, service.ErrTOTPRecoveryInvalid):
			shared.RespondError(c, response.CodeUnauthorized, "error.recovery_code_invalid", nil)
		default:
			shared.RespondError(c, response.CodeUnauthorized, "error.totp_challenge_invalid", nil)
		}
		return
	}
	h.revokeUserChallenge(ctx, claims.JTI)
	loginRes, err := h.UserAuthService.CompleteLoginAfter2FA(claims.UserID, claims.RememberMe)
	if err != nil {
		shared.RespondError(c, response.CodeInternal, "error.login_failed", err)
		return
	}
	h.recordUserLogin(c, loginRes.User.Email, loginRes.User.ID, constants.LoginLogStatusSuccess, "", constants.LoginLogSourceWeb)
	response.Success(c, gin.H{
		"requires_totp": false,
		"user":          dto.NewUserAuthBriefResp(loginRes.User),
		"token":         loginRes.Token,
		"expires_at":    loginRes.ExpiresAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

func (h *Handler) verifyUserChallengeAttempt(userID uint, code, recoveryCode string) error {
	if recoveryCode != "" {
		return h.UserTOTPService.VerifyChallengeRecoveryCode(userID, recoveryCode)
	}
	return h.UserTOTPService.VerifyChallengeCode(userID, code)
}

func (h *Handler) bumpUserChallengeFails(ctx context.Context, jti string) int64 {
	rdb := cache.Client()
	if rdb == nil {
		return 0
	}
	cnt, err := rdb.Incr(ctx, service.UserChallengeFailKey(jti)).Result()
	if err == nil && cnt == 1 {
		_ = rdb.Expire(ctx, service.UserChallengeFailKey(jti), service.UserChallengeTTL).Err()
	}
	return cnt
}

func (h *Handler) revokeUserChallenge(ctx context.Context, jti string) {
	rdb := cache.Client()
	if rdb == nil {
		return
	}
	_ = rdb.Set(ctx, service.UserChallengeRevokedKey(jti), "1", service.UserChallengeTTL).Err()
}
