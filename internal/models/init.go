package models

import (
	"strings"

	"github.com/NexaCard/API/internal/logger"

	"golang.org/x/crypto/bcrypt"
)

func normalizeBootstrapAdminUsername(username string) string {
	trimmed := strings.TrimSpace(username)
	if trimmed == "" {
		return "admin"
	}
	return trimmed
}

func InitDefaultAdmin(username, password string) error {
	bootstrapUsername := normalizeBootstrapAdminUsername(username)
	bootstrapPassword := strings.TrimSpace(password)

	var count int64
	DB.Model(&Admin{}).Count(&count)

	if count > 0 {
		if err := DB.Model(&Admin{}).Where("username = ?", bootstrapUsername).Update("is_super", true).Error; err != nil {
			logger.Warnw("ensure_default_admin_super_failed", "error", err)
		}
		return nil
	}

	if bootstrapPassword == "" {
		logger.Warnw("default_admin_skipped_without_password", "username", bootstrapUsername)
		return nil
	}

	username = bootstrapUsername
	hash, err := bcrypt.GenerateFromPassword([]byte(bootstrapPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	admin := Admin{
		Username:     username,
		PasswordHash: string(hash),
		IsSuper:      true,
	}

	if err := DB.Create(&admin).Error; err != nil {
		return err
	}

	logger.Warnw("default_admin_created", "username", username, "password_hidden", true)

	return nil
}
