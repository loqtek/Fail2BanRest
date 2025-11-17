package models

import "time"

// APIResponse is the standard API response structure
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
}

// LoginRequest represents a login request
type LoginRequest struct {
	Token string `json:"token" binding:"required"`
}

// LoginResponse represents a login response
type LoginResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

// JailInfo represents information about a jail
type JailInfo struct {
	Name      string                 `json:"name"`
	Status    map[string]interface{} `json:"status,omitempty"`
	BannedIPs []string               `json:"banned_ips,omitempty"`
	Stats     map[string]interface{} `json:"stats,omitempty"`
}

// BanRequest represents a request to ban an IP
type BanRequest struct {
	IP string `json:"ip" binding:"required"`
}

// UnbanRequest represents a request to unban an IP
type UnbanRequest struct {
	IP string `json:"ip" binding:"required"`
}

// StatusResponse represents the status of fail2ban
type StatusResponse struct {
	Status    map[string]interface{} `json:"status"`
	Timestamp int64                  `json:"timestamp"`
}

// StatsResponse represents statistics
type StatsResponse struct {
	Stats     map[string]interface{} `json:"stats"`
	Timestamp int64                  `json:"timestamp"`
}
