package admin

import (
	"testing"

	"github.com/NexaCard/API/internal/models"
)

func TestPaymentChannelAdminResponseMasksConfigSecrets(t *testing.T) {
	channel := &models.PaymentChannel{
		ConfigJSON: models.JSON{
			"gateway_url":    "https://pay.example",
			"merchant_key":   "secret-value",
			"webhook_secret": "whsec-value",
		},
	}

	resp := paymentChannelAdminResponse(channel)
	config, ok := resp["config_json"].(models.JSON)
	if !ok {
		t.Fatalf("config_json should be models.JSON")
	}
	if got := config["merchant_key"]; got != "" {
		t.Fatalf("merchant_key should be masked, got %v", got)
	}
	if got := config["webhook_secret"]; got != "" {
		t.Fatalf("webhook_secret should be masked, got %v", got)
	}
	if got := config["gateway_url"]; got != "https://pay.example" {
		t.Fatalf("gateway_url should be preserved, got %v", got)
	}

	meta, ok := resp["config_secrets"].(map[string]bool)
	if !ok {
		t.Fatalf("config_secrets should be map[string]bool")
	}
	if meta["merchant_key"] != true || meta["webhook_secret"] != true {
		t.Fatalf("config_secrets should mark existing secrets, got %#v", meta)
	}
}

func TestMergePaymentChannelConfigForUpdateKeepsSubmittedBlankSecrets(t *testing.T) {
	existing := models.JSON{
		"gateway_url":  "https://old.example",
		"merchant_key": "old-secret",
	}
	submitted := map[string]interface{}{
		"gateway_url":  "https://new.example",
		"merchant_key": "",
	}

	next := mergePaymentChannelConfigForUpdate(existing, submitted, true)
	if got := next["gateway_url"]; got != "https://new.example" {
		t.Fatalf("gateway_url should update, got %v", got)
	}
	if got := next["merchant_key"]; got != "old-secret" {
		t.Fatalf("blank merchant_key should keep old secret, got %v", got)
	}
}

func TestMergePaymentChannelConfigForUpdateAllowsExplicitSecretRemoval(t *testing.T) {
	existing := models.JSON{
		"gateway_url":  "https://old.example",
		"merchant_key": "old-secret",
	}
	submitted := map[string]interface{}{
		"gateway_url": "https://new.example",
	}

	next := mergePaymentChannelConfigForUpdate(existing, submitted, true)
	if _, ok := next["merchant_key"]; ok {
		t.Fatalf("omitted merchant_key should remain omitted")
	}
}

func TestMergePaymentChannelConfigForUpdateDoesNotKeepSecretsAcrossProviderChange(t *testing.T) {
	existing := models.JSON{
		"merchant_key": "old-secret",
	}
	submitted := map[string]interface{}{
		"merchant_key": "",
	}

	next := mergePaymentChannelConfigForUpdate(existing, submitted, false)
	if got := next["merchant_key"]; got != "" {
		t.Fatalf("provider change should not keep old secret, got %v", got)
	}
}
