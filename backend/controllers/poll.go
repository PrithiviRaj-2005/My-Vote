package controllers

import (
	"net/http"
	"strings"

	"pulsevote/models"
	"pulsevote/services"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// PollController handles all poll management and reading HTTP endpoints
type PollController struct {
	pollService *services.PollService
}

// NewPollController creates a new PollController
func NewPollController(pollService *services.PollService) *PollController {
	return &PollController{
		pollService: pollService,
	}
}

// getUserID extracts the authenticated user's ObjectID from the Gin context
func getUserID(c *gin.Context) (primitive.ObjectID, bool) {
	val, exists := c.Get("userID")
	if !exists {
		return primitive.NilObjectID, false
	}
	idStr, ok := val.(string)
	if !ok {
		return primitive.NilObjectID, false
	}
	objID, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		return primitive.NilObjectID, false
	}
	return objID, true
}

// CreatePoll handles POST /api/polls
func (ctrl *PollController) CreatePoll(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req models.CreatePollRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	poll, err := ctrl.pollService.CreatePoll(c.Request.Context(), userID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, poll)
}

// GetUserPolls handles GET /api/polls
func (ctrl *PollController) GetUserPolls(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	polls, err := ctrl.pollService.GetUserPolls(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user polls"})
		return
	}

	c.JSON(http.StatusOK, polls)
}

// GetPublicPoll handles GET /api/polls/:shareCode
func (ctrl *PollController) GetPublicPoll(c *gin.Context) {
	shareCode := strings.TrimSpace(c.Param("shareCode"))
	if shareCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Share code is required"})
		return
	}

	poll, err := ctrl.pollService.GetPollByShareCode(c.Request.Context(), shareCode)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Poll not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch poll"})
		return
	}

	// Fetch current live results along with public poll metadata
	results, err := ctrl.pollService.GetPollResults(c.Request.Context(), shareCode)
	if err != nil {
		// Return basic poll if results fail
		c.JSON(http.StatusOK, gin.H{
			"poll": poll,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"poll":    poll,
		"results": results,
	})
}

// GetPollResults handles GET /api/polls/:shareCode/results
func (ctrl *PollController) GetPollResults(c *gin.Context) {
	shareCode := strings.TrimSpace(c.Param("shareCode"))
	if shareCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Share code is required"})
		return
	}

	results, err := ctrl.pollService.GetPollResults(c.Request.Context(), shareCode)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Poll not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch results"})
		return
	}

	c.JSON(http.StatusOK, results)
}

// UpdatePoll handles PUT /api/polls/:id
func (ctrl *PollController) UpdatePoll(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	idParam := c.Param("id")
	pollID, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid poll ID format"})
		return
	}

	var req models.UpdatePollRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	updatedPoll, err := ctrl.pollService.UpdatePoll(c.Request.Context(), pollID, userID, req)
	if err != nil {
		if strings.Contains(err.Error(), "unauthorized") || strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized to update this poll"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updatedPoll)
}

// ClosePoll handles POST /api/polls/:id/close
func (ctrl *PollController) ClosePoll(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	idParam := c.Param("id")
	pollID, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid poll ID format"})
		return
	}

	closedPoll, err := ctrl.pollService.ClosePoll(c.Request.Context(), pollID, userID)
	if err != nil {
		if strings.Contains(err.Error(), "unauthorized") || strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized to close this poll"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to close poll"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Poll closed successfully",
		"poll":    closedPoll,
	})
}

// DeletePoll handles DELETE /api/polls/:id
func (ctrl *PollController) DeletePoll(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	idParam := c.Param("id")
	pollID, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid poll ID format"})
		return
	}

	err = ctrl.pollService.DeletePoll(c.Request.Context(), pollID, userID)
	if err != nil {
		if strings.Contains(err.Error(), "unauthorized") || strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized to delete this poll"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete poll"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Poll deleted successfully"})
}
