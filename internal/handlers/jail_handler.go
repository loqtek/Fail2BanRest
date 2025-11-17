package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/fail2rest/v2/internal/fail2ban"
	"github.com/fail2rest/v2/internal/models"
)

type JailHandler struct {
	f2bClient *fail2ban.Client
}

func NewJailHandler(f2bClient *fail2ban.Client) *JailHandler {
	return &JailHandler{
		f2bClient: f2bClient,
	}
}

// GetJails returns a list of all jails
func (h *JailHandler) GetJails(c *gin.Context) {
	jails, err := h.f2bClient.GetJails()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   "Failed to get jails: " + err.Error(),
		})
		return
	}

	var jailInfos []models.JailInfo
	for _, jail := range jails {
		jailInfos = append(jailInfos, models.JailInfo{
			Name: jail,
		})
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    jailInfos,
	})
}

// GetJail returns detailed information about a specific jail
func (h *JailHandler) GetJail(c *gin.Context) {
	jailName := c.Param("name")
	if jailName == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "Jail name is required",
		})
		return
	}

	status, err := h.f2bClient.GetJailStatus(jailName)
	if err != nil {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error:   "Jail not found or error: " + err.Error(),
		})
		return
	}

	jailInfo := models.JailInfo{
		Name:   jailName,
		Status: status,
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    jailInfo,
	})
}

// GetJailStatus returns the status of a specific jail
func (h *JailHandler) GetJailStatus(c *gin.Context) {
	jailName := c.Param("name")
	if jailName == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "Jail name is required",
		})
		return
	}

	status, err := h.f2bClient.GetJailStatus(jailName)
	if err != nil {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error:   "Jail not found or error: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    status,
	})
}

// StartJail starts a jail
func (h *JailHandler) StartJail(c *gin.Context) {
	jailName := c.Param("name")
	if jailName == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "Jail name is required",
		})
		return
	}

	if err := h.f2bClient.StartJail(jailName); err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   "Failed to start jail: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Jail started successfully",
	})
}

// StopJail stops a jail
func (h *JailHandler) StopJail(c *gin.Context) {
	jailName := c.Param("name")
	if jailName == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "Jail name is required",
		})
		return
	}

	if err := h.f2bClient.StopJail(jailName); err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   "Failed to stop jail: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Jail stopped successfully",
	})
}

// RestartJail restarts a jail
func (h *JailHandler) RestartJail(c *gin.Context) {
	jailName := c.Param("name")
	if jailName == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "Jail name is required",
		})
		return
	}

	if err := h.f2bClient.RestartJail(jailName); err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   "Failed to restart jail: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Jail restarted successfully",
	})
}

// ReloadJail reloads a jail configuration
func (h *JailHandler) ReloadJail(c *gin.Context) {
	jailName := c.Param("name")
	if jailName == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "Jail name is required",
		})
		return
	}

	if err := h.f2bClient.ReloadJail(jailName); err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   "Failed to reload jail: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Jail reloaded successfully",
	})
}

