package service_auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	dto_auth "blog_server/auth/dto"
	model_auth "blog_server/auth/model"
	repository_auth "blog_server/auth/repository"
	mysql_auth "blog_server/auth/repository/mysql"
	jwts "blog_server/share/jwt"
	utility "blog_server/share/utils"

	"go.mongodb.org/mongo-driver/v2/bson"
	"golang.org/x/crypto/bcrypt"
)

type authService struct {
	cfg  string
	repo repository_auth.AuthEntity
}

func NewAuthService(cfg string, db *sql.DB) AuthService {
	return &authService{
		cfg:  cfg,
		repo: mysql_auth.NewAuthRepository(db),
	}
}

func (a *authService) Register(
	ctx context.Context,
	req dto_auth.RegisterRequest,
) (*dto_auth.RegisterResponse, error) {
	// 1. Check if email already exists
	existing, err := a.repo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to check email: %w", err)
	}
	if existing != nil {
		return nil, errors.New("email already registered")
	}

	// 2. Hash password
	pass, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	now := time.Now().UTC()

	// 3. Create user entity
	user := model_auth.Entity{
		ID:        utility.UniqueID(32),
		Name:      req.Name,
		Status:    req.Status,
		Email:     req.Email,
		Password:  string(pass),
		Created:   now,
		UpdatedAt: now,
	}

	// 4. Save to DB
	if err := a.repo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// 5. Return response (never return password)
	return &dto_auth.RegisterResponse{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
	}, nil
}

// Login implements [AuthService].
func (a *authService) Login(ctx context.Context, req dto_auth.LoginRequest) (*dto_auth.LoginResponse, error) {
	user, err := a.repo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to find email %v", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, fmt.Errorf("incorect password %v", err)
	}

	j := jwts.NewJWTUtils(a.cfg)

	access, aT, err := j.GenerateAccessToken(user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to generated access token")
	}

	refresh, rT, err := j.GenerateRefreshToken(user.ID, user.Email, user.Status)
	if err != nil {
		return nil, fmt.Errorf("failed to generated refresh token")
	}
	token := dto_auth.LoginResponse{
		User: dto_auth.UserInfo{
			ID:     user.ID,
			Status: user.Status,
			Email:  user.Email,
		},
		AccessToken:  access,
		RefreshToken: refresh,
		ExpAccess:    aT.Unix(),
		ExpRefresh:   rT.Unix(),
	}

	return &token, nil
}

// RefreshToken implements [AuthService].
func (a *authService) RefreshToken(ctx context.Context, refreshToken string) (*dto_auth.TokenResponse, error) {
	j := jwts.NewJWTUtils(a.cfg)

	// Validate refresh token
	claims, err := j.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	// Ensure user still exists
	user, err := a.repo.GetByID(ctx, claims.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if user == nil {
		return nil, errors.New("user not found")
	}

	// Optional: verify status hasn't changed
	if user.Status != claims.Role {
		return nil, errors.New("user status changed")
	}

	// Generate new tokens
	access, accessExp, err := j.GenerateAccessToken(user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refresh, refreshExp, err := j.GenerateRefreshToken(
		user.ID,
		user.Email,
		user.Status,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &dto_auth.TokenResponse{
		AccessToken:  access,
		RefreshToken: refresh,
		ExpAccess:    accessExp.Unix(),
		ExpRefresh:   refreshExp.Unix(),
	}, nil
}

func (a *authService) ChangePassword(
	ctx context.Context,
	req dto_auth.ChangePasswordRequest,
) error {
	// 1. Get user
	user, err := a.repo.GetByID(ctx, req.UserID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	if user == nil {
		return errors.New("user not found")
	}

	// 2. Check old password
	if err := bcrypt.CompareHashAndPassword(
		[]byte(user.Password),
		[]byte(req.OldPassword),
	); err != nil {
		return errors.New("invalid old password")
	}

	// 3. Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword(
		[]byte(req.NewPassword),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// 4. Update password in DB
	user.Password = string(hashedPassword)
	user.UpdatedAt = time.Now()

	if err := a.repo.Update(ctx, bson.M{
		"id":       user.ID,
		"password": user.Password,
		"updated":  user.UpdatedAt,
	}); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}
