package handler_auth

import (
	"net/http"

	dto_auth "blog_server/auth/dto"
	service_auth "blog_server/auth/service" // Adjust import path based on your project structure

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService service_auth.AuthService
}

func NewAuthHandler(authService service_auth.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// Register handles user signup
func (h *AuthHandler) Register(c *gin.Context) {
	var req dto_auth.RegisterRequest

	// Bind JSON body to request DTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload: " + err.Error()})
		return
	}

	// Call service layer
	res, err := h.authService.Register(c.Request.Context(), req)
	if err != nil {
		// Differentiate error types if necessary (e.g., email already registered)
		if err.Error() == "email already registered" {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, res)
}

// Login handles user authentication and issues tokens
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto_auth.LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email or password format"})
		return
	}

	res, err := h.authService.Login(c.Request.Context(), req)
	if err != nil {
		// Intentionally generic message for security, though you can check underlying error
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	c.JSON(http.StatusOK, res)
}

// RefreshToken handles token rotation
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	// Expecting a JSON payload with the refresh token.
	// Alternatively, you could read this from a secure HTTP-only cookie.
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Refresh token is required"})
		return
	}

	res, err := h.authService.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired refresh token"})
		return
	}

	c.JSON(http.StatusOK, res)
}

// ChangePassword updates user password
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	var req dto_auth.ChangePasswordRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// NOTE: In a real app, UserID should be extracted from your Auth Middleware (JWT context)
	// instead of trusting the client to pass it via JSON body. Example:
	// userID, exists := c.Get("user_id")
	// if exists { req.UserID = userID.(string) }

	err := h.authService.ChangePassword(c.Request.Context(), req)
	if err != nil {
		if err.Error() == "invalid old password" {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password updated successfully"})
}
