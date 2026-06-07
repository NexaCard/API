package repository

import (
	"errors"

	"github.com/NexaCard/API/internal/models"

	"gorm.io/gorm"
)

// TelegramBroadcastListFilter Telegram 广播列表筛选。
type TelegramBroadcastListFilter struct {
	Page     int
	PageSize int
}

// TelegramBroadcastRepository Telegram 广播仓储接口。
type TelegramBroadcastRepository interface {
	Create(broadcast *models.TelegramBroadcast) error
	GetByID(id uint) (*models.TelegramBroadcast, error)
	List(filter TelegramBroadcastListFilter) ([]models.TelegramBroadcast, int64, error)
	Update(broadcast *models.TelegramBroadcast) error
}

// GormTelegramBroadcastRepository GORM 实现。
type GormTelegramBroadcastRepository struct {
	db *gorm.DB
}

// NewTelegramBroadcastRepository 创建 Telegram 广播仓储。
func NewTelegramBroadcastRepository(db *gorm.DB) *GormTelegramBroadcastRepository {
	return &GormTelegramBroadcastRepository{db: db}
}

// Create 创建广播记录。
func (r *GormTelegramBroadcastRepository) Create(broadcast *models.TelegramBroadcast) error {
	if broadcast == nil {
		return nil
	}
	return r.db.Create(broadcast).Error
}

// GetByID 按 ID 获取广播记录。
func (r *GormTelegramBroadcastRepository) GetByID(id uint) (*models.TelegramBroadcast, error) {
	if id == 0 {
		return nil, nil
	}
	var broadcast models.TelegramBroadcast
	if err := r.db.First(&broadcast, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &broadcast, nil
}

// List 获取广播记录列表。
func (r *GormTelegramBroadcastRepository) List(filter TelegramBroadcastListFilter) ([]models.TelegramBroadcast, int64, error) {
	query := r.db.Model(&models.TelegramBroadcast{})

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if filter.PageSize > 0 {
		query = applyPagination(query, filter.Page, filter.PageSize)
	}

	var items []models.TelegramBroadcast
	if err := query.Order("created_at DESC").Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

// Update 更新广播记录。
func (r *GormTelegramBroadcastRepository) Update(broadcast *models.TelegramBroadcast) error {
	if broadcast == nil {
		return nil
	}
	return r.db.Save(broadcast).Error
}
