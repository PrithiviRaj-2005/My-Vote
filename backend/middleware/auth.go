package middleware

import (
	"net/http"
	"strings"

	"pulsevote/utils"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware inspects the Authorization header for a valid Bearer JWT
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header is required",
			})
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid Authorization header format. Expected 'Bearer <token>'",
			})
			c.Abort()
			return
		}

		tokenString := strings.TrimSpace(parts[1])
		claims, err := utils.ValidateJWT(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid or expired authorization token: " + err.Error(),
			})
			c.Abort()
			return
		}

		// Inject user identity into Gin context for downstream handlers
		c.Set("userID", claims.UserID)
		c.Set("userEmail", claims.Email)
		c.Set("userName", claims.Name)

		c.Next()
	}
}
