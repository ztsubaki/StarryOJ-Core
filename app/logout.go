package app

import (
	"github.com/gin-gonic/gin"

	"soj_core/config"
	"soj_core/utils/resp"
)

// GetLogout handles user logout
// Deletes the current session, invalidating all tokens
// Requires a valid access token
// GET /logout
func GetLogout(c *gin.Context) {
	// Check if authenticated
	if !c.GetBool("authorized") {
		resp.Unauthorized(c, resp.CodeUnauthorized, c.GetString("auth_error"))
		return
	}

	sessionID := c.GetUint64("session_id")

	// Delete session from database
	_, err := config.DB.Exec(c.Request.Context(),
		"DELETE FROM sessions WHERE session_id = $1", sessionID)
	if err != nil {
		resp.InternalError(c, resp.CodeDatabaseError, "Failed to delete session: "+err.Error())
		return
	}

	resp.Success(c, gin.H{
		"message": "Logout successful",
	})
}
