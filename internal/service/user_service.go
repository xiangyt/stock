package service

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"stock/internal/model"
	"stock/internal/repository"
)

// UserService 用户服务
type UserService struct {
	userRepo *repository.UserRepository
	roleRepo *repository.RoleRepository
}

// NewUserService 创建用户服务实例
func NewUserService(userRepo *repository.UserRepository, roleRepo *repository.RoleRepository) *UserService {
	return &UserService{
		userRepo: userRepo,
		roleRepo: roleRepo,
	}
}

// CreateUserRequest 创建用户请求
type CreateUserRequest struct {
	Username string  `json:"username" binding:"required,min=3,max=50"`
	Email    string  `json:"email" binding:"required,email"`
	Password string  `json:"password" binding:"required,min=6"`
	RealName *string `json:"real_name"`
	Phone    *string `json:"phone"`
	RoleID   *uint   `json:"role_id"`
}

// UpdateUserRequest 更新用户请求
type UpdateUserRequest struct {
	Email    *string `json:"email" binding:"omitempty,email"`
	RealName *string `json:"real_name"`
	Phone    *string `json:"phone"`
	RoleID   *uint   `json:"role_id"`
	Status   *int8   `json:"status"`
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username  string `json:"username" binding:"required"`
	Password  string `json:"password" binding:"required"`
	IP        string `json:"-"`
	UserAgent string `json:"-"`
}

// UserListResponse 用户列表响应
type UserListResponse struct {
	Users []model.User `json:"users"`
	Total int64        `json:"total"`
}

// CreateUser 创建用户
func (s *UserService) CreateUser(req *CreateUserRequest) (*model.User, error) {
	// 检查用户名是否已存在
	if exists, err := s.userRepo.ExistsByUsername(req.Username); err != nil {
		return nil, fmt.Errorf("检查用户名失败: %w", err)
	} else if exists {
		return nil, errors.New("用户名已存在")
	}

	// 检查邮箱是否已存在
	if exists, err := s.userRepo.ExistsByEmail(req.Email); err != nil {
		return nil, fmt.Errorf("检查邮箱失败: %w", err)
	} else if exists {
		return nil, errors.New("邮箱已存在")
	}

	// 验证角色是否存在
	if req.RoleID != nil {
		if exists, err := s.roleRepo.ExistsByID(*req.RoleID); err != nil {
			return nil, fmt.Errorf("检查角色失败: %w", err)
		} else if !exists {
			return nil, errors.New("指定的角色不存在")
		}
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("密码加密失败: %w", err)
	}

	// 创建用户
	user := &model.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		RealName:     req.RealName,
		Phone:        req.Phone,
		RoleID:       req.RoleID,
		Status:       model.UserStatusActive,
		CreatedBy:    "system", // TODO: 从上下文获取当前用户
		UpdatedBy:    "system",
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, fmt.Errorf("创建用户失败: %w", err)
	}

	return user, nil
}

// GetUserByID 根据ID获取用户
func (s *UserService) GetUserByID(id uint) (*model.User, error) {
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户不存在")
		}
		return nil, fmt.Errorf("获取用户失败: %w", err)
	}
	return user, nil
}

// GetUserByUsername 根据用户名获取用户
func (s *UserService) GetUserByUsername(username string) (*model.User, error) {
	user, err := s.userRepo.GetByUsername(username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户不存在")
		}
		return nil, fmt.Errorf("获取用户失败: %w", err)
	}
	return user, nil
}

// GetUserList 获取用户列表
func (s *UserService) GetUserList(page, pageSize int, keyword string) (*UserListResponse, error) {
	users, total, err := s.userRepo.GetListWithRole(page, pageSize, keyword)
	if err != nil {
		return nil, fmt.Errorf("获取用户列表失败: %w", err)
	}

	return &UserListResponse{
		Users: users,
		Total: total,
	}, nil
}

// UpdateUser 更新用户
func (s *UserService) UpdateUser(id uint, req *UpdateUserRequest) (*model.User, error) {
	// 获取用户
	user, err := s.GetUserByID(id)
	if err != nil {
		return nil, err
	}

	// 检查邮箱是否被其他用户使用
	if req.Email != nil && *req.Email != user.Email {
		if exists, err := s.userRepo.ExistsByEmail(*req.Email); err != nil {
			return nil, fmt.Errorf("检查邮箱失败: %w", err)
		} else if exists {
			return nil, errors.New("邮箱已被其他用户使用")
		}
		user.Email = *req.Email
	}

	// 验证角色是否存在
	if req.RoleID != nil && (user.RoleID == nil || *req.RoleID != *user.RoleID) {
		if exists, err := s.roleRepo.ExistsByID(*req.RoleID); err != nil {
			return nil, fmt.Errorf("检查角色失败: %w", err)
		} else if !exists {
			return nil, errors.New("指定的角色不存在")
		}
		user.RoleID = req.RoleID
	}

	// 更新其他字段
	if req.RealName != nil {
		user.RealName = req.RealName
	}
	if req.Phone != nil {
		user.Phone = req.Phone
	}
	if req.Status != nil {
		user.Status = *req.Status
	}

	user.UpdatedBy = "system" // TODO: 从上下文获取当前用户

	if err := s.userRepo.Update(user); err != nil {
		return nil, fmt.Errorf("更新用户失败: %w", err)
	}

	return user, nil
}

// DeleteUser 删除用户（软删除）
func (s *UserService) DeleteUser(id uint) error {
	user, err := s.GetUserByID(id)
	if err != nil {
		return err
	}

	if err := s.userRepo.Delete(user); err != nil {
		return fmt.Errorf("删除用户失败: %w", err)
	}

	return nil
}

// CreateRoleRequest 创建角色请求
type CreateRoleRequest struct {
	RoleName    string  `json:"role_name" binding:"required,min=2,max=50"`
	RoleCode    string  `json:"role_code" binding:"required,min=2,max=50"`
	Description *string `json:"description"`
	SortOrder   *int    `json:"sort_order"`
	Status      *int8   `json:"status"`
}

// UpdateRoleRequest 更新角色请求
type UpdateRoleRequest struct {
	RoleName    *string `json:"role_name" binding:"omitempty,min=2,max=50"`
	RoleCode    *string `json:"role_code" binding:"omitempty,min=2,max=50"`
	Description *string `json:"description"`
	SortOrder   *int    `json:"sort_order"`
	Status      *int8   `json:"status"`
}

// RoleListResponse 角色列表响应
type RoleListResponse struct {
	Roles []model.Role `json:"roles"`
	Total int64        `json:"total"`
}

// CreateRole 创建角色
func (s *UserService) CreateRole(req *CreateRoleRequest) (*model.Role, error) {
	// 检查角色名称是否已存在
	if exists, err := s.roleRepo.ExistsByName(req.RoleName); err != nil {
		return nil, fmt.Errorf("检查角色名称失败: %w", err)
	} else if exists {
		return nil, errors.New("角色名称已存在")
	}

	// 检查角色编码是否已存在
	if exists, err := s.roleRepo.ExistsByCode(req.RoleCode); err != nil {
		return nil, fmt.Errorf("检查角色编码失败: %w", err)
	} else if exists {
		return nil, errors.New("角色编码已存在")
	}

	// 创建角色
	role := &model.Role{
		RoleName:    req.RoleName,
		RoleCode:    req.RoleCode,
		Description: req.Description,
		Status:      1,        // 默认启用
		CreatedBy:   "system", // TODO: 从上下文获取当前用户
		UpdatedBy:   "system",
	}

	if req.Status != nil {
		role.Status = *req.Status
	}

	if req.SortOrder != nil {
		role.SortOrder = *req.SortOrder
	} else {
		role.SortOrder = 0 // 默认值0
	}

	if err := s.roleRepo.Create(role); err != nil {
		return nil, fmt.Errorf("创建角色失败: %w", err)
	}

	return role, nil
}

// GetRoles 获取角色列表
func (s *UserService) GetRoles(page, pageSize int, keyword string) (*RoleListResponse, error) {
	roles, total, err := s.roleRepo.GetList(page, pageSize, keyword)
	if err != nil {
		return nil, fmt.Errorf("获取角色列表失败: %w", err)
	}

	return &RoleListResponse{
		Roles: roles,
		Total: total,
	}, nil
}

// GetRoleByID 根据ID获取角色
func (s *UserService) GetRoleByID(id uint) (*model.Role, error) {
	role, err := s.roleRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("角色不存在")
		}
		return nil, fmt.Errorf("获取角色失败: %w", err)
	}
	return role, nil
}

// UpdateRole 更新角色
func (s *UserService) UpdateRole(id uint, req *UpdateRoleRequest) (*model.Role, error) {
	fmt.Printf("[DEBUG] 开始更新角色，ID: %d, 请求数据: %+v\n", id, req)

	role, err := s.GetRoleByID(id)
	if err != nil {
		fmt.Printf("[ERROR] 获取角色失败: %v\n", err)
		return nil, err
	}

	fmt.Printf("[DEBUG] 原始角色数据: %+v\n", role)

	// 记录是否有字段被修改
	hasChanges := false

	// 检查角色名称是否已存在（排除当前角色）
	if req.RoleName != nil && *req.RoleName != role.RoleName {
		if exists, err := s.roleRepo.ExistsByName(*req.RoleName); err != nil {
			return nil, fmt.Errorf("检查角色名称失败: %w", err)
		} else if exists {
			return nil, errors.New("角色名称已存在")
		}
		fmt.Printf("[DEBUG] 更新角色名称: %s -> %s\n", role.RoleName, *req.RoleName)
		role.RoleName = *req.RoleName
		hasChanges = true
	}

	// 检查角色编码是否已存在（排除当前角色）
	if req.RoleCode != nil && *req.RoleCode != role.RoleCode {
		if exists, err := s.roleRepo.ExistsByCode(*req.RoleCode); err != nil {
			return nil, fmt.Errorf("检查角色编码失败: %w", err)
		} else if exists {
			return nil, errors.New("角色编码已存在")
		}
		fmt.Printf("[DEBUG] 更新角色编码: %s -> %s\n", role.RoleCode, *req.RoleCode)
		role.RoleCode = *req.RoleCode
		hasChanges = true
	}

	// 更新其他字段
	if req.Description != nil {
		if role.Description == nil || *role.Description != *req.Description {
			fmt.Printf("[DEBUG] 更新描述: %v -> %s\n", role.Description, *req.Description)
			role.Description = req.Description
			hasChanges = true
		}
	}
	if req.SortOrder != nil {
		if role.SortOrder != *req.SortOrder {
			fmt.Printf("[DEBUG] 更新排序: %d -> %d\n", role.SortOrder, *req.SortOrder)
			role.SortOrder = *req.SortOrder
			hasChanges = true
		}
	}
	if req.Status != nil {
		if role.Status != *req.Status {
			fmt.Printf("[DEBUG] 更新状态: %d -> %d\n", role.Status, *req.Status)
			role.Status = *req.Status
			hasChanges = true
		}
	}

	role.UpdatedBy = "system" // TODO: 从上下文获取当前用户

	fmt.Printf("[DEBUG] 是否有字段变更: %v\n", hasChanges)
	fmt.Printf("[DEBUG] 准备保存的角色数据: %+v\n", role)

	if err := s.roleRepo.Update(role); err != nil {
		fmt.Printf("[ERROR] 数据库更新失败: %v\n", err)
		return nil, fmt.Errorf("更新角色失败: %w", err)
	}

	fmt.Printf("[DEBUG] 角色更新成功，返回数据: %+v\n", role)
	return role, nil
}

// DeleteRole 删除角色
func (s *UserService) DeleteRole(id uint) error {
	// 检查角色是否存在
	role, err := s.GetRoleByID(id)
	if err != nil {
		return err
	}

	// 检查是否有用户使用此角色
	if hasUsers, err := s.userRepo.HasUsersByRoleID(id); err != nil {
		return fmt.Errorf("检查角色使用情况失败: %w", err)
	} else if hasUsers {
		return errors.New("该角色下还有用户，无法删除")
	}

	if err := s.roleRepo.Delete(role); err != nil {
		return fmt.Errorf("删除角色失败: %w", err)
	}

	return nil
}

// ChangePassword 修改密码
func (s *UserService) ChangePassword(userID uint, req *ChangePasswordRequest) error {
	user, err := s.GetUserByID(userID)
	if err != nil {
		return err
	}

	// 验证旧密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.OldPassword)); err != nil {
		return errors.New("原密码不正确")
	}

	// 加密新密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("密码加密失败: %w", err)
	}

	user.PasswordHash = string(hashedPassword)
	user.UpdatedBy = "system" // TODO: 从上下文获取当前用户

	if err := s.userRepo.Update(user); err != nil {
		return fmt.Errorf("更新密码失败: %w", err)
	}

	return nil
}

// Login 用户登录
func (s *UserService) Login(req *LoginRequest) (*model.User, error) {
	// 获取用户（不预加载权限数据，避免数据过大）
	user, err := s.userRepo.GetByUsernameForLogin(req.Username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 记录登录失败日志
			s.recordLoginLog(nil, req.Username, req.IP, req.UserAgent, false, "用户不存在")
			return nil, errors.New("用户名或密码错误")
		}
		return nil, fmt.Errorf("获取用户失败: %w", err)
	}

	// 检查用户状态
	if !user.IsActive() {
		s.recordLoginLog(&user.ID, req.Username, req.IP, req.UserAgent, false, "用户已被禁用或锁定")
		return nil, errors.New("用户已被禁用或锁定")
	}

	// 检查是否被锁定
	if user.IsLocked() {
		s.recordLoginLog(&user.ID, req.Username, req.IP, req.UserAgent, false, "用户已被锁定")
		return nil, errors.New("用户已被锁定，请稍后再试")
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		// 增加失败次数
		user.FailedLoginCount++
		if user.FailedLoginCount >= 5 {
			// 锁定30分钟
			lockUntil := time.Now().Add(30 * time.Minute)
			user.LockedUntil = &lockUntil
		}
		s.userRepo.Update(user)

		s.recordLoginLog(&user.ID, req.Username, req.IP, req.UserAgent, false, "密码错误")
		return nil, errors.New("用户名或密码错误")
	}

	// 登录成功，重置失败次数
	user.FailedLoginCount = 0
	user.LockedUntil = nil
	user.LoginCount++
	now := time.Now()
	user.LastLoginAt = &now
	user.LastLoginIP = &req.IP

	if err := s.userRepo.Update(user); err != nil {
		return nil, fmt.Errorf("更新用户登录信息失败: %w", err)
	}

	// 记录登录成功日志
	s.recordLoginLog(&user.ID, req.Username, req.IP, req.UserAgent, true, "")

	return user, nil
}

// recordLoginLog 记录登录日志
func (s *UserService) recordLoginLog(userID *uint, username, ip, userAgent string, success bool, failureReason string) {
	log := &model.UserLoginLog{
		UserID:      userID,
		Username:    username,
		LoginIP:     ip,
		UserAgent:   &userAgent,
		LoginStatus: success,
		LoginTime:   time.Now(),
	}

	if !success && failureReason != "" {
		log.FailureReason = &failureReason
	}

	// 生成会话ID
	if success {
		sessionID := s.generateSessionID()
		log.SessionID = &sessionID
	}

	// 这里应该调用登录日志的repository，暂时忽略错误
	// TODO: 实现登录日志的repository
}

// generateSessionID 生成会话ID
func (s *UserService) generateSessionID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// EnableUser 启用用户
func (s *UserService) EnableUser(id uint) error {
	user, err := s.GetUserByID(id)
	if err != nil {
		return err
	}

	user.Status = model.UserStatusActive
	user.UpdatedBy = "system" // TODO: 从上下文获取当前用户

	if err := s.userRepo.Update(user); err != nil {
		return fmt.Errorf("启用用户失败: %w", err)
	}

	return nil
}

// DisableUser 禁用用户
func (s *UserService) DisableUser(id uint) error {
	user, err := s.GetUserByID(id)
	if err != nil {
		return err
	}

	user.Status = model.UserStatusDisabled
	user.UpdatedBy = "system" // TODO: 从上下文获取当前用户

	if err := s.userRepo.Update(user); err != nil {
		return fmt.Errorf("禁用用户失败: %w", err)
	}

	return nil
}

// LockUser 锁定用户
func (s *UserService) LockUser(id uint, duration time.Duration) error {
	user, err := s.GetUserByID(id)
	if err != nil {
		return err
	}

	lockUntil := time.Now().Add(duration)
	user.Status = model.UserStatusLocked
	user.LockedUntil = &lockUntil
	user.UpdatedBy = "system" // TODO: 从上下文获取当前用户

	if err := s.userRepo.Update(user); err != nil {
		return fmt.Errorf("锁定用户失败: %w", err)
	}

	return nil
}

// UnlockUser 解锁用户
func (s *UserService) UnlockUser(id uint) error {
	user, err := s.GetUserByID(id)
	if err != nil {
		return err
	}

	user.Status = model.UserStatusActive
	user.LockedUntil = nil
	user.FailedLoginCount = 0
	user.UpdatedBy = "system" // TODO: 从上下文获取当前用户

	if err := s.userRepo.Update(user); err != nil {
		return fmt.Errorf("解锁用户失败: %w", err)
	}

	return nil
}
