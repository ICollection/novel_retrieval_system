package service

import (
	"errors"
	"fmt"
	"service/internal/config"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TokenService 登录令牌服务。
type TokenService struct {
}

// Sign 签发登录令牌。
func (service *TokenService) Sign(account *Account) (*string, error) {
	// 组装令牌载荷。
	moment := time.Now()
	expire := moment.Add(time.Duration(config.Config.Token.Expire) * time.Hour)
	claims := Claims{
		ID:         account.Id,
		Name:       account.Name,
		Login:      account.Login,
		Permission: account.Permission,
		Subject:    *account.Id,
		IssuedAt:   jwt.NewNumericDate(moment),
		NotBefore:  jwt.NewNumericDate(moment),
		ExpiresAt:  jwt.NewNumericDate(expire),
	}
	// 签发登录令牌。
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(config.Config.Token.Secret)
	if err != nil {
		return nil, fmt.Errorf("签发登录令牌失败，详细错误信息：%w", err)
	}
	// 返回登录令牌。
	return &token, nil
}

// Parse 解析并校验登录令牌。
func (service *TokenService) Parse(raw string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(raw, claims, func(*jwt.Token) (any, error) { return config.Config.Token.Secret, nil }, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}), jwt.WithExpirationRequired())
	if err != nil || !token.Valid {
		return nil, errors.New("登录令牌已经失效。")
	}
	return claims, nil
}

// Claims 登录令牌载荷，存放用户密码以外的各项信息。
type Claims struct {
	ID         *string `json:"id"`
	Name       *string `json:"name"`
	Login      *string `json:"login"`
	Permission *string `json:"permission"`
	jwt.RegisteredClaims
}
