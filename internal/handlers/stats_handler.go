package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/fail2rest/v2/internal/fail2ban"
	"github.com/fail2rest/v2/internal/models"
)

type StatsHandler struct {
	f2bClient *fail2ban.Client
}

func NewStatsHandler(f2bClient *fail2ban.Client) *StatsHandler {
	return &StatsHandler{
		f2bClient: f2bClient,
	}
}

// GetStats returns overall statistics
func (h *StatsHandler) GetStats(c *gin.Context) {
	stats, err := h.f2bClient.GetOverallStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   "Failed to get statistics: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data: models.StatsResponse{
			Stats:     stats,
			Timestamp: time.Now().Unix(),
		},
	})
}

// GetJailStats returns statistics for a specific jail
func (h *StatsHandler) GetJailStats(c *gin.Context) {
	jailName := c.Param("name")
	if jailName == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "Jail name is required",
		})
		return
	}

	stats, err := h.f2bClient.GetJailStats(jailName)
	if err != nil {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error:   "Jail not found or error: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data: models.StatsResponse{
			Stats:     stats,
			Timestamp: time.Now().Unix(),
		},
	})
}

