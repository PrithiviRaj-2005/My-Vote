package controllers

import (
	"net/http"
	"strings"

	"pulsevote/models"
	"pulsevote/services"

	"github.com/gin-gonic/gin"
)

// AuthController handles authentication HTTP endpoints
type AuthController struct {
	authService *services.AuthService
}

// NewAuthController creates a new AuthController
func NewAuthController(authService *services.AuthService) *AuthController {
	return &AuthController{
		authService: authService,
	}
}

// Signup handles user registration: POST /api/auth/signup
func (ctrl *AuthController) Signup(c *gin.Context) {
	var req models.SignupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON request format"})
		return
	}

	res, err := ctrl.authService.Signup(c.Request.Context(), req)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "already exists") {
			c.JSON(http.StatusConflict, gin.H{"error": errMsg})
			return
		}
		if strings.Contains(errMsg, "required") || strings.Contains(errMsg, "invalid email") || strings.Contains(errMsg, "characters long") || strings.Contains(errMsg, "do not match") {
			c.JSON(http.StatusBadRequest, gin.H{"error": errMsg})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to complete registration. Please try again later."})
		return
	}

	c.JSON(http.StatusCreated, res)
}

// Login handles user authentication: POST /api/auth/login
func (ctrl *AuthController) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON request format"})
		return
	}

	res, err := ctrl.authService.Login(c.Request.Context(), req)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "invalid email or password") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
			return
		}
		if strings.Contains(errMsg, "required") {
			c.JSON(http.StatusBadRequest, gin.H{"error": errMsg})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Authentication service error. Please try again later."})
		return
	}

	c.JSON(http.StatusOK, res)
}
