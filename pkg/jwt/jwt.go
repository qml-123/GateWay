package jwt

import (
	"context"
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/qml-123/GateWay/model"
	"github.com/qml-123/GateWay/pkg/redis"
	"github.com/qml-123/app_log/logger"
)

type Claims struct {
	UserID string `json:"user_id"`
	jwt.StandardClaims
}

var jwtKey = []byte("your_secret_key")
var refreshKey = []byte("your_refresh_secret_key")

func GenerateToken(userID string, expiryTime time.Duration, refresh bool) (string, string, error) {
	expirationTime := time.Now().Add(expiryTime)
	claims := &Claims{
		UserID: userID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		return "", "", err
	}

	var refreshTokenString string
	if refresh {
		refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		refreshTokenString, err = refreshToken.SignedString(refreshKey)
		if err != nil {
			return "", "", err
		}
	}

	return tokenString, refreshTokenString, nil
}

func parseToken(tokenStr string) (*Claims, error) {
	claims := &Claims{}
	tkn, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil {
		return nil, err
	}
	if !tkn.Valid {
		return nil, jwt.ErrSignatureInvalid
	}

	return claims, nil
}

func CheckTokenExpiry(ctx context.Context, tokenStr string) (string, bool, error) {
	claims, err := parseToken(tokenStr)
	if err != nil {
		return "", false, err
	}

	// 检查是否过期
	if time.Unix(claims.ExpiresAt, 0).Before(time.Now()) {
		// get redis token
		re_token, err := redis.Get(fmt.Sprintf("%d_refresh_token", claims.UserID))
		if err != nil {
			return "", false, err
		}
		if len(re_token) == 0 {
			logger.Warn(ctx, "[redis] not found refresh_token, user_id=%v", claims.UserID)
			return "", false, fmt.Errorf("not found refresh_token")
		}

		user_re_token, ok := ctx.Value(model.RefreshToken).(string)
		if !ok || len(user_re_token) == 0 {
			logger.Warn(ctx, "not found request refresh_token, user_id=%v", claims.UserID)
			return "", false, fmt.Errorf("not found refresh_token")
		}

		if user_re_token != re_token {
			logger.Warn(ctx, "request token is valid, request_token: %v, ori_token: %v, user_id=%v", user_re_token, re_token, claims.UserID)
			return "", false, fmt.Errorf("not found refresh_token")
		}

		newToken, _, err := GenerateToken(claims.UserID, 5*time.Minute, false)
		if err != nil {
			return "", false, err
		}

		return newToken, true, nil
	}
	return tokenStr, false, nil
}
