package database

import (
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	// ErrUserNotFound 用户不存在
	ErrUserNotFound = errors.New("user not found")
	// ErrUserAlreadyExists 用户已存在
	ErrUserAlreadyExists = errors.New("user already exists")
	// ErrInvalidCredential 无效的凭证
	ErrInvalidCredential = errors.New("invalid credential")
	// ErrUpdateFailed 更新失败
	ErrUpdateFailed = errors.New("update failed")
	// ErrDeleteFailed 删除失败
	ErrDeleteFailed = errors.New("delete failed")
)

// DB 全局数据库连接
var DB *gorm.DB

// UserRepository 用户仓库接口
type UserRepository interface {
	// Create 创建用户
	Create(user *User) error
	// FindByID 根据ID获取用户
	FindByID(id uint) (*User, error)
	// FindByUsername 根据用户名获取用户
	FindByUsername(username string) (*User, error)
	// FindByEmail 根据邮箱获取用户
	FindByEmail(email string) (*User, error)
	// FindAll 获取所有用户
	FindAll() ([]*User, error)
	// FindByFilter 根据过滤器获取用户列表
	FindByFilter(filter *UserFilter) ([]*User, int64, error)
	// ExistsByUsername 检查用户名是否存在
	ExistsByUsername(username string) (bool, error)
	// ExistsByEmail 检查邮箱是否存在
	ExistsByEmail(email string) (bool, error)
	// Update 更新用户
	Update(user *User) error
	// Delete 删除用户
	Delete(id uint) error
	// HashPassword 密码加密
	HashPassword(user *User, password string) error
	// VerifyPassword 密码验证
	VerifyPassword(user *User, password string) bool
}

// userRepository GORM用户仓库实现
type userRepository struct {
	db *gorm.DB
}

// NewUserRepository 创建用户仓库实例
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

// Create 创建用户
func (r *userRepository) Create(user *User) error {
	// 检查用户名是否已存在
	exists, err := r.ExistsByUsername(user.Username)
	if err != nil {
		return err
	}
	if exists {
		return ErrUserAlreadyExists
	}

	// 检查邮箱是否已存在
	if user.Email != "" {
		exists, err := r.ExistsByEmail(user.Email)
		if err != nil {
			return err
		}
		if exists {
			return errors.New("email already exists")
		}
	}

	return r.db.Create(user).Error
}

// FindByID 根据ID获取用户
func (r *userRepository) FindByID(id uint) (*User, error) {
	var user User
	err := r.db.First(&user, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

// FindByUsername 根据用户名获取用户
func (r *userRepository) FindByUsername(username string) (*User, error) {
	var user User
	err := r.db.Where("username = ?", username).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

// FindByEmail 根据邮箱获取用户
func (r *userRepository) FindByEmail(email string) (*User, error) {
	var user User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

// FindAll 获取所有用户
func (r *userRepository) FindAll() ([]*User, error) {
	var users []*User
	err := r.db.Order("created_at DESC").Find(&users).Error
	if err != nil {
		return nil, err
	}
	return users, nil
}

// FindByFilter 根据过滤器获取用户列表
func (r *userRepository) FindByFilter(filter *UserFilter) ([]*User, int64, error) {
	var users []*User
	var total int64

	query := r.db.Model(&User{})

	// 应用过滤条件
	if filter.Username != "" {
		query = query.Where("username LIKE ?", "%"+filter.Username+"%")
	}
	if filter.Email != "" {
		query = query.Where("email LIKE ?", "%"+filter.Email+"%")
	}
	if filter.Status != nil {
		query = query.Where("status = ?", *filter.Status)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页
	page := filter.Page
	pageSize := filter.PageSize
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	// 获取数据
	err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&users).Error
	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// ExistsByUsername 检查用户名是否存在
func (r *userRepository) ExistsByUsername(username string) (bool, error) {
	var count int64
	err := r.db.Model(&User{}).Where("username = ?", username).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// ExistsByEmail 检查邮箱是否存在
func (r *userRepository) ExistsByEmail(email string) (bool, error) {
	var count int64
	err := r.db.Model(&User{}).Where("email = ?", email).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// Update 更新用户
func (r *userRepository) Update(user *User) error {
	result := r.db.Save(user)
	if result.Error != nil {
		return ErrUpdateFailed
	}
	if result.RowsAffected == 0 {
		return ErrUserNotFound
	}
	return nil
}

// Delete 删除用户
func (r *userRepository) Delete(id uint) error {
	result := r.db.Delete(&User{}, id)
	if result.Error != nil {
		return ErrDeleteFailed
	}
	if result.RowsAffected == 0 {
		return ErrUserNotFound
	}
	return nil
}

// HashPassword 密码加密
func (r *userRepository) HashPassword(user *User, password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.PasswordHash = string(hash)
	return nil
}

// VerifyPassword 密码验证
func (r *userRepository) VerifyPassword(user *User, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	return err == nil
}

// InitDB 初始化数据库连接
func InitDB(host string, port int, username, password, dbname string) error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		username, password, host, port, dbname)

	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// 自动迁移
	if err := DB.AutoMigrate(&User{}); err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	return nil
}

// GetDB 获取数据库连接
func GetDB() *gorm.DB {
	return DB
}
