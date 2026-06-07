package dto

import (
	"time"

	"github.com/NexaCard/API/internal/models"
)

// LoginLogResp 登录日志响应
type LoginLogResp struct {
	ID          uint      `json:"id"`
	Email       string    `json:"email"`
	Status      string    `json:"status"`
	ClientIP    string    `json:"client_ip"`
	UserAgent   string    `json:"user_agent"`
	LoginSource string    `json:"login_source"`
	CreatedAt   time.Time `json:"created_at"`
}

// NewLoginLogResp 从 models.UserLoginLog 构造响应
func NewLoginLogResp(l *models.UserLoginLog) LoginLogResp {
	return LoginLogResp{
		ID:          l.ID,
		Email:       l.Email,
		Status:      l.Status,
		ClientIP:    l.ClientIP,
		UserAgent:   l.UserAgent,
		LoginSource: l.LoginSource,
		CreatedAt:   l.CreatedAt,
	}
	// 排除：UserID、FailReason、RequestID
}

// NewLoginLogRespList 批量转换登录日志
func NewLoginLogRespList(logs []models.UserLoginLog) []LoginLogResp {
	result := make([]LoginLogResp, 0, len(logs))
	for i := range logs {
		result = append(result, NewLoginLogResp(&logs[i]))
	}
	return result
}
