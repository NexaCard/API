package repository

import (
	"errors"

	"github.com/NexaCard/API/internal/constants"
	"github.com/NexaCard/API/internal/models"

	"gorm.io/gorm"
)

// DownstreamOrderRefRepository 下游订单引用数据访问接口
type DownstreamOrderRefRepository interface {
	GetByID(id uint) (*models.DownstreamOrderRef, error)
	GetByOrderID(orderID uint) (*models.DownstreamOrderRef, error)
	GetByCredentialAndDownstreamNo(credentialID uint, downstreamOrderNo string) (*models.DownstreamOrderRef, error)
	Create(ref *models.DownstreamOrderRef) error
	Update(ref *models.DownstreamOrderRef) error
	ListPendingCallbacks(limit int) ([]models.DownstreamOrderRef, error)
	ListByCredentialID(credentialID uint, filter DownstreamOrderRefListFilter) ([]models.DownstreamOrderRef, int64, error)
}

// DownstreamOrderRefListFilter 下游订单引用列表筛选
type DownstreamOrderRefListFilter struct {
	CallbackStatus string
	Pagination
}

// GormDownstreamOrderRefRepository GORM 实现
type GormDownstreamOrderRefRepository struct {
	db *gorm.DB
}

// NewDownstreamOrderRefRepository 创建下游订单引用仓库
func NewDownstreamOrderRefRepository(db *gorm.DB) *GormDownstreamOrderRefRepository {
	return &GormDownstreamOrderRefRepository{db: db}
}

// GetByID 根据 ID 获取
func (r *GormDownstreamOrderRefRepository) GetByID(id uint) (*models.DownstreamOrderRef, error) {
	var ref models.DownstreamOrderRef
	if err := r.db.First(&ref, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &ref, nil
}

// GetByOrderID 根据订单 ID 获取
func (r *GormDownstreamOrderRefRepository) GetByOrderID(orderID uint) (*models.DownstreamOrderRef, error) {
	var ref models.DownstreamOrderRef
	if err := r.db.Where("order_id = ?", orderID).First(&ref).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &ref, nil
}

// GetByCredentialAndDownstreamNo 根据凭证 ID 和下游订单号查询（用于幂等性检查）
func (r *GormDownstreamOrderRefRepository) GetByCredentialAndDownstreamNo(credentialID uint, downstreamOrderNo string) (*models.DownstreamOrderRef, error) {
	if credentialID == 0 || downstreamOrderNo == "" {
		return nil, nil
	}
	var ref models.DownstreamOrderRef
	if err := r.db.Where("api_credential_id = ? AND downstream_order_no = ?", credentialID, downstreamOrderNo).First(&ref).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &ref, nil
}

// Create 创建下游订单引用
func (r *GormDownstreamOrderRefRepository) Create(ref *models.DownstreamOrderRef) error {
	return r.db.Create(ref).Error
}

// Update 更新下游订单引用
func (r *GormDownstreamOrderRefRepository) Update(ref *models.DownstreamOrderRef) error {
	return r.db.Save(ref).Error
}

// ListPendingCallbacks 获取待发送回调的记录
func (r *GormDownstreamOrderRefRepository) ListPendingCallbacks(limit int) ([]models.DownstreamOrderRef, error) {
	var refs []models.DownstreamOrderRef
	q := r.db.Where("callback_status = ? AND callback_url != ''", constants.CallbackStatusPending).
		Order("created_at ASC").
		Limit(limit)
	if err := q.Find(&refs).Error; err != nil {
		return nil, err
	}
	return refs, nil
}

// ListByCredentialID 根据凭证 ID 列表查询
func (r *GormDownstreamOrderRefRepository) ListByCredentialID(credentialID uint, filter DownstreamOrderRefListFilter) ([]models.DownstreamOrderRef, int64, error) {
	var refs []models.DownstreamOrderRef
	var total int64

	q := r.db.Model(&models.DownstreamOrderRef{}).Where("api_credential_id = ?", credentialID)
	if filter.CallbackStatus != "" {
		q = q.Where("callback_status = ?", filter.CallbackStatus)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	q = q.Order("created_at DESC")
	if filter.Page > 0 && filter.PageSize > 0 {
		q = q.Offset((filter.Page - 1) * filter.PageSize).Limit(filter.PageSize)
	}

	if err := q.Find(&refs).Error; err != nil {
		return nil, 0, err
	}
	return refs, total, nil
}
