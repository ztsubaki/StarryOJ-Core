package app

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"

	"soj_core/config"
	"soj_core/utils/jwt"
	"soj_core/utils/resp"
)

// GetFresh handles token refresh
// Requires a valid refresh token in the Authorization header
// Returns a new token pair (access_token + refresh_token)
// GET /refresh
func GetFresh(c *gin.Context) {
	// Verify token type is refresh
	if c.GetString("token_type") != "refresh" {
		resp.Unauthorized(c, resp.CodeInvalidTokenType, "Refresh token required")
		return
	}

	tokenJTI := c.GetString("token_jti")
	sessionID := c.GetUint64("session_id")

	// Verify session exists and JTI matches
	var storedJTI string
	err := config.DB.QueryRow(c.Request.Context(),
		"SELECT jti FROM sessions WHERE session_id = $1", sessionID).Scan(&storedJTI)
	if err != nil || storedJTI != tokenJTI {
		resp.Unauthorized(c, resp.CodeSessionInvalid, "Session expired or revoked")
		return
	}

	// Generate new JTI and update session
	newJTI := jwt.GenerateJTI()
	_, err = config.DB.Exec(context.TODO(),
		"UPDATE sessions SET jti = $1, last_active = $2 WHERE session_id = $3",
		newJTI, time.Now(), sessionID)
	if err != nil {
		resp.InternalError(c, resp.CodeDatabaseError, "Failed to update session: "+err.Error())
		return
	}

	// Generate new token pair
	refreshToken, accessToken, err := jwt.GenerateTokenPair(c.GetUint64("uid"), sessionID, newJTI)
	if err != nil {
		resp.InternalError(c, resp.CodeTokenGenError, err.Error())
		return
	}

	resp.Success(c, gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"expires_in":    int64(config.JWTAccessTokenDuration.Seconds()),
		"token_type":    "Bearer",
	})
}
