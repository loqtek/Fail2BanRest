package auth

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	jwtSecret   []byte
	tokenExpiry time.Duration
	apiKeys     map[string]bool
	users       map[string]string // username -> bcrypt hashed password
}

type AuthConfig struct {
	APIKeys []string
	Users   map[string]string // username -> bcrypt hashed password
}

func NewAuthService(jwtSecret string, tokenExpiry time.Duration, authConfig AuthConfig) *AuthService {
	// Build API key map for fast lookup
	apiKeyMap := make(map[string]bool)
	for _, key := range authConfig.APIKeys {
		if key != "" {
			apiKeyMap[key] = true
		}
	}

	return &AuthService{
		jwtSecret:   []byte(jwtSecret),
		tokenExpiry: tokenExpiry,
		apiKeys:     apiKeyMap,
		users:       authConfig.Users,
	}
}

type Claims struct {
	Authorized bool `json:"authorized"`
	jwt.RegisteredClaims
}

func (a *AuthService) GenerateToken() (string, time.Time, error) {
	expirationTime := time.Now().Add(a.tokenExpiry)
	claims := &Claims{
		Authorized: true,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(a.jwtSecret)
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expirationTime, nil
}

func (a *AuthService) ValidateToken(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return a.jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

// ValidateAPIKey checks if the provided API key is valid
func (a *AuthService) ValidateAPIKey(apiKey string) bool {
	return a.apiKeys[apiKey]
}

// ValidateCredentials checks if username and password are valid
func (a *AuthService) ValidateCredentials(username, password string) bool {
	hashedPassword, exists := a.users[username]
	if !exists {
		return false
	}

	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

// HasAuthConfigured returns true if any authentication method is configured
func (a *AuthService) HasAuthConfigured() bool {
	return len(a.apiKeys) > 0 || len(a.users) > 0
}

func (a *AuthService) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Authorization header required",
			})
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Invalid authorization header format",
			})
			c.Abort()
			return
		}

		tokenString := parts[1]
		claims, err := a.ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Invalid or expired token",
			})
			c.Abort()
			return
		}

		if !claims.Authorized {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "Not authorized",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
