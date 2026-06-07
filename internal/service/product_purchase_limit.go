package service

import "github.com/NexaCard/API/internal/models"

// normalizeMaxPurchaseQuantity 归一化商品单次购买数量上限。
func normalizeMaxPurchaseQuantity(value int) int {
	if value <= 0 {
		return 0
	}
	return value
}

// normalizeMinPurchaseQuantity 归一化商品单次购买数量下限。
func normalizeMinPurchaseQuantity(value int) int {
	if value <= 0 {
		return 0
	}
	return value
}

// productMaxPurchaseQuantity 返回商品当前有效的单次购买上限。
func productMaxPurchaseQuantity(product *models.Product) int {
	if product == nil {
		return 0
	}
	return normalizeMaxPurchaseQuantity(product.MaxPurchaseQuantity)
}

// productMinPurchaseQuantity 返回商品当前有效的单次购买下限。
func productMinPurchaseQuantity(product *models.Product) int {
	if product == nil {
		return 0
	}
	return normalizeMinPurchaseQuantity(product.MinPurchaseQuantity)
}

// validateProductPurchaseQuantity 校验单次购买数量是否在商品上下限内。
func validateProductPurchaseQuantity(product *models.Product, quantity int) error {
	if quantity <= 0 {
		return ErrInvalidOrderItem
	}
	if minLimit := productMinPurchaseQuantity(product); minLimit > 0 && quantity < minLimit {
		return ErrProductMinPurchaseNotMet
	}
	if maxLimit := productMaxPurchaseQuantity(product); maxLimit > 0 && quantity > maxLimit {
		return ErrProductMaxPurchaseExceeded
	}
	return nil
}
