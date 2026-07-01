package service_auth

import (
	"context"

	dto_auth "blog_server/auth/dto"
)

type AuthService interface {
	Register(ctx context.Context, req dto_auth.RegisterRequest) (*dto_auth.RegisterResponse, error)
	Login(ctx context.Context, req dto_auth.LoginRequest) (*dto_auth.LoginResponse, error)
	RefreshToken(ctx context.Context, refreshToken string) (*dto_auth.TokenResponse, error)
	ChangePassword(ctx context.Context, req dto_auth.ChangePasswordRequest) error
}
