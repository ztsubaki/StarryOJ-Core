package app

import (
	"context"
	"crypto/rand"
	"math/big"

	"github.com/gin-gonic/gin"

	"soj_core/config"
	"soj_core/structs"
	"soj_core/utils/resp"
)

// GetPreLogin returns user salt and generates a one-time login salt
// Client must hash: SHA256(stored_password + loginSalt) and send as password
// GET /prelogin/:username
func GetPreLogin(c *gin.Context) {
	username := c.Param("username")

	var user structs.User
	err := config.DB.QueryRow(context.TODO(),
		"SELECT uid, salt FROM users WHERE username = $1",
		username).Scan(&user.Uid, &user.Salt)
	if err != nil {
		resp.NotFound(c, resp.CodeUserNotFound, err.Error())
		return
	}

	// Generate one-time login salt (16 characters)
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ@#$*_"
	saltBytes := make([]byte, 16)
	for i := 0; i < 16; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			resp.InternalError(c, resp.CodeGenerateSaltError, err.Error())
			return
		}
		saltBytes[i] = letters[n.Int64()]
	}
	loginSalt := string(saltBytes)

	// Store login salt in database (one-time use)
	_, err = config.DB.Exec(context.TODO(),
		"UPDATE users SET login_salt = $1 WHERE username = $2",
		loginSalt, username)
	if err != nil {
		resp.InternalError(c, resp.CodeDatabaseError, err.Error())
		return
	}

	resp.Success(c, gin.H{
		"uid":        user.Uid,
		"username":   username,
		"salt":       user.Salt,
		"login_salt": loginSalt,
	})
}
