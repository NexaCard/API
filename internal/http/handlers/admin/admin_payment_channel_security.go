package admin

import (
	"strings"

	"github.com/NexaCard/API/internal/models"
)

const paymentSecretMaskValue = "********"

var paymentChannelSensitiveConfigKeys = map[string]struct{}{
	"api_secret":           {},
	"api_v3_key":           {},
	"auth_token":           {},
	"client_secret":        {},
	"merchant_key":         {},
	"merchant_private_key": {},
	"merchant_token":       {},
	"notify_secret":        {},
	"private_key":          {},
	"secret_key":           {},
	"webhook_secret":       {},
}

func paymentChannelAdminResponse(channel *models.PaymentChannel) map[string]interface{} {
	if channel == nil {
		return nil
	}
	maskedConfig, secretMeta := maskPaymentChannelConfig(channel.ConfigJSON)
	return map[string]interface{}{
		"id":                    channel.ID,
		"name":                  channel.Name,
		"icon":                  channel.Icon,
		"provider_type":         channel.ProviderType,
		"channel_type":          channel.ChannelType,
		"interaction_mode":      channel.InteractionMode,
		"fee_rate":              channel.FeeRate,
		"fixed_fee":             channel.FixedFee,
		"min_amount":            channel.MinAmount,
		"max_amount":            channel.MaxAmount,
		"hide_amount_out_range": channel.HideAmountOutRange,
		"payment_roles":         channel.PaymentRoles,
		"member_levels":         channel.MemberLevels,
		"payment_types":         channel.PaymentTypes,
		"config_json":           maskedConfig,
		"config_secrets":        secretMeta,
		"is_active":             channel.IsActive,
		"sort_order":            channel.SortOrder,
		"created_at":            channel.CreatedAt,
		"updated_at":            channel.UpdatedAt,
	}
}

func paymentChannelAdminListResponse(channels []models.PaymentChannel) []map[string]interface{} {
	items := make([]map[string]interface{}, 0, len(channels))
	for i := range channels {
		items = append(items, paymentChannelAdminResponse(&channels[i]))
	}
	return items
}

func maskPaymentChannelConfig(config models.JSON) (models.JSON, map[string]bool) {
	masked := make(models.JSON, len(config))
	secretMeta := map[string]bool{}
	for key, value := range config {
		if isPaymentChannelSensitiveConfigKey(key) {
			secretMeta[key] = !isBlankConfigValue(value)
			masked[key] = ""
			continue
		}
		masked[key] = value
	}
	return masked, secretMeta
}

func mergePaymentChannelConfigForUpdate(existing models.JSON, submitted map[string]interface{}, preserveSecrets bool) models.JSON {
	next := make(models.JSON, len(submitted))
	for key, value := range submitted {
		next[key] = value
	}
	if !preserveSecrets {
		return next
	}
	for key, existingValue := range existing {
		if !isPaymentChannelSensitiveConfigKey(key) {
			continue
		}
		submittedValue, ok := next[key]
		if ok && isMaskedOrBlankSecretValue(submittedValue) {
			next[key] = existingValue
		}
	}
	return next
}

func samePaymentChannelProvider(currentProvider, currentChannel, nextProvider, nextChannel string) bool {
	return strings.EqualFold(strings.TrimSpace(currentProvider), strings.TrimSpace(nextProvider)) &&
		strings.EqualFold(strings.TrimSpace(currentChannel), strings.TrimSpace(nextChannel))
}

func isPaymentChannelSensitiveConfigKey(key string) bool {
	_, ok := paymentChannelSensitiveConfigKeys[strings.ToLower(strings.TrimSpace(key))]
	return ok
}

func isMaskedOrBlankSecretValue(value interface{}) bool {
	if isBlankConfigValue(value) {
		return true
	}
	text, ok := value.(string)
	if !ok {
		return false
	}
	trimmed := strings.TrimSpace(text)
	if trimmed == paymentSecretMaskValue {
		return true
	}
	return strings.Trim(trimmed, "*") == ""
}

func isBlankConfigValue(value interface{}) bool {
	if value == nil {
		return true
	}
	text, ok := value.(string)
	return ok && strings.TrimSpace(text) == ""
}
