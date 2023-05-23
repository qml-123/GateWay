package jwt

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/dgrijalva/jwt-go"
	goredis "github.com/go-redis/redis"
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

func GenerateJWT(ctx context.Context, userID int64, expiryTime time.Duration, refresh bool) (string, string, error) {
	claims := &Claims{
		UserID: fmt.Sprintf("%d", userID),
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(expiryTime).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(jwtKey)
	if err != nil {
		return "", "", err
	}

	var refreshTokenString string
	if refresh {
		refreshClaims := &Claims{
			UserID: fmt.Sprintf("%d", userID),
			StandardClaims: jwt.StandardClaims{
				ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
				IssuedAt:  time.Now().Unix(),
			},
		}
		refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
		refreshTokenString, err = refreshToken.SignedString(refreshKey)
		if err != nil {
			return "", "", err
		}
	}

	return signedToken, refreshTokenString, nil
}

// ParseJWT parses a JWT token
func parseJWT(tokenStr string) (*jwt.Token, error) {
	return jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return jwtKey, nil
	})
}

func parseRefreshJWT(tokenStr string) (*jwt.Token, error) {
	return jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return refreshKey, nil
	})
}

func validateToken(tokenString string, secretKey []byte) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})

	if err != nil {
		if ve, ok := err.(*jwt.ValidationError); ok {
			if ve.Errors&jwt.ValidationErrorExpired != 0 {
				// Token 过期，但仍然解析 Claims 数据
				if claims, ok := token.Claims.(*Claims); ok {
					return claims, nil
				}
			}
		}
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}

func CheckTokenExpiry(ctx context.Context, tokenStr string) (int64, string, bool, error) {
	claims, err := validateToken(tokenStr, jwtKey)
	if err != nil {
		logger.Warn(ctx, "jwt-token parseJWT failed, err: %v", err)
		return 0, "", false, err
	}

	var userID int64
	var userIDStr string
	if userID, err = strconv.ParseInt(claims.UserID, 10, 64); err != nil {
		logger.Warn(ctx, "userID ParseInt failed, str: %v", userIDStr)
		return 0, "", false, err
	}

	// 检查是否过期
	if claims.ExpiresAt < time.Now().Unix() {
		// get redis token
		re_token, err := redis.Get(fmt.Sprintf("%d_refresh_token", userID))
		if err != nil && err != goredis.Nil {
			return 0, "", false, err
		}
		if err == goredis.Nil {
			logger.Warn(ctx, "[redis] not found refresh_token, user_id=%v", userID)
			return 0, "", false, fmt.Errorf("not found refresh_token")
		}

		user_re_token, ok := ctx.Value(model.RefreshToken).(string)
		if !ok || len(user_re_token) == 0 {
			logger.Warn(ctx, "not found request refresh_token from context, user_id=%v", userID)
			return 0, "", false, fmt.Errorf("not found refresh_token")
		}

		if user_re_token != re_token {
			logger.Warn(ctx, "request token is valid, request_token: %v, ori_token: %v, user_id=%v", user_re_token, re_token, userID)
			return 0, "", false, fmt.Errorf("not found refresh_token")
		}

		newToken, _, err := GenerateJWT(ctx, userID, 5*time.Minute, false)
		if err != nil {
			logger.Warn(ctx, "GenerateJWT failed, err: %v", err)
			return 0, "", false, err
		}

		return userID, newToken, true, nil
	}
	return userID, tokenStr, false, nil
}
