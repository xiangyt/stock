package repository

import (
	"stock/internal/model"

	"gorm.io/gorm"
)

// UserRepository 用户数据访问层
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository 创建用户仓库实例
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create 创建用户
func (r *UserRepository) Create(user *model.User) error {
	return r.db.Create(user).Error
}

// GetByID 根据ID获取用户
func (r *UserRepository) GetByID(id uint) (*model.User, error) {
	var user model.User
	err := r.db.First(&user, id).Error
	return &user, err
}

// GetByUsername 根据用户名获取用户
func (r *UserRepository) GetByUsername(username string) (*model.User, error) {
	var user model.User
	err := r.db.Where("username = ?", username).First(&user).Error
	return &user, err
}

// GetByUsernameForLogin 根据用户名获取用户（用于登录）
func (r *UserRepository) GetByUsernameForLogin(username string) (*model.User, error) {
	var user model.User
	err := r.db.Where("username = ?", username).First(&user).Error
	return &user, err
}

// GetByEmail 根据邮箱获取用户
func (r *UserRepository) GetByEmail(email string) (*model.User, error) {
	var user model.User
	err := r.db.Where("email = ?", email).First(&user).Error
	return &user, err
}

// GetList 获取用户列表
func (r *UserRepository) GetList(page, pageSize int, keyword string) ([]model.User, int64, error) {
	var users []model.User
	var total int64

	query := r.db.Model(&model.User{})

	// 关键词搜索
	if keyword != "" {
		query = query.Where("username LIKE ? OR email LIKE ? OR real_name LIKE ?",
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// Update 更新用户
func (r *UserRepository) Update(user *model.User) error {
	// 只更新用户字段，不更新关联关系
	return r.db.Model(user).
		Select("username", "email", "phone", "real_name", "avatar", "status", "role_id", "updated_by", "updated_at").
		Updates(user).Error
}

// Delete 删除用户（软删除）
func (r *UserRepository) Delete(user *model.User) error {
	return r.db.Delete(user).Error
}

// ExistsByUsername 检查用户名是否存在
func (r *UserRepository) ExistsByUsername(username string) (bool, error) {
	var count int64
	err := r.db.Model(&model.User{}).Where("username = ?", username).Count(&count).Error
	return count > 0, err
}

// ExistsByEmail 检查邮箱是否存在
func (r *UserRepository) ExistsByEmail(email string) (bool, error) {
	var count int64
	err := r.db.Model(&model.User{}).Where("email = ?", email).Count(&count).Error
	return count > 0, err
}

// HasUsersByRoleID 检查是否有用户使用指定角色
func (r *UserRepository) HasUsersByRoleID(roleID uint) (bool, error) {
	var count int64
	err := r.db.Model(&model.User{}).Where("role_id = ?", roleID).Count(&count).Error
	return count > 0, err
}

// GetActiveUsers 获取活跃用户列表
func (r *UserRepository) GetActiveUsers() ([]model.User, error) {
	var users []model.User
	err := r.db.Where("status = ?", model.UserStatusActive).Find(&users).Error
	return users, err
}

// GetUsersByRoleID 根据角色ID获取用户列表
func (r *UserRepository) GetUsersByRoleID(roleID uint) ([]model.User, error) {
	var users []model.User
	err := r.db.Where("role_id = ?", roleID).Find(&users).Error
	return users, err
}

// UpdateLastLogin 更新最后登录信息
func (r *UserRepository) UpdateLastLogin(userID uint, ip string) error {
	return r.db.Model(&model.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"last_login_at": gorm.Expr("NOW()"),
		"last_login_ip": ip,
		"login_count":   gorm.Expr("login_count + 1"),
	}).Error
}

// IncrementFailedLoginCount 增加失败登录次数
func (r *UserRepository) IncrementFailedLoginCount(userID uint) error {
	return r.db.Model(&model.User{}).Where("id = ?", userID).Update("failed_login_count", gorm.Expr("failed_login_count + 1")).Error
}

// ResetFailedLoginCount 重置失败登录次数
func (r *UserRepository) ResetFailedLoginCount(userID uint) error {
	return r.db.Model(&model.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"failed_login_count": 0,
		"locked_until":       nil,
	}).Error
}

// GetByIDWithRole 根据ID获取用户（包含角色信息）
func (r *UserRepository) GetByIDWithRole(id uint) (*model.User, error) {
	var user model.User
	err := r.db.Preload("Role").First(&user, id).Error
	return &user, err
}

// GetWithRole 根据ID获取用户（不使用预加载）
func (r *UserRepository) GetWithRole(id uint) (*model.User, error) {
	var user model.User
	err := r.db.First(&user, id).Error
	return &user, err
}

// GetListWithRole 获取用户列表（包含角色信息）
func (r *UserRepository) GetListWithRole(page, pageSize int, keyword string) ([]model.User, int64, error) {
	var users []model.User
	var total int64

	query := r.db.Model(&model.User{}).Preload("Role")

	// 关键词搜索
	if keyword != "" {
		query = query.Where("username LIKE ? OR email LIKE ? OR real_name LIKE ?",
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}
