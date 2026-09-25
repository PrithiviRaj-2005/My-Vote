package routes

import (
	"net/http"
	"os"

	"pulsevote/controllers"
	"pulsevote/middleware"
	ws "pulsevote/websocket"

	"github.com/gin-gonic/gin"
)

// SetupRouter initializes Gin engine with CORS, middleware, and API routes
func SetupRouter(
	authCtrl *controllers.AuthController,
	pollCtrl *controllers.PollController,
	voteCtrl *controllers.VoteController,
	hub *ws.Hub,
) *gin.Engine {
	r := gin.Default()

	// Configure CORS
	allowedOriginEnv := os.Getenv("CORS_ORIGIN")
	if allowedOriginEnv == "" {
		allowedOriginEnv = "*"
	}

	r.Use(func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Allow requests from browser origin (including any Vercel deployment and localhost)
		if origin != "" {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		} else if allowedOriginEnv != "" && allowedOriginEnv != "*" {
			c.Writer.Header().Set("Access-Control-Allow-Origin", allowedOriginEnv)
		} else {
			c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		}

		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, Accept")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH, HEAD")
		c.Writer.Header().Set("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	})

	// Public Health Check
	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "My Vote Live Polling API",
		})
	})

	// WebSocket Realtime Route
	r.GET("/ws/polls/:shareCode", func(c *gin.Context) {
		ws.ServeWs(hub, c)
	})

	api := r.Group("/api")
	{
		// Authentication routes
		auth := api.Group("/auth")
		{
			auth.POST("/signup", authCtrl.Signup)
			auth.POST("/login", authCtrl.Login)
		}

		// Public Poll routes
		pollsPublic := api.Group("/polls")
		{
			pollsPublic.GET("/:shareCode", pollCtrl.GetPublicPoll)
			pollsPublic.POST("/:shareCode/vote", voteCtrl.SubmitVote)
			pollsPublic.GET("/:shareCode/results", pollCtrl.GetPollResults)
		}

		// Protected Poll routes (requires JWT)
		// Note: Use consistent :shareCode wildcard to prevent Gin radix tree conflicts
		pollsProtected := api.Group("/polls")
		pollsProtected.Use(middleware.AuthMiddleware())
		{
			pollsProtected.POST("", pollCtrl.CreatePoll)
			pollsProtected.GET("", pollCtrl.GetUserPolls)
			pollsProtected.PUT("/:shareCode", pollCtrl.UpdatePoll)
			pollsProtected.DELETE("/:shareCode", pollCtrl.DeletePoll)
			pollsProtected.POST("/:shareCode/close", pollCtrl.ClosePoll)
		}
	}

	return r
}
