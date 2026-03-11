package app

import (
	"strings"

	"github.com/gin-gonic/gin"

	"soj_core/utils/jwt"
	"soj_core/utils/resp"
)

// AuthenV2 is the JWT authentication middleware
// It validates the access token and sets user context
// Required header: Authorization: Bearer <token>
func authen(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.Set("authorized", false)
		c.Set("auth_error", resp.CodeMissingAuthHeader)
		return
	}

	authFields := strings.Fields(authHeader)
	if len(authFields) != 2 || strings.ToLower(authFields[0]) != "bearer" {
		c.Set("authorized", false)
		c.Set("auth_error", resp.CodeInvalidAuthHeader)
		return
	}

	tokenString := authFields[1]
	claims, err := jwt.ParseJwtToken(tokenString)
	if err != nil {
		c.Set("authorized", false)
		if strings.HasPrefix(err.Error(), "token expired by") {
			c.Set("auth_error", resp.CodeTokenExpired)
		} else {
			c.Set("auth_error", resp.CodeInvalidToken)
		}
		return
	}

	// Only access tokens can be used for authorization
	if claims.Type != "access" {
		c.Set("authorized", false)
		c.Set("auth_error", resp.CodeInvalidTokenType)
	} else {
		c.Set("authorized", true)
	}

	// Set context values for downstream handlers
	c.Set("uid", claims.UID)
	c.Set("token_type", claims.Type)
	c.Set("token_jti", claims.JTI)
	c.Set("session_id", claims.SessionID)
}

// RequireAuth middleware rejects requests without valid authentication
func RequireAuth(c *gin.Context) {
	authen(c)
	if !c.GetBool("authorized") {
		resp.Unauthorized(c, resp.CodeUnauthorized, c.GetString("auth_error"))
		c.Abort()
		return
	}
	c.Next()
}

func OptionalAuth(c *gin.Context) {
	// No need to do anything, just pass through
	authen(c)
	c.Next()
}
