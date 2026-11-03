package service

import (
	"context"
	"service/internal/storage"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// AccountService 系统用户服务。
type AccountService struct {
	postgresql *storage.PostgreSQL
}

// NewAccountService 创建系统用户服务。
func NewAccountService(postgresql *storage.PostgreSQL) *AccountService {
	return &AccountService{postgresql: postgresql}
}

// Append 添加系统用户。
func (service *AccountService) Append(ctx context.Context, name *string, login *string, password *string, permission *string) (*string, error) {
	id := strings.ToUpper(uuid.NewString())
	// 计算密码哈希。
	hash, err := bcrypt.GenerateFromPassword([]byte(*password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	// 插入系统用户。
	err = service.postgresql.QueryEmpty(ctx, `insert into account(id,name,login,password,permission) values($1,$2,$3,$4,$5)`, id, name, login, string(hash), permission)
	if err != nil {
		return nil, err
	}
	return &id, nil
}

// Modify 修改系统用户。
func (service *AccountService) Modify(ctx context.Context, id *string, name *string, permission *string) (*string, error) {
	return id, nil
}

// Change 改密系统用户。
func (service *AccountService) Change(ctx context.Context, id *string, password *string) (*string, error) {
	return id, nil
}

// Delete 删除系统用户。
func (service *AccountService) Delete(ctx context.Context, id *string) (*string, error) {
	return id, nil
}

// Verify 验证系统用户。
func (service *AccountService) Verify(ctx context.Context, login *string, password *string) (*string, error) {
	return nil, nil
}

// Filter 查询系统用户。
func (service *AccountService) Filter(ctx context.Context) ([]Account, error) {
	return nil, nil
}

// Account 系统用户。
type Account struct {
	Id         *string `db:"id" json:"id"`
	Name       *string `db:"name" json:"name"`
	Login      *string `db:"login" json:"login"`
	Password   *string `db:"password" json:"password"`
	Permission *string `db:"permission" json:"permission"`
}

func derefOrEmpty(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}
