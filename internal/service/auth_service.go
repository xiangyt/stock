package service

import (
	"fmt"
	"stock/internal/middleware"
	"stock/internal/model"
	"stock/internal/repository"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// AuthService 认证服务
type AuthService struct {
	db             *gorm.DB
	userRepo       *repository.UserRepository
	roleRepo       *repository.RoleRepository
	authMiddleware *middleware.AuthMiddleware
}

// NewAuthService 创建认证服务
func NewAuthService(db *gorm.DB, jwtSecret string) *AuthService {
	return &AuthService{
		db:             db,
		userRepo:       repository.NewUserRepository(db),
		roleRepo:       repository.NewRoleRepository(db),
		authMiddleware: middleware.NewAuthMiddleware(db, jwtSecret),
	}
}

// AuthLoginRequest 认证登录请求
type AuthLoginRequest struct {
	Username  string `json:"username" binding:"required"`
	Password  string `json:"password" binding:"required"`
	IP        string `json:"ip"`
	UserAgent string `json:"user_agent"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token     string      `json:"token"`
	User      *model.User `json:"user"`
	ExpiresAt time.Time   `json:"expires_at"`
}

// Logout 用户登出
func (s *AuthService) Logout(tokenString string) error {
	return s.authMiddleware.RevokeToken(tokenString, "用户主动登出")
}

// RefreshToken 刷新Token
func (s *AuthService) RefreshToken(userID uint) (*LoginResponse, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("用户不存在")
	}

	if !user.IsActive() {
		return nil, fmt.Errorf("用户已被禁用")
	}

	if user.IsLocked() {
		return nil, fmt.Errorf("用户已被锁定")
	}

	// 生成新Token
	token, err := s.authMiddleware.GenerateToken(user)
	if err != nil {
		return nil, fmt.Errorf("生成Token失败: %v", err)
	}

	// 清除敏感信息
	user.PasswordHash = ""

	return &LoginResponse{
		Token:     token,
		User:      user,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}, nil
}

// GetUserMenuPermissions 获取用户菜单权限
func (s *AuthService) GetUserMenuPermissions(userID uint) (map[string]interface{}, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("用户不存在")
	}

	// 获取用户角色的权限
	var permissions []model.Permission
	err = s.db.Table("permissions").
		Joins("JOIN role_permissions ON permissions.id = role_permissions.permission_id").
		Where("role_permissions.role_id = ? AND permissions.status = 1", user.RoleID).
		Order("permissions.sort_order ASC").
		Find(&permissions).Error

	if err != nil {
		return nil, fmt.Errorf("获取用户权限失败: %v", err)
	}

	// 获取用户角色信息
	var roleName, roleCode string
	if user.RoleID != nil {
		role, err := s.roleRepo.GetByID(*user.RoleID)
		if err == nil {
			roleName = role.RoleName
			roleCode = role.RoleCode
		}
	}

	return map[string]interface{}{
		"permissions":    permissions,
		"user_role":      roleName,
		"user_role_code": roleCode,
		"message":        "获取用户菜单权限成功",
	}, nil
}

// GetUserLoginLogs 获取用户登录日志
func (s *AuthService) GetUserLoginLogs(userID uint, page, pageSize int) ([]model.UserLoginLog, int64, error) {
	var logs []model.UserLoginLog
	var total int64

	offset := (page - 1) * pageSize

	// 获取总数
	if err := s.db.Model(&model.UserLoginLog{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("获取登录日志总数失败: %v", err)
	}

	// 获取分页数据
	if err := s.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&logs).Error; err != nil {
		return nil, 0, fmt.Errorf("获取登录日志失败: %v", err)
	}

	return logs, total, nil
}

// Login 用户登录
func (s *AuthService) Login(req *AuthLoginRequest) (*LoginResponse, error) {
	// 记录登录尝试
	loginLog := &model.UserLoginLog{
		Username:    req.Username,
		LoginIP:     req.IP,
		UserAgent:   &req.UserAgent,
		LoginStatus: false,
		LoginTime:   time.Now(),
	}

	// 查找用户（使用登录专用方法，不预加载权限数据）
	user, err := s.userRepo.GetByUsernameForLogin(req.Username)
	if err != nil {
		loginLog.FailureReason = stringPtr("用户不存在")
		s.db.Create(loginLog)
		return nil, fmt.Errorf("用户名或密码错误")
	}

	loginLog.UserID = &user.ID

	// 检查用户状态
	if !user.IsActive() {
		loginLog.FailureReason = stringPtr("用户已被禁用")
		s.db.Create(loginLog)
		return nil, fmt.Errorf("用户已被禁用")
	}

	if user.IsLocked() {
		loginLog.FailureReason = stringPtr("用户已被锁定")
		s.db.Create(loginLog)
		return nil, fmt.Errorf("用户已被锁定")
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		loginLog.FailureReason = stringPtr("密码错误")
		s.db.Create(loginLog)

		// 增加失败登录次数
		s.userRepo.IncrementFailedLoginCount(user.ID)

		// 检查是否需要锁定用户（连续失败5次）
		if user.FailedLoginCount >= 4 {
			lockUntil := time.Now().Add(30 * time.Minute) // 锁定30分钟
			s.db.Model(user).Update("locked_until", lockUntil)
		}

		return nil, fmt.Errorf("用户名或密码错误")
	}

	// 登录成功，重置失败次数
	s.userRepo.ResetFailedLoginCount(user.ID)

	// 更新最后登录信息
	s.userRepo.UpdateLastLogin(user.ID, req.IP)

	// 生成JWT令牌
	token, err := s.authMiddleware.GenerateToken(user)
	if err != nil {
		loginLog.FailureReason = stringPtr("生成令牌失败")
		s.db.Create(loginLog)
		return nil, fmt.Errorf("生成令牌失败: %v", err)
	}

	// 记录成功登录
	loginLog.LoginStatus = true
	loginLog.FailureReason = nil
	s.db.Create(loginLog)

	// 清除密码哈希
	user.PasswordHash = ""

	return &LoginResponse{
		Token:     token,
		User:      user,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}, nil
}

// ChangePassword 修改密码
func (s *AuthService) ChangePassword(userID uint, oldPassword, newPassword string) error {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return fmt.Errorf("用户不存在")
	}

	// 验证旧密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword)); err != nil {
		return fmt.Errorf("原密码错误")
	}

	// 生成新密码哈希
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("密码加密失败")
	}

	// 更新密码
	user.PasswordHash = string(hashedPassword)
	if err := s.userRepo.Update(user); err != nil {
		return fmt.Errorf("密码更新失败")
	}

	return nil
}

// GetProfile 获取用户资料
func (s *AuthService) GetProfile(userID uint) (*model.User, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("用户不存在")
	}

	// 清除敏感信息
	user.PasswordHash = ""

	return user, nil
}

// UpdateProfile 更新用户资料
func (s *AuthService) UpdateProfile(userID uint, req *UpdateProfileRequest) (*model.User, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("用户不存在")
	}

	// 更新允许修改的字段
	if req.RealName != nil {
		user.RealName = req.RealName
	}
	if req.Phone != nil {
		user.Phone = req.Phone
	}
	if req.Avatar != nil {
		user.Avatar = req.Avatar
	}

	if err := s.userRepo.Update(user); err != nil {
		return nil, fmt.Errorf("更新用户资料失败")
	}

	// 清除敏感信息
	user.PasswordHash = ""

	return user, nil
}

// UpdateProfileRequest 更新资料请求
type UpdateProfileRequest struct {
	RealName *string `json:"real_name"`
	Phone    *string `json:"phone"`
	Avatar   *string `json:"avatar"`
}

func stringPtr(s string) *string {
	return &s
}
