// internal/middleware/attach-user.go
/*
Middleware to attach user to context
As we are not using auth, we are hardcoding the user id.
In production, we would use JWT or something similar to get the user id.
*/
package middleware

import (
	"github.com/gin-gonic/gin"
)

// AttachUserMiddleware attaches the user to the context
func AttachUserMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userId := "11111111-1111-1111-1111-111111111111"

		c.Set("user_id", userId)
		c.Next()
	}
}
