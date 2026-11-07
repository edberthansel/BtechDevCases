package service

import (
	"context"
	"net/mail"
	"strings"

	"github.com/edberthansel/BtechDevCases/backend/internal/model"
	"github.com/edberthansel/BtechDevCases/backend/internal/repository"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// ValidationError membawa pesan error per field.
type ValidationError struct {
	Fields map[string]string
}

func (e *ValidationError) Error() string { return "validation failed" }

type AuthService struct {
	users          *repository.UserRepository
	initialBalance int64
}

func NewAuthService(users *repository.UserRepository, initialBalance int64) *AuthService {
	return &AuthService{users: users, initialBalance: initialBalance}
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