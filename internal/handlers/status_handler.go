package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/fail2rest/v2/internal/fail2ban"
	"github.com/fail2rest/v2/internal/models"
)

type StatusHandler struct {
	f2bClient *fail2ban.Client
}

func NewStatusHandler(f2bClient *fail2ban.Client) *StatusHandler {
	return &StatusHandler{
		f2bClient: f2bClient,
	}
}

// GetStatus returns the overall status of fail2ban
func (h *StatusHandler) GetStatus(c *gin.Context) {
	status, err := h.f2bClient.GetStatus()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   "Failed to get status: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data: models.StatusResponse{
			Status:    status,
			Timestamp: time.Now().Unix(),
		},
	})
}

