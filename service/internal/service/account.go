package service

import (
	"errors"
	"service/internal/server"
	"service/internal/storage"
	"strings"
	"uuid"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// accountService 系统用户服务。
type accountService struct {
}

// FilterAccount 查询系统用户。
//
// param: ctx 上下文。
// return: 用户集合，错误信息。
func (as *accountService) FilterAccount(ctx *gin.Context) ([]account, error) {
	result, err := storage.RDB.QueryMatrix[account](ctx, `select id,name,login,permission from account order by id;`)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// AppendAccount 添加系统用户。
//
// param: ctx 上下文。
// param: name 用户名称。
// param: login 登录名称。
// param: password 用户密码。
// param: permission 用户权限。
// return: 用户编码，错误信息。
func (as *accountService) AppendAccount(ctx *gin.Context, name *string, login *string, password *string, permission *string) (*string, error) {
	id := strings.ToUpper(uuid.NewV4().String())
	// 计算密码哈希。
	hash, err := bcrypt.GenerateFromPassword([]byte(*password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	// 插入系统用户。
	err = storage.RDB.QueryNaught(ctx, `insert into account(id,name,login,password,permission) values($1,$2,$3,$4,$5)`, id, name, login, string(hash), permission)
	if err != nil {
		return nil, err
	}
	// 返回用户编码。
	return &id, nil
}

// ModifyAccount 修改系统用户。
//
// param: ctx 上下文。
// param: id 用户编码。
// param: name 用户名称。
// param: permission 用户权限。
// return: 用户编码，错误信息。
func (as *accountService) ModifyAccount(ctx *gin.Context, id *string, name *string, permission *string) (*string, error) {
	// 修改系统用户。
	err := storage.RDB.QueryNaught(ctx, `update account set name=$2,permission=$3 where id=$1;`, id, name, permission)
	if err != nil {
		return nil, err
	}
	// 返回用户编码。
	return id, nil
}

// ChangeAccount 改密系统用户。
//
// param: ctx 上下文。
// param: id 用户编码。
// param: password 用户密码。
// return: 用户编码，错误信息。
func (as *accountService) ChangeAccount(ctx *gin.Context, id *string, password *string) (*string, error) {
	// 计算密码哈希。
	hash, err := bcrypt.GenerateFromPassword([]byte(*password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	// 改密系统用户。
	err = storage.RDB.QueryNaught(ctx, `update account set password=$2 where id=$1;`, id, string(hash))
	if err != nil {
		return nil, err
	}
	// 返回用户编码。
	return id, nil
}

// DeleteAccount 删除系统用户。
//
// param: ctx 上下文。
// param: id 用户编码。
// return: 用户编码，错误信息。
func (as *accountService) DeleteAccount(ctx *gin.Context, id *string) (*string, error) {
	// 删除系统用户。
	err := storage.RDB.QueryNaught(ctx, `delete from account where id=$1;`, id)
	if err != nil {
		return nil, err
	}
	// 返回用户编码。
	return id, nil
}

// VerifyAccount 验证系统用户。
//
// param: ctx 上下文。
// param: login 登录名称。
// param: password 用户密码。
// return: 令牌内容，错误信息。
func (as *accountService) VerifyAccount(ctx *gin.Context, login *string, password *string) (*string, error) {
	// 查询系统用户。
	account, err := storage.RDB.QuerySingle[account](ctx, `select id,name,login,password,permission from account where login=$1;`, login)
	if err != nil {
		return nil, err
	}
	if account == nil {
		return nil, errors.New("用户不存在或密码错误。")
	}
	// 比较系统用户。
	if err := bcrypt.CompareHashAndPassword([]byte(*account.Password), []byte(*password)); err != nil {
		return nil, errors.New("用户不存在或密码错误。")
	}
	// 返回用户令牌。
	token, err := server.IssueToken(account.ID, account.Name, account.Login, account.Permission)
	if err != nil {
		return nil, err
	}
	// 返回用户令牌。
	return token, nil
}

// ConsumeAccount 统计用户词元。
//
// param: ctx 上下文。
// param: id 用户编码。
// return: 消费词元，错误信息。
func (as *accountService) ConsumeAccount(ctx *gin.Context, id *string) (*consume, error) {
	// 计算消费词元。
	rows, err := storage.RDB.QueryMatrix[struct {
		Model  *string `db:"model"`
		Prompt *int32  `db:"prompt"`
		Answer *int32  `db:"answer"`
	}](ctx, "select model,sum(prompt) as prompt,sum(answer) as answer from consume where account=$1 group by model;", id)
	if err != nil {
		return nil, err
	}
	var llmPrompt, llmAnswer, slmPrompt, slmAnswer, vemPrompt int32
	for _, row := range rows {
		switch *row.Model {
		case "LLM":
			llmPrompt = *row.Prompt
			llmAnswer = *row.Answer
			break
		case "SLM":
			slmPrompt = *row.Prompt
			slmAnswer = *row.Answer
			break
		case "VEM":
			vemPrompt = *row.Prompt
			break
		}
	}
	// 返回消费词元。
	return &consume{LlmPrompt: llmPrompt, LlmAnswer: llmAnswer, SlmPrompt: slmPrompt, SlmAnswer: slmAnswer, VemPrompt: vemPrompt}, nil
}

// account 系统用户。
type account struct {
	// ID 用户编码。
	ID *string `db:"id" json:"id"`
	// Name 用户名称。
	Name *string `db:"name" json:"name"`
	// Login 登录名称。
	Login *string `db:"login" json:"login"`
	// Password 用户密码。
	Password *string `db:"password" json:"-"`
	// Permission 用户权限。
	Permission *string `db:"permission" json:"permission,omitempty"`
}
