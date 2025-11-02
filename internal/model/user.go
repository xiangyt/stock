package model

import (
	"time"

	"gorm.io/gorm"
)

// User 用户模型
type User struct {
	ID               uint           `json:"id" gorm:"primaryKey;autoIncrement;comment:用户ID，主键"`
	Username         string         `json:"username" gorm:"type:varchar(50);uniqueIndex;not null;comment:用户名，登录账号"`
	Email            string         `json:"email" gorm:"type:varchar(100);uniqueIndex;not null;comment:邮箱地址"`
	Phone            *string        `json:"phone" gorm:"type:varchar(20);index;comment:手机号码"`
	PasswordHash     string         `json:"-" gorm:"type:varchar(255);not null;comment:密码哈希值"`
	RealName         *string        `json:"real_name" gorm:"type:varchar(50);comment:真实姓名"`
	Avatar           *string        `json:"avatar" gorm:"type:varchar(255);comment:头像URL"`
	Status           int8           `json:"status" gorm:"type:tinyint(1);default:1;index;comment:用户状态：1=正常，0=禁用，2=锁定"`
	RoleID           *uint          `json:"role_id" gorm:"index;comment:角色ID，外键关联roles表"`
	LastLoginAt      *time.Time     `json:"last_login_at" gorm:"type:datetime(3);index;comment:最后登录时间"`
	LastLoginIP      *string        `json:"last_login_ip" gorm:"type:varchar(45);comment:最后登录IP"`
	LoginCount       int            `json:"login_count" gorm:"default:0;comment:登录次数"`
	FailedLoginCount int            `json:"failed_login_count" gorm:"default:0;comment:连续失败登录次数"`
	LockedUntil      *time.Time     `json:"locked_until" gorm:"type:datetime(3);comment:锁定到期时间"`
	EmailVerified    bool           `json:"email_verified" gorm:"type:tinyint(1);default:0;comment:邮箱是否验证：1=已验证，0=未验证"`
	PhoneVerified    bool           `json:"phone_verified" gorm:"type:tinyint(1);default:0;comment:手机是否验证：1=已验证，0=未验证"`
	CreatedAt        time.Time      `json:"created_at" gorm:"type:datetime(3);default:CURRENT_TIMESTAMP(3);comment:创建时间"`
	UpdatedAt        time.Time      `json:"updated_at" gorm:"type:datetime(3);default:CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3);comment:更新时间"`
	CreatedBy        string         `json:"created_by" gorm:"type:varchar(50);default:'system';comment:创建人"`
	UpdatedBy        string         `json:"updated_by" gorm:"type:varchar(50);default:'system';comment:更新人"`
	DeletedAt        gorm.DeletedAt `json:"-" gorm:"index;comment:软删除时间"`

	// 关联关系
	Role          *Role              `json:"role,omitempty" gorm:"foreignKey:RoleID;references:ID"`
	LoginLogs     []UserLoginLog     `json:"login_logs,omitempty" gorm:"foreignKey:UserID"`
	OperationLogs []UserOperationLog `json:"operation_logs,omitempty" gorm:"foreignKey:UserID"`
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}

// Role 角色模型
type Role struct {
	ID          uint           `json:"id" gorm:"primaryKey;autoIncrement;comment:角色ID，主键"`
	RoleName    string         `json:"role_name" gorm:"type:varchar(50);uniqueIndex;not null;comment:角色名称"`
	RoleCode    string         `json:"role_code" gorm:"type:varchar(50);uniqueIndex;not null;comment:角色编码"`
	Description *string        `json:"description" gorm:"type:varchar(255);comment:角色描述"`
	IsSystem    bool           `json:"is_system" gorm:"type:tinyint(1);default:0;comment:是否系统角色"`
	Status      int8           `json:"status" gorm:"type:tinyint(1);default:1;index;comment:角色状态：1=启用，0=禁用"`
	SortOrder   int            `json:"sort_order" gorm:"default:0;comment:排序顺序"`
	CreatedAt   time.Time      `json:"created_at" gorm:"type:datetime(3);default:CURRENT_TIMESTAMP(3);comment:创建时间"`
	UpdatedAt   time.Time      `json:"updated_at" gorm:"type:datetime(3);default:CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3);comment:更新时间"`
	CreatedBy   string         `json:"created_by" gorm:"type:varchar(50);default:'system';comment:创建人"`
	UpdatedBy   string         `json:"updated_by" gorm:"type:varchar(50);default:'system';comment:更新人"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index;comment:软删除时间"`

	// 关联关系
	Users       []User       `json:"users,omitempty" gorm:"foreignKey:RoleID"`
	Permissions []Permission `json:"permissions,omitempty" gorm:"many2many:role_permissions;"`
}

// TableName 指定表名
func (Role) TableName() string {
	return "roles"
}

// Permission 权限模型
type Permission struct {
	ID             uint           `json:"id" gorm:"primaryKey;autoIncrement;comment:权限ID，主键"`
	PermissionName string         `json:"permission_name" gorm:"type:varchar(100);not null;comment:权限名称"`
	PermissionCode string         `json:"permission_code" gorm:"type:varchar(100);uniqueIndex;not null;comment:权限编码"`
	ResourceType   *string        `json:"resource_type" gorm:"type:varchar(50);index;comment:资源类型"`
	ResourcePath   *string        `json:"resource_path" gorm:"type:varchar(255);comment:资源路径"`
	ParentID       *uint          `json:"parent_id" gorm:"index;comment:父权限ID"`
	Description    *string        `json:"description" gorm:"type:varchar(255);comment:权限描述"`
	Status         int8           `json:"status" gorm:"type:tinyint(1);default:1;index;comment:权限状态：1=启用，0=禁用"`
	SortOrder      int            `json:"sort_order" gorm:"default:0;comment:排序顺序"`
	CreatedAt      time.Time      `json:"created_at" gorm:"type:datetime(3);default:CURRENT_TIMESTAMP(3);comment:创建时间"`
	UpdatedAt      time.Time      `json:"updated_at" gorm:"type:datetime(3);default:CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3);comment:更新时间"`
	DeletedAt      gorm.DeletedAt `json:"-" gorm:"index;comment:软删除时间"`

	// 关联关系
	Parent   *Permission  `json:"parent,omitempty" gorm:"foreignKey:ParentID;references:ID"`
	Children []Permission `json:"children,omitempty" gorm:"foreignKey:ParentID;references:ID"`
	Roles    []Role       `json:"roles,omitempty" gorm:"many2many:role_permissions;"`
}

// TableName 指定表名
func (Permission) TableName() string {
	return "permissions"
}

// RolePermission 角色权限关联模型
type RolePermission struct {
	ID           uint      `json:"id" gorm:"primaryKey;autoIncrement;comment:关联ID，主键"`
	RoleID       uint      `json:"role_id" gorm:"not null;index;comment:角色ID"`
	PermissionID uint      `json:"permission_id" gorm:"not null;index;comment:权限ID"`
	CreatedAt    time.Time `json:"created_at" gorm:"type:datetime(3);default:CURRENT_TIMESTAMP(3);comment:创建时间"`
	CreatedBy    string    `json:"created_by" gorm:"type:varchar(50);default:'system';comment:创建人"`

	// 关联关系
	Role       Role       `json:"role,omitempty" gorm:"foreignKey:RoleID;references:ID"`
	Permission Permission `json:"permission,omitempty" gorm:"foreignKey:PermissionID;references:ID"`
}

// TableName 指定表名
func (RolePermission) TableName() string {
	return "role_permissions"
}

// UserLoginLog 用户登录日志模型
type UserLoginLog struct {
	ID            uint      `json:"id" gorm:"primaryKey;autoIncrement;comment:日志ID，主键"`
	UserID        *uint     `json:"user_id" gorm:"index;comment:用户ID"`
	Username      string    `json:"username" gorm:"type:varchar(50);not null;index;comment:登录用户名"`
	LoginIP       string    `json:"login_ip" gorm:"type:varchar(45);not null;index;comment:登录IP地址"`
	UserAgent     *string   `json:"user_agent" gorm:"type:varchar(500);comment:用户代理信息"`
	LoginStatus   bool      `json:"login_status" gorm:"type:tinyint(1);not null;index;comment:登录状态：1=成功，0=失败"`
	FailureReason *string   `json:"failure_reason" gorm:"type:varchar(255);comment:失败原因"`
	LoginTime     time.Time `json:"login_time" gorm:"type:datetime(3);default:CURRENT_TIMESTAMP(3);index;comment:登录时间"`
	SessionID     *string   `json:"session_id" gorm:"type:varchar(255);comment:会话ID"`
	Location      *string   `json:"location" gorm:"type:varchar(100);comment:登录地点"`
	DeviceType    *string   `json:"device_type" gorm:"type:varchar(50);comment:设备类型"`

	// 关联关系
	User *User `json:"user,omitempty" gorm:"foreignKey:UserID;references:ID"`
}

// TableName 指定表名
func (UserLoginLog) TableName() string {
	return "user_login_logs"
}

// UserOperationLog 用户操作日志模型
type UserOperationLog struct {
	ID             uint      `json:"id" gorm:"primaryKey;autoIncrement;comment:日志ID，主键"`
	UserID         uint      `json:"user_id" gorm:"not null;index;comment:用户ID"`
	Username       string    `json:"username" gorm:"type:varchar(50);not null;comment:用户名"`
	OperationType  string    `json:"operation_type" gorm:"type:varchar(50);not null;index;comment:操作类型"`
	ResourceType   string    `json:"resource_type" gorm:"type:varchar(50);not null;index;comment:资源类型"`
	ResourceID     *string   `json:"resource_id" gorm:"type:varchar(100);comment:资源ID"`
	OperationDesc  string    `json:"operation_desc" gorm:"type:varchar(500);not null;comment:操作描述"`
	RequestMethod  *string   `json:"request_method" gorm:"type:varchar(10);comment:请求方法"`
	RequestURL     *string   `json:"request_url" gorm:"type:varchar(500);comment:请求URL"`
	RequestParams  *string   `json:"request_params" gorm:"type:text;comment:请求参数"`
	ResponseStatus *int      `json:"response_status" gorm:"comment:响应状态码"`
	IPAddress      *string   `json:"ip_address" gorm:"type:varchar(45);comment:IP地址"`
	UserAgent      *string   `json:"user_agent" gorm:"type:varchar(500);comment:用户代理"`
	ExecutionTime  *int      `json:"execution_time" gorm:"comment:执行时间（毫秒）"`
	CreatedAt      time.Time `json:"created_at" gorm:"type:datetime(3);default:CURRENT_TIMESTAMP(3);index;comment:操作时间"`

	// 关联关系
	User User `json:"user,omitempty" gorm:"foreignKey:UserID;references:ID"`
}

// TableName 指定表名
func (UserOperationLog) TableName() string {
	return "user_operation_logs"
}

// JWTBlacklist JWT令牌黑名单模型
type JWTBlacklist struct {
	ID           uint      `json:"id" gorm:"primaryKey;autoIncrement;comment:记录ID，主键"`
	TokenID      string    `json:"token_id" gorm:"type:varchar(255);uniqueIndex;not null;comment:JWT Token ID"`
	UserID       uint      `json:"user_id" gorm:"not null;index;comment:用户ID"`
	TokenHash    string    `json:"token_hash" gorm:"type:varchar(255);not null;comment:Token哈希值"`
	ExpiresAt    time.Time `json:"expires_at" gorm:"type:datetime(3);not null;index;comment:Token过期时间"`
	RevokedAt    time.Time `json:"revoked_at" gorm:"type:datetime(3);default:CURRENT_TIMESTAMP(3);comment:撤销时间"`
	RevokeReason *string   `json:"revoke_reason" gorm:"type:varchar(100);comment:撤销原因"`

	// 关联关系
	User User `json:"user,omitempty" gorm:"foreignKey:UserID;references:ID"`
}

// TableName 指定表名
func (JWTBlacklist) TableName() string {
	return "jwt_blacklist"
}

// UserStatus 用户状态常量
const (
	UserStatusDisabled = 0 // 禁用
	UserStatusActive   = 1 // 正常
	UserStatusLocked   = 2 // 锁定
)

// RoleStatus 角色状态常量
const (
	RoleStatusDisabled = 0 // 禁用
	RoleStatusActive   = 1 // 启用
)

// PermissionStatus 权限状态常量
const (
	PermissionStatusDisabled = 0 // 禁用
	PermissionStatusActive   = 1 // 启用
)

// ResourceType 资源类型常量
const (
	ResourceTypeMenu   = "menu"   // 菜单
	ResourceTypeButton = "button" // 按钮
	ResourceTypeAPI    = "api"    // API接口
	ResourceTypeData   = "data"   // 数据权限
)

// OperationType 操作类型常量
const (
	OperationTypeCreate = "CREATE" // 创建
	OperationTypeUpdate = "UPDATE" // 更新
	OperationTypeDelete = "DELETE" // 删除
	OperationTypeView   = "VIEW"   // 查看
	OperationTypeExport = "EXPORT" // 导出
	OperationTypeLogin  = "LOGIN"  // 登录
	OperationTypeLogout = "LOGOUT" // 登出
)

// DeviceType 设备类型常量
const (
	DeviceTypePC     = "PC"     // 电脑
	DeviceTypeMobile = "Mobile" // 手机
	DeviceTypeTablet = "Tablet" // 平板
)

// RevokeReason 撤销原因常量
const (
	RevokeReasonLogout         = "logout"          // 用户登出
	RevokeReasonPasswordChange = "password_change" // 密码修改
	RevokeReasonAdminRevoke    = "admin_revoke"    // 管理员撤销
)

// IsActive 检查用户是否处于活跃状态
func (u *User) IsActive() bool {
	return u.Status == UserStatusActive
}

// IsLocked 检查用户是否被锁定
func (u *User) IsLocked() bool {
	if u.Status == UserStatusLocked {
		return true
	}
	if u.LockedUntil != nil && time.Now().Before(*u.LockedUntil) {
		return true
	}
	return false
}

// HasPermission 检查角色是否拥有指定权限
func (r *Role) HasPermission(permissionCode string) bool {
	for _, permission := range r.Permissions {
		if permission.PermissionCode == permissionCode {
			return true
		}
	}
	return false
}

// GetUserPermissions 获取用户的所有权限
// 已弃用：请使用服务层的 GetUserPermissions 方法
func (u *User) GetUserPermissions() []string {
	// 此方法已弃用，因为模型层不应直接访问关联数据
	// 请使用 UserService.GetUserPermissions 或 PermissionMiddleware.GetUserPermissions
	return []string{}
}

// HasPermission 检查用户是否拥有指定权限
// 已弃用：请使用服务层的权限检查方法
func (u *User) HasPermission(permissionCode string) bool {
	// 此方法已弃用，因为模型层不应直接访问关联数据
	// 请使用 PermissionMiddleware.HasPermission 方法
	return false
}
