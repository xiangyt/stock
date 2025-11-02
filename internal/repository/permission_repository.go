package repository

import (
	"stock/internal/model"

	"gorm.io/gorm"
)

// PermissionRepository 权限数据访问层
type PermissionRepository struct {
	db *gorm.DB
}

// NewPermissionRepository 创建权限仓库实例
func NewPermissionRepository(db *gorm.DB) *PermissionRepository {
	return &PermissionRepository{db: db}
}

// Create 创建权限
func (r *PermissionRepository) Create(permission *model.Permission) error {
	return r.db.Create(permission).Error
}

// GetByID 根据ID获取权限
func (r *PermissionRepository) GetByID(id uint) (*model.Permission, error) {
	var permission model.Permission
	err := r.db.Preload("Parent").Preload("Children").First(&permission, id).Error
	return &permission, err
}

// GetByCode 根据权限编码获取权限
func (r *PermissionRepository) GetByCode(code string) (*model.Permission, error) {
	var permission model.Permission
	err := r.db.Preload("Parent").Preload("Children").Where("permission_code = ?", code).First(&permission).Error
	return &permission, err
}

// GetList 获取权限列表
func (r *PermissionRepository) GetList(page, pageSize int, keyword string) ([]model.Permission, int64, error) {
	var permissions []model.Permission
	var total int64

	query := r.db.Model(&model.Permission{}).Preload("Parent")

	// 关键词搜索
	if keyword != "" {
		query = query.Where("permission_name LIKE ? OR permission_code LIKE ? OR description LIKE ?",
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("sort_order ASC, created_at DESC").Find(&permissions).Error; err != nil {
		return nil, 0, err
	}

	return permissions, total, nil
}

// GetAll 获取所有权限
func (r *PermissionRepository) GetAll() ([]model.Permission, error) {
	var permissions []model.Permission
	err := r.db.Preload("Parent").Preload("Children").Where("status = ?", model.PermissionStatusActive).Order("sort_order ASC").Find(&permissions).Error
	return permissions, err
}

// GetTree 获取权限树
func (r *PermissionRepository) GetTree() ([]model.Permission, error) {
	var permissions []model.Permission
	err := r.db.Preload("Children", func(db *gorm.DB) *gorm.DB {
		return db.Where("status = ?", model.PermissionStatusActive).Order("sort_order ASC")
	}).Where("parent_id IS NULL AND status = ?", model.PermissionStatusActive).Order("sort_order ASC").Find(&permissions).Error
	return permissions, err
}

// Update 更新权限
func (r *PermissionRepository) Update(permission *model.Permission) error {
	// 只更新权限字段，不更新关联关系
	return r.db.Model(permission).Select("permission_name", "permission_code", "resource_type", "resource_path", "parent_id", "description", "status", "sort_order", "updated_at").Updates(permission).Error
}

// Delete 删除权限（软删除）
func (r *PermissionRepository) Delete(permission *model.Permission) error {
	return r.db.Delete(permission).Error
}

// ExistsByID 检查权限ID是否存在
func (r *PermissionRepository) ExistsByID(id uint) (bool, error) {
	var count int64
	err := r.db.Model(&model.Permission{}).Where("id = ?", id).Count(&count).Error
	return count > 0, err
}

// ExistsByCode 检查权限编码是否存在
func (r *PermissionRepository) ExistsByCode(code string) (bool, error) {
	var count int64
	err := r.db.Model(&model.Permission{}).Where("permission_code = ?", code).Count(&count).Error
	return count > 0, err
}

// GetByParentID 根据父权限ID获取子权限列表
func (r *PermissionRepository) GetByParentID(parentID uint) ([]model.Permission, error) {
	var permissions []model.Permission
	err := r.db.Where("parent_id = ? AND status = ?", parentID, model.PermissionStatusActive).Order("sort_order ASC").Find(&permissions).Error
	return permissions, err
}

// GetRootPermissions 获取根权限列表
func (r *PermissionRepository) GetRootPermissions() ([]model.Permission, error) {
	var permissions []model.Permission
	err := r.db.Where("parent_id IS NULL AND status = ?", model.PermissionStatusActive).Order("sort_order ASC").Find(&permissions).Error
	return permissions, err
}

// BatchCreate 批量创建权限
func (r *PermissionRepository) BatchCreate(permissions []model.Permission) error {
	return r.db.Create(&permissions).Error
}
