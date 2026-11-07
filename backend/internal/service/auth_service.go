package service

import (
	"context"
	"errors"
	"net/mail"
	"strings"
	"time"

	"github.com/edberthansel/BtechDevCases/backend/internal/model"
	"github.com/edberthansel/BtechDevCases/backend/internal/repository"
	"github.com/edberthansel/BtechDevCases/backend/internal/token"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var ErrInvalidCredentials = errors.New("invalid email or password")

// ValidationError membawa pesan error per field.
type ValidationError struct {
	Fields map[string]string
}

func (e *ValidationError) Error() string { return "validation failed" }

type LoginResult struct {
	Token     string
	ExpiresAt time.Time
	User      *model.User
}

type AuthService struct {
	users          *repository.UserRepository
	tokens         *token.Manager
	initialBalance int64
	dummyHash      []byte
}

func NewAuthService(users *repository.UserRepository, tokens *token.Manager, initialBalance int64) *AuthService {
	// Hash palsu untuk menyamakan waktu respons saat email tidak ditemukan.
	dummy, _ := bcrypt.GenerateFromPassword([]byte("dummy-password"), bcrypt.DefaultCost)
	return &AuthService{
		users:          users,
		tokens:         tokens,
		initialBalance: initialBalance,
		dummyHash:      dummy,
	}
}

func (s *AuthService) Register(ctx context.Context, email, password, confirmPassword string) (*model.User, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	fields := map[string]string{}

	switch {
	case email == "":
		fields["email"] = "email is required"
	case !validEmail(email):
		fields["email"] = "email format is invalid"
	}

	switch {
	case password == "":
		fields["password"] = "password is required"
	case len(password) < 8:
		fields["password"] = "password must be at least 8 characters"
	case len(password) > 72: // batas bcrypt
		fields["password"] = "password must be at most 72 characters"
	}

	switch {
	case confirmPassword == "":
		fields["confirmPassword"] = "confirmPassword is required"
	case password != confirmPassword:
		fields["confirmPassword"] = "passwords do not match"
	}

	if len(fields) > 0 {
		return nil, &ValidationError{Fields: fields}
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: string(hash),
	}

	if err := s.users.CreateWithWallet(ctx, user, s.initialBalance); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *AuthService) Login(ctx context.Context, email, password string) (*LoginResult, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	fields := map[string]string{}
	if email == "" {
		fields["email"] = "email is required"
	}
	if password == "" {
		fields["password"] = "password is required"
	}
	if len(fields) > 0 {
		return nil, &ValidationError{Fields: fields}
	}

	user, err := s.users.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			_ = bcrypt.CompareHashAndPassword(s.dummyHash, []byte(password))
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	tok, exp, err := s.tokens.Generate(user.ID.String(), user.Email)
	if err != nil {
		return nil, err
	}

	return &LoginResult{Token: tok, ExpiresAt: exp, User: user}, nil
}

func validEmail(email string) bool {
	if len(email) > 254 {
		return false
	}
	addr, err := mail.ParseAddress(email)
	if err != nil || addr.Address != email { // tolak format "Nama <a@b.com>"
		return false
	}
	at := strings.LastIndex(email, "@")
	return strings.Contains(email[at+1:], ".") // domain harus punya titik
}