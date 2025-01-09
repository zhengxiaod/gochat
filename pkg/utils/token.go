package utils

import (
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"github.com/spf13/viper"
)

type UserClaims struct {
	UserId uint64 `json:"user_id"`
	jwt.RegisteredClaims
}

// GenerateToken
// 生成 token
func GenerateToken(userId uint64) (string, error) {
	UserClaim := &UserClaims{
		UserId:           userId,
		RegisteredClaims: jwt.RegisteredClaims{},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, UserClaim)
	tokenString, err := token.SignedString([]byte(viper.GetString("jwt.sign_key")))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

// AnalyseToken
// 解析 token
func AnalyseToken(tokenString string) (*UserClaims, error) {
	userClaim := new(UserClaims)
	claims, err := jwt.ParseWithClaims(tokenString, userClaim, func(token *jwt.Token) (interface{}, error) {
		return []byte(viper.GetString("jwt.sign_key")), nil
	})
	if err != nil {
		return nil, err
	}
	if !claims.Valid {
		return nil, fmt.Errorf("analyse Token Error:%v", err)
	}
	return userClaim, nil
}
