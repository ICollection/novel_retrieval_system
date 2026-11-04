package service

import (
	"context"
	"errors"
	"service/internal/storage"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// AccountService 系统用户服务。
type AccountService struct {
	postgre *storage.Postgre
}

// NewAccountService 创建系统用户服务。
func NewAccountService(postgre *storage.Postgre) *AccountService {
	return &AccountService{postgre: postgre}
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
	err = service.postgre.QueryNaught(ctx, `insert into account(id,name,login,password,permission) values($1,$2,$3,$4,$5)`, id, name, login, string(hash), permission)
	if err != nil {
		return nil, err
	}
	// 返沪用户编码。
	return &id, nil
}

// Modify 修改系统用户。
func (service *AccountService) Modify(ctx context.Context, id *string, name *string, permission *string) (*string, error) {
	// 修改系统用户。
	err := service.postgre.QueryNaught(ctx, `update account set name=$2,permission=$3 where id=$1;`, id, name, permission)
	if err != nil {
		return nil, err
	}
	// 返回用户编码。
	return id, nil
}

// Change 改密系统用户。
func (service *AccountService) Change(ctx context.Context, id *string, password *string) (*string, error) {
	// 计算密码哈希。
	hash, err := bcrypt.GenerateFromPassword([]byte(*password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	// 改密系统用户。
	err = service.postgre.QueryNaught(ctx, `update account set password=$2 where id=$1;`, id, string(hash))
	if err != nil {
		return nil, err
	}
	// 返回用户编码。
	return id, nil
}

// Delete 删除系统用户。
func (service *AccountService) Delete(ctx context.Context, id *string) (*string, error) {
	// 删除系统用户。
	err := service.postgre.QueryNaught(ctx, `delete from account where id=$1;`, id)
	if err != nil {
		return nil, err
	}
	// 返回用户编码。
	return id, nil
}

// Verify 验证系统用户。
func (service *AccountService) Verify(ctx context.Context, login *string, password *string) (*string, error) {
	// 查询系统用户。
	account, err := service.postgre.QuerySingle[Account](ctx, `select id,name,login,password,permission from account where login=$1;`, login)
	if err != nil {
		return nil, errors.New("用户不存在或密码错误。")
	}
	// 比较系统用户。
	if err := bcrypt.CompareHashAndPassword([]byte(*account.Password), []byte(*password)); err != nil {
		return nil, errors.New("用户不存在或密码错误。")
	}
	// 返回用户令牌。
	token, err := IssueToken(account.Id, account.Name, account.Login, account.Password)
	if err != nil {
		return nil, err
	}
	// 返回用户令牌。
	return token, nil
}

// Filter 查询系统用户。
func (service *AccountService) Filter(ctx context.Context) ([]Account, error) {
	result, err := service.postgre.QueryMatrix[Account](ctx, `select id,name,login,password,permission from account order by id;`)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Account 系统用户。
type Account struct {
	Id         *string `db:"id" json:"id"`
	Name       *string `db:"name" json:"name"`
	Login      *string `db:"login" json:"login"`
	Password   *string `db:"password" json:"password"`
	Permission *string `db:"permission" json:"permission"`
}
