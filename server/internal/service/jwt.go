package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims JWT 载荷，sub 形如 "user:12" / "admin:1"
type Claims struct {
	Role   string `json:"role"`
	UserID int64  `json:"user_id"`
	jwt.RegisteredClaims
}

func GenerateToken(secret, role string, userID int64) (string, error) {
	claims := Claims{
		Role:   role,
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
			Issuer:    "xianmengfang",
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
}

func ParseToken(secret, tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("非法签名算法 %v", t.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, errors.New("登录状态无效，请重新登录")
	}
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("登录状态无效，请重新登录")
}
