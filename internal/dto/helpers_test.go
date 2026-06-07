package dto

import (
	"github.com/NexaCard/API/internal/models"
	"github.com/shopspring/decimal"
)

func newMoney(s string) models.Money {
	return models.NewMoneyFromDecimal(decimal.RequireFromString(s))
}
