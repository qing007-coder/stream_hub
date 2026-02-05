package security

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"stream_hub/internal/infra"
	"stream_hub/pkg/errors"
	"stream_hub/pkg/model/auth"
	"stream_hub/pkg/model/config"
	"stream_hub/pkg/utils"
	"time"
)

type Auth struct {
	redis           *infra.Redis
	accessDuration  time.Duration
	refreshDuration time.Duration
	secretKey       string
}

func NewAuth(conf *config.CommonConfig) *Auth {
	return &Auth{
		redis:           infra.NewRedis(conf),
		accessDuration:  time.Hour * time.Duration(conf.JWT.Access),
		refreshDuration: time.Hour * time.Duration(conf.JWT.Refresh),
		secretKey:       conf.SecretKey,
	}
}

func (a *Auth) GenerateToken(claims *auth.Claims) (map[string]string, error) {
	accessToken, err := a.createAccessToken(claims)
	if err != nil {
		return nil, err
	}

	refreshToken := utils.CreateUUID()
	data, err := json.Marshal(claims)
	if err != nil {
		return nil, err
	}

	if err := a.redis.Set(context.Background(), refreshToken, data, a.refreshDuration); err != nil {
		return nil, err
	}
	if err := a.redis.Set(context.Background(), claims.UserID, refreshToken, a.refreshDuration); err != nil {
		return nil, err
	}

	tokenMap := make(map[string]string)
	tokenMap["access_token"] = accessToken
	tokenMap["refresh_token"] = refreshToken
	return tokenMap, nil
}

func (a *Auth) ParseToken(accessToken string) (*auth.Claims, error) {
	token, err := jwt.Parse(accessToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(a.secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return &auth.Claims{
			UserID:    claims["user_id"].(string),
			Role:      claims["role"].(string),
			CreatedAt: int64(claims["created_at"].(float64)),
		}, nil
	} else {
		return nil, err
	}
}

func (a *Auth) RefreshToken(refreshToken string) (map[string]string, error) {
	ctx := context.Background()
	existed, err := a.redis.IsExisted(ctx, refreshToken)
	if err != nil {
		return nil, err
	}

	if !existed {
		return nil, errors.RefreshTokenExpiredError
	}

	data, err := a.redis.Get(ctx, refreshToken)
	if err != nil {
		return nil, err
	}

	var claims auth.Claims
	if err := json.Unmarshal(data, &claims); err != nil {
		return nil, errors.UnmarshalError
	}

	if err := a.redis.Expire(ctx, refreshToken, a.refreshDuration); err != nil {
		return nil, err
	}
	if err := a.redis.Expire(ctx, claims.UserID, a.refreshDuration); err != nil {
		return nil, err
	}

	accessToken, err := a.createAccessToken(&claims)
	if err != nil {
		return nil, err
	}

	tokenMap := make(map[string]string)
	tokenMap["access_token"] = accessToken
	tokenMap["refresh_token"] = refreshToken
	return tokenMap, nil
}

func (a *Auth) createAccessToken(claims *auth.Claims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"authorized": true,
		"user_id":    claims.UserID,
		"role":       claims.Role,
		"created_at": claims.CreatedAt,
		"exp":        time.Now().Add(a.accessDuration).Unix(), // 过期时间
	})

	return token.SignedString([]byte(a.secretKey))
}
