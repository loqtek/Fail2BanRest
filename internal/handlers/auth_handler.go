package handlers

import (
	"net/http"

	"github.com/fail2rest/v2/internal/auth"
	"github.com/fail2rest/v2/internal/models"
	"github.com/gin-gonic/gin"
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
// Supports API key or username/password authentication
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "Invalid request: " + err.Error(),
		})
		return
	}

	// Check if authentication is configured
	if !h.authService.HasAuthConfigured() {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   "Authentication not configured. Please configure API keys or users in config file.",
		})
		return
	}

	// Validate credentials - try API key first, then username/password
	authenticated := false

	if req.APIKey != "" {
		authenticated = h.authService.ValidateAPIKey(req.APIKey)
		if !authenticated {
			c.JSON(http.StatusUnauthorized, models.APIResponse{
				Success: false,
				Error:   "Invalid API key",
			})
			return
		}
	} else if req.Username != "" && req.Password != "" {
		authenticated = h.authService.ValidateCredentials(req.Username, req.Password)
		if !authenticated {
			c.JSON(http.StatusUnauthorized, models.APIResponse{
				Success: false,
				Error:   "Invalid username or password",
			})
			return
		}
	} else {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "Either 'api_key' or 'username' and 'password' must be provided",
		})
		return
	}

	if !authenticated {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error:   "Authentication failed",
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
