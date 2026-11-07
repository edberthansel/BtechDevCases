package token

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var ErrInvalidToken = errors.New("invalid or expired token")

type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

type Manager struct {
	secret []byte
	ttl    time.Duration
}

func NewManager(secret string, ttl time.Duration) *Manager {
	return &Manager{secret: []byte(secret), ttl: ttl}
}

// Generate membuat token baru dan mengembalikan waktu kedaluwarsanya.
func (m *Manager) Generate(userID, email string) (string, time.Time, error) {
	now := time.Now()
	exp := now.Add(m.ttl)

	claims := Claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(exp),
		},
	}

	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.secret)
	if err != nil {
		return "", time.Time{}, err
	}
	return signed, exp, nil
}

// Parse memvalidasi token. Dipakai di middleware (Step 9).
func (m *Manager) Parse(tokenString string) (*Claims, error) {
	claims := &Claims{}
	tok, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(t *jwt.Token) (any, error) { return m.secret, nil },
		jwt.WithValidMethods([]string{"HS256"}), // tolak algoritma lain (mis. "none")
		jwt.WithExpirationRequired(),            // tolak token tanpa exp
	)
	if err != nil || !tok.Valid {
		return nil, ErrInvalidToken
	}
	return claims, nil
}