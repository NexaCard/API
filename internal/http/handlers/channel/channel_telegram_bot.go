package channel

import (
	"time"

	"github.com/NexaCard/API/internal/logger"
	"github.com/NexaCard/API/internal/service"

	"github.com/gin-gonic/gin"
)

// GetBotConfig GET /api/v1/channel/telegram/config
// 返回 Telegram Bot 配置 + config_version（嵌套结构）
func (h *Handler) GetBotConfig(c *gin.Context) {
	config, err := h.SettingService.GetTelegramBotConfig()
	if err != nil {
		respondChannelError(c, 500, 500, "internal_error", "error.internal_error", err)
		return
	}

	// 从已认证的 channel client 获取 bot_token
	var botToken string
	if clientID, exists := c.Get("channel_client_id"); exists {
		if id, ok := clientID.(uint); ok {
			client, err := h.ChannelClientService.GetChannelClient(id)
			if err == nil && client != nil {
				botToken, _ = h.ChannelClientService.DecryptBotToken(client)
			}
		}
	}

	runtimeStatus, err := h.SettingService.GetTelegramBotRuntimeStatus()
	if err != nil {
		logger.Warnw("channel_get_runtime_status_failed", "error", err)
		runtimeStatus = &service.TelegramBotRuntimeStatusSetting{}
	}

	respondChannelSuccess(c, gin.H{
		"config":         service.SerializeTelegramBotConfigForChannel(*config, botToken),
		"config_version": runtimeStatus.ConfigVersion,
	})
}

type reportHeartbeatRequest struct {
	BotVersion       string   `json:"bot_version"`
	WebhookStatus    string   `json:"webhook_status"`
	MachineCode      string   `json:"machine_code"`
	LicenseStatus    string   `json:"license_status"`
	LicenseExpiresAt string   `json:"license_expires_at"`
	Warnings         []string `json:"warnings"`
}

// ReportHeartbeat POST /api/v1/channel/telegram/heartbeat
// Bot 上报心跳，更新 runtime_status
func (h *Handler) ReportHeartbeat(c *gin.Context) {
	var req reportHeartbeatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondChannelBindError(c, err)
		return
	}

	// 获取当前运行时状态以保留 config_version 等字段
	current, err := h.SettingService.GetTelegramBotRuntimeStatus()
	if err != nil {
		logger.Warnw("channel_heartbeat_get_status_failed", "error", err)
		current = &service.TelegramBotRuntimeStatusSetting{}
	}

	now := time.Now().UTC().Format(time.RFC3339)
	updated := service.TelegramBotRuntimeStatusSetting{
		Connected:        true,
		LastSeenAt:       now,
		BotVersion:       req.BotVersion,
		WebhookStatus:    req.WebhookStatus,
		MachineCode:      req.MachineCode,
		LicenseStatus:    req.LicenseStatus,
		LicenseExpiresAt: req.LicenseExpiresAt,
		Warnings:         append([]string(nil), req.Warnings...),
		ConfigVersion:    current.ConfigVersion,
		LastConfigSyncAt: current.LastConfigSyncAt,
	}

	if err := h.SettingService.UpdateTelegramBotRuntimeStatus(updated); err != nil {
		logger.Errorw("channel_heartbeat_update_failed", "error", err)
		respondChannelError(c, 500, 500, "internal_error", "error.internal_error", err)
		return
	}

	respondChannelSuccess(c, gin.H{"config_version": updated.ConfigVersion})
}
