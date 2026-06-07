package repository

import (
	"errors"
	"strings"

	"github.com/NexaCard/API/internal/models"

	"gorm.io/gorm"
)

// PostRepository 文章数据访问接口
type PostRepository interface {
	List(filter PostListFilter) ([]models.Post, int64, error)
	GetBySlug(slug string, onlyPublished bool) (*models.Post, error)
	GetByID(id string) (*models.Post, error)
	Create(post *models.Post) error
	Update(post *models.Post) error
	Delete(id string) error
	CountBySlug(slug string, excludeID *string) (int64, error)
	GetRelatedProductIDs(postID uint) ([]uint, error)
	SetRelatedProductIDs(postID uint, productIDs []uint) error
	ListRelatedProducts(postID uint) ([]models.Product, error)
	ListPostsForProduct(productID uint, postType string, onlyPublished bool, limit int) ([]models.Post, error)
}

// GormPostRepository GORM 实现
type GormPostRepository struct {
	db *gorm.DB
}

// NewPostRepository 创建文章仓库
func NewPostRepository(db *gorm.DB) *GormPostRepository {
	return &GormPostRepository{db: db}
}

// List 文章列表
func (r *GormPostRepository) List(filter PostListFilter) ([]models.Post, int64, error) {
	var posts []models.Post
	query := r.db.Model(&models.Post{})

	if filter.OnlyPublished {
		query = query.Where("is_published = ?", true)
	}
	if filter.Type != "" {
		query = query.Where("type = ?", filter.Type)
	}
	if search := strings.TrimSpace(filter.Search); search != "" {
		like := "%" + search + "%"
		condition, argCount := buildLocalizedLikeCondition(r.db, []string{"slug"}, []string{"title_json"})
		query = query.Where(condition, repeatLikeArgs(like, argCount)...)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	query = applyPagination(query, filter.Page, filter.PageSize)

	orderBy := filter.OrderBy
	if orderBy == "" {
		orderBy = "created_at DESC"
	}

	if err := query.Order(orderBy).Find(&posts).Error; err != nil {
		return nil, 0, err
	}
	return posts, total, nil
}

// GetBySlug 根据 slug 获取文章
func (r *GormPostRepository) GetBySlug(slug string, onlyPublished bool) (*models.Post, error) {
	query := r.db.Where("slug = ?", slug)
	if onlyPublished {
		query = query.Where("is_published = ?", true)
	}

	var post models.Post
	if err := query.First(&post).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &post, nil
}

// GetByID 根据 ID 获取文章
func (r *GormPostRepository) GetByID(id string) (*models.Post, error) {
	var post models.Post
	if err := r.db.First(&post, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &post, nil
}

// Create 创建文章
func (r *GormPostRepository) Create(post *models.Post) error {
	return r.db.Create(post).Error
}

// Update 更新文章
func (r *GormPostRepository) Update(post *models.Post) error {
	return r.db.Save(post).Error
}

// Delete 删除文章
func (r *GormPostRepository) Delete(id string) error {
	return r.db.Delete(&models.Post{}, id).Error
}

// CountBySlug 统计 slug 数量
func (r *GormPostRepository) CountBySlug(slug string, excludeID *string) (int64, error) {
	var count int64
	query := r.db.Model(&models.Post{}).Where("slug = ?", slug)
	if excludeID != nil {
		query = query.Where("id != ?", *excludeID)
	}
	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// GetRelatedProductIDs 获取文章关联的商品 ID 列表（按 sort 升序）
func (r *GormPostRepository) GetRelatedProductIDs(postID uint) ([]uint, error) {
	var ids []uint
	if err := r.db.Model(&models.PostProduct{}).
		Where("post_id = ?", postID).
		Order("sort ASC, id ASC").
		Pluck("product_id", &ids).Error; err != nil {
		return nil, err
	}
	return ids, nil
}

// SetRelatedProductIDs 替换文章关联的商品 ID 列表（按入参顺序作为 sort）
func (r *GormPostRepository) SetRelatedProductIDs(postID uint, productIDs []uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("post_id = ?", postID).Delete(&models.PostProduct{}).Error; err != nil {
			return err
		}
		if len(productIDs) == 0 {
			return nil
		}
		seen := make(map[uint]struct{}, len(productIDs))
		records := make([]models.PostProduct, 0, len(productIDs))
		for i, pid := range productIDs {
			if pid == 0 {
				continue
			}
			if _, ok := seen[pid]; ok {
				continue
			}
			seen[pid] = struct{}{}
			records = append(records, models.PostProduct{
				PostID:    postID,
				ProductID: pid,
				Sort:      i,
			})
		}
		if len(records) == 0 {
			return nil
		}
		return tx.Create(&records).Error
	})
}

// ListRelatedProducts 获取文章关联的商品（已按 sort 排序，过滤未删除）
func (r *GormPostRepository) ListRelatedProducts(postID uint) ([]models.Product, error) {
	var products []models.Product
	err := r.db.
		Joins("INNER JOIN post_products pp ON pp.product_id = products.id").
		Where("pp.post_id = ?", postID).
		Order("pp.sort ASC, pp.id ASC").
		Find(&products).Error
	if err != nil {
		return nil, err
	}
	return products, nil
}

// ListPostsForProduct 获取与某商品关联的文章列表
// postType 非空时按 type 过滤；onlyPublished 只取已发布；limit > 0 时限制条数
func (r *GormPostRepository) ListPostsForProduct(productID uint, postType string, onlyPublished bool, limit int) ([]models.Post, error) {
	var posts []models.Post
	query := r.db.
		Joins("INNER JOIN post_products pp ON pp.post_id = posts.id").
		Where("pp.product_id = ?", productID)
	if postType != "" {
		query = query.Where("posts.type = ?", postType)
	}
	if onlyPublished {
		query = query.Where("posts.is_published = ?", true)
	}
	if limit > 0 {
		query = query.Limit(limit)
	}
	if err := query.Order("pp.sort ASC, pp.id ASC").Find(&posts).Error; err != nil {
		return nil, err
	}
	return posts, nil
}
