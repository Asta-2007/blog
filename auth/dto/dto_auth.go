package dto_auth

type RegisterRequest struct {
	Name     string `json:"name"`
	Status   string `json:"status"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterResponse struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	User         UserInfo `json:"user"`
	AccessToken  string   `json:"access_token"`
	RefreshToken string   `json:"refresh_token"`
	ExpAccess    int64    `json:"exp_access"`
	ExpRefresh   int64    `json:"exp_refresh"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpAccess    int64  `json:"exp_access"`
	ExpRefresh   int64  `json:"exp_refresh"`
}

type ChangePasswordRequest struct {
	UserID      string `json:"user_id"`
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

type UserInfo struct {
	ID     string `json:"id"`
	Email  string `json:"email"`
	Status string `json:"status"`
}
