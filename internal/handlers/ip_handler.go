package handlers

import (
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/fail2rest/v2/internal/fail2ban"
	"github.com/fail2rest/v2/internal/models"
)

type IPHandler struct {
	f2bClient *fail2ban.Client
}

func NewIPHandler(f2bClient *fail2ban.Client) *IPHandler {
	return &IPHandler{
		f2bClient: f2bClient,
	}
}

// GetBannedIPs returns a list of banned IPs for a jail
func (h *IPHandler) GetBannedIPs(c *gin.Context) {
	jailName := c.Param("name")
	if jailName == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "Jail name is required",
		})
		return
	}

	ips, err := h.f2bClient.GetBannedIPs(jailName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   "Failed to get banned IPs: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    gin.H{"jail": jailName, "banned_ips": ips},
	})
}

// BanIP bans an IP address in a jail
func (h *IPHandler) BanIP(c *gin.Context) {
	jailName := c.Param("name")
	if jailName == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "Jail name is required",
		})
		return
	}

	var req models.BanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "Invalid request: " + err.Error(),
		})
		return
	}

	// Validate IP address
	if net.ParseIP(req.IP) == nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "Invalid IP address",
		})
		return
	}

	if err := h.f2bClient.BanIP(jailName, req.IP); err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   "Failed to ban IP: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "IP banned successfully",
		Data:    gin.H{"jail": jailName, "ip": req.IP},
	})
}

// UnbanIP unbans an IP address in a jail
func (h *IPHandler) UnbanIP(c *gin.Context) {
	jailName := c.Param("name")
	if jailName == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "Jail name is required",
		})
		return
	}

	var req models.UnbanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "Invalid request: " + err.Error(),
		})
		return
	}

	// Validate IP address
	if net.ParseIP(req.IP) == nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "Invalid IP address",
		})
		return
	}

	if err := h.f2bClient.UnbanIP(jailName, req.IP); err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   "Failed to unban IP: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "IP unbanned successfully",
		Data:    gin.H{"jail": jailName, "ip": req.IP},
	})
}

