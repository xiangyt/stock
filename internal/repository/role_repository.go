package repository

import (
	"fmt"
	"stock/internal/model"

	"gorm.io/gorm"
)

// RoleRepository 角色数据访问层
type RoleRepository struct {
	db *gorm.DB
}

// NewRoleRepository 创建角色仓库实例
func NewRoleRepository(db *gorm.DB) *RoleRepository {
	return &RoleRepository{db: db}
}

// Create 创建角色
func (r *RoleRepository) Create(role *model.Role) error {
	return r.db.Create(role).Error
}

// GetByID 根据ID获取角色
func (r *RoleRepository) GetByID(id uint) (*model.Role, error) {
	var role model.Role
	err := r.db.First(&role, id).Error
	return &role, err
}

// GetByCode 根据角色编码获取角色
func (r *RoleRepository) GetByCode(code string) (*model.Role, error) {
	var role model.Role
	err := r.db.Where("role_code = ?", code).First(&role).Error
	return &role, err
}

// GetList 获取角色列表
func (r *RoleRepository) GetList(page, pageSize int, keyword string) ([]model.Role, int64, error) {
	var roles []model.Role
	var total int64

	query := r.db.Model(&model.Role{})

	// 关键词搜索
	if keyword != "" {
		query = query.Where("role_name LIKE ? OR role_code LIKE ? OR description LIKE ?",
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("sort_order ASC, created_at DESC").Find(&roles).Error; err != nil {
		return nil, 0, err
	}

	return roles, total, nil
}

// GetAll 获取所有角色
func (r *RoleRepository) GetAll() ([]model.Role, error) {
	var roles []model.Role
	err := r.db.Where("status = ?", model.RoleStatusActive).Order("sort_order ASC").Find(&roles).Error
	return roles, err
}

// Update 更新角色
func (r *RoleRepository) Update(role *model.Role) error {
	fmt.Printf("[DEBUG] Repository: 准备更新角色到数据库: %+v\n", role)

	// 只更新角色字段，不更新关联关系
	result := r.db.Model(role).Select("role_name", "role_code", "description", "status", "sort_order", "updated_by", "updated_at").Updates(role)
	if result.Error != nil {
		fmt.Printf("[ERROR] Repository: 数据库更新失败: %v\n", result.Error)
		return result.Error
	}

	fmt.Printf("[DEBUG] Repository: 数据库更新成功，影响行数: %d\n", result.RowsAffected)
	return nil
}

// Delete 删除角色（软删除）
func (r *RoleRepository) Delete(role *model.Role) error {
	return r.db.Delete(role).Error
}

// ExistsByID 检查角色ID是否存在
func (r *RoleRepository) ExistsByID(id uint) (bool, error) {
	var count int64
	err := r.db.Model(&model.Role{}).Where("id = ?", id).Count(&count).Error
	return count > 0, err
}

// ExistsByCode 检查角色编码是否存在
func (r *RoleRepository) ExistsByCode(code string) (bool, error) {
	var count int64
	err := r.db.Model(&model.Role{}).Where("role_code = ?", code).Count(&count).Error
	return count > 0, err
}

// ExistsByName 检查角色名称是否存在
func (r *RoleRepository) ExistsByName(name string) (bool, error) {
	var count int64
	err := r.db.Model(&model.Role{}).Where("role_name = ?", name).Count(&count).Error
	return count > 0, err
}

// AssignPermissions 为角色分配权限
func (r *RoleRepository) AssignPermissions(roleID uint, permissionIDs []uint) error {
	// 使用事务确保数据一致性
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 先清除现有权限
		if err := tx.Where("role_id = ?", roleID).Delete(&model.RolePermission{}).Error; err != nil {
			return err
		}

		// 添加新权限 - 分批处理避免占位符过多的问题
		if len(permissionIDs) > 0 {
			const batchSize = 100 // 每批处理100个权限

			for i := 0; i < len(permissionIDs); i += batchSize {
				end := i + batchSize
				if end > len(permissionIDs) {
					end = len(permissionIDs)
				}

				var rolePermissions []model.RolePermission
				for _, permissionID := range permissionIDs[i:end] {
					rolePermissions = append(rolePermissions, model.RolePermission{
						RoleID:       roleID,
						PermissionID: permissionID,
						CreatedBy:    "system", // TODO: 从上下文获取当前用户
					})
				}

				if err := tx.Create(&rolePermissions).Error; err != nil {
					return err
				}
			}
		}

		return nil
	})
}

// GetRolePermissions 获取角色权限
func (r *RoleRepository) GetRolePermissions(roleID uint) ([]model.Permission, error) {
	var permissions []model.Permission
	err := r.db.Table("permissions").
		Joins("JOIN role_permissions ON permissions.id = role_permissions.permission_id").
		Where("role_permissions.role_id = ? AND permissions.status = ?", roleID, model.PermissionStatusActive).
		Find(&permissions).Error
	return permissions, err
}

// GetUserCount 获取角色下的用户数量
func (r *RoleRepository) GetUserCount(roleID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.User{}).Where("role_id = ?", roleID).Count(&count).Error
	return count, err
}

// GetByIDWithPermissions 根据ID获取角色（包含权限信息）
func (r *RoleRepository) GetByIDWithPermissions(id uint) (*model.Role, error) {
	var role model.Role
	err := r.db.Preload("Permissions").First(&role, id).Error
	return &role, err
}

// GetListWithPermissions 获取角色列表（包含权限信息）
func (r *RoleRepository) GetListWithPermissions(page, pageSize int, keyword string) ([]model.Role, int64, error) {
	var roles []model.Role
	var total int64

	query := r.db.Model(&model.Role{}).Preload("Permissions")

	// 关键词搜索
	if keyword != "" {
		query = query.Where("role_name LIKE ? OR role_code LIKE ? OR description LIKE ?",
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("sort_order ASC, created_at DESC").Find(&roles).Error; err != nil {
		return nil, 0, err
	}

	return roles, total, nil
}
