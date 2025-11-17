package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/fail2rest/v2/internal/auth"
	"github.com/fail2rest/v2/internal/models"
)

type AuthHandler struct {
	authService *auth.AuthService
}

func NewAuthHandler(authService *auth.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// Login handles authentication and returns a JWT token
// For simplicity, we accept any token in the request (you can enhance this with proper validation)
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "Invalid request: " + err.Error(),
		})
		return
	}

	// Generate JWT token
	token, expiresAt, err := h.authService.GenerateToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   "Failed to generate token",
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data: models.LoginResponse{
			Token:     token,
			ExpiresAt: expiresAt,
		},
	})
}

