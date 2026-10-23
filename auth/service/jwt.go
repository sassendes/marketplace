package service

import (
	"strconv"
	"time"

	"mercatus/auth/model"

	"github.com/golang-jwt/jwt/v5"
)

const issuer = "mercatus"

type JWTManager struct {
	secret         []byte
	accessDuration time.Duration
	parser         *jwt.Parser
}

func NewJWTManager(secret string, accessDuration time.Duration) *JWTManager {
	return &JWTManager{
		secret:         []byte(secret),
		accessDuration: accessDuration,
		parser: jwt.NewParser(
			jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
			jwt.WithIssuer(issuer),
			jwt.WithExpirationRequired(),
		),
	}
}

type Claims struct {
	UserID int64      `json:"user_id"`
	Role   model.Role `json:"role"`
	jwt.RegisteredClaims
}

func (m *JWTManager) AccessDuration() time.Duration { return m.accessDuration }

func (m *JWTManager) Generate(user *model.User) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID: user.ID,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    issuer,
			Subject:   strconv.FormatInt(user.ID, 10),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.accessDuration)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now.Add(-30 * time.Second)),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.secret)
}

func (m *JWTManager) Verify(tokenStr string) (*Claims, error) {
	claims := &Claims{}
	token, err := m.parser.ParseWithClaims(tokenStr, claims, func(*jwt.Token) (any, error) { return m.secret, nil })
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, jwt.ErrTokenInvalidClaims
	}
	return claims, nil
}
