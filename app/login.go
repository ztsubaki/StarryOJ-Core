package app

import (
	"context"
	"crypto/sha256"
	"fmt"

	"github.com/gin-gonic/gin"

	"soj_core/config"
	"soj_core/structs"
	"soj_core/utils/jwt"
	"soj_core/utils/resp"
)

// LoginV2Form represents the login request form
type LoginV2Form struct {
	Username string `json:"username" binding:"required,min=3,max=15"`
	Password string `json:"password" binding:"required,len=64,hexadecimal"`
}

// PostLoginV2 handles user login
// Client must send: SHA256(stored_password + loginSalt) as password
// POST /login
func PostLoginV2(c *gin.Context) {
	var form LoginV2Form
	if err := c.ShouldBindJSON(&form); err != nil {
		resp.BadRequest(c, resp.CodeInvalidForm, err.Error())
		return
	}

	// Query user with login_salt
	var loginUser structs.User
	err := config.DB.QueryRow(context.TODO(),
		"SELECT uid, username, password, salt, login_salt FROM users WHERE username = $1",
		form.Username).Scan(&loginUser.Uid, &loginUser.Username, &loginUser.Password, &loginUser.Salt, &loginUser.LoginSalt)
	if err != nil {
		resp.NotFound(c, resp.CodeUserNotFound, err.Error())
		return
	}

	// Clear login_salt immediately (one-time use)
	_, err = config.DB.Exec(context.TODO(),
		"UPDATE users SET login_salt = NULL WHERE username = $1",
		form.Username)
	if err != nil {
		resp.InternalError(c, resp.CodeDatabaseError, "Failed to clear login salt: "+err.Error())
		return
	}

	// Verify password: SHA256(stored_password + loginSalt)
	passwordInput := loginUser.Password + loginUser.LoginSalt
	hashedPassword := fmt.Sprintf("%x", sha256.Sum256([]byte(passwordInput)))
	if hashedPassword != form.Password {
		resp.Unauthorized(c, resp.CodeWrongPassword, "")
		return
	}

	// Create session
	jti := jwt.GenerateJTI()
	var sessionID uint64
	err = config.DB.QueryRow(context.TODO(),
		"INSERT INTO sessions (uid, jti) VALUES ($1, $2) RETURNING session_id",
		loginUser.Uid, jti).Scan(&sessionID)
	if err != nil {
		resp.InternalError(c, resp.CodeDatabaseError, "Failed to create session: "+err.Error())
		return
	}

	// Generate token pair
	refreshToken, accessToken, err := jwt.GenerateTokenPair(loginUser.Uid, sessionID, jti)
	if err != nil {
		resp.InternalError(c, resp.CodeTokenGenError, err.Error())
		return
	}

	resp.Success(c, gin.H{
		"uid":           loginUser.Uid,
		"username":      loginUser.Username,
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"expires_in":    int64(config.JWTAccessTokenDuration.Seconds()),
		"token_type":    "Bearer",
	})
}
