package server

import (
	"errors"
	"fmt"
	"service/internal/config"
	"strings"
	"time"
	"uuid"

	"github.com/golang-jwt/jwt/v5"
)

// Claims 令牌声明。
type Claims struct {
	Name       *string `json:"name"`
	Login      *string `json:"login"`
	Permission *string `json:"permission"`
	jwt.RegisteredClaims
}

// HasPermission 是否具有权限。
func (claims *Claims) HasPermission(permission string) bool {
	if claims.Permission == nil {
		return false
	}
	if *claims.Permission == permission {
		return true
	}
	if strings.HasPrefix(*claims.Permission, permission+",") == true {
		return true
	}
	if strings.HasSuffix(*claims.Permission, ","+permission) == true {
		return true
	}
	if strings.Contains(*claims.Permission, ","+permission+",") == true {
		return true
	}
	return false
}

// IssueToken 签发登录令牌。
func IssueToken(id *string, name *string, login *string, permission *string) (*string, error) {
	// 组装令牌载荷。
	moment := time.Now().UTC()
	expire := moment.Add(time.Duration(config.Config.Station.Expire) * time.Hour)
	claims := Claims{
		Name:       name,
		Login:      login,
		Permission: permission,
		Issuer:     "service",
		Subject:    *id,
		Audience:   []string{"surface"},
		ExpiresAt:  jwt.NewNumericDate(expire),
		NotBefore:  jwt.NewNumericDate(moment),
		IssuedAt:   jwt.NewNumericDate(moment),
		ID:         strings.ToUpper(uuid.NewV4().String()),
	}
	// 签发登录令牌。
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(config.Config.Station.Secret))
	if err != nil {
		return nil, fmt.Errorf("签发登录令牌失败，详细错误信息：%w", err)
	}
	// 返回登录令牌。
	return &token, nil
}

// ParseToken 验证登录令牌。
func ParseToken(buffer string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(buffer, claims, func(*jwt.Token) (any, error) { return []byte(config.Config.Station.Secret), nil }, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}), jwt.WithExpirationRequired())
	if err != nil {
		return nil, fmt.Errorf("验证登录令牌失败，详细错误信息：%w", err)
	}
	if token.Valid == false {
		return nil, errors.New("登录令牌已经失效。")
	}
	return claims, nil
}
