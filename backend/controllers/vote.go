package controllers

import (
	"net/http"
	"strings"

	"pulsevote/models"
	"pulsevote/services"

	"github.com/gin-gonic/gin"
)

// VoteController handles voting HTTP endpoints
type VoteController struct {
	pollService *services.PollService
}

// NewVoteController creates a new VoteController
func NewVoteController(pollService *services.PollService) *VoteController {
	return &VoteController{
		pollService: pollService,
	}
}

// SubmitVote handles POST /api/polls/:shareCode/vote
func (ctrl *VoteController) SubmitVote(c *gin.Context) {
	shareCode := strings.TrimSpace(c.Param("shareCode"))
	if shareCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Share code is required"})
		return
	}

	var req models.VoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid vote payload"})
		return
	}

	optionID := strings.TrimSpace(req.OptionID)
	if optionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Option ID is required"})
		return
	}

	results, err := ctrl.pollService.SubmitVote(c.Request.Context(), shareCode, optionID)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Poll not found"})
			return
		}
		if strings.Contains(errMsg, "closed") {
			c.JSON(http.StatusForbidden, gin.H{"error": "This poll is closed and no longer accepting votes"})
			return
		}
		if strings.Contains(errMsg, "invalid option") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid option selected"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to submit vote. Please try again."})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Vote submitted successfully!",
		"results": results,
	})
}
