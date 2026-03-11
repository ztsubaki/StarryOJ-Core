package app

import (
	"context"
	"regexp"

	"github.com/gin-gonic/gin"

	"soj_core/config"
	"soj_core/utils/resp"
)

// registerForm represents the user registration form
type registerForm struct {
	Username string `json:"username" form:"username" binding:"required"`
	Password string `json:"password" form:"password" binding:"required"`
	Salt     string `json:"salt" form:"salt" binding:"required"`
	Email    string `json:"email" form:"email" binding:"omitempty,email"`
	Nickname string `json:"nickname" form:"nickname" binding:"omitempty"`
}

// PostRegister handles user registration
// POST /register
func PostRegister(ctx *gin.Context) {
	var form registerForm
	if err := ctx.ShouldBind(&form); err != nil {
		resp.BadRequest(ctx, resp.CodeInvalidForm, err.Error())
		return
	}

	// Validate password format: 64 hex characters (SHA-256)
	if !regexp.MustCompile("^[0-9a-f]{64}$").MatchString(form.Password) {
		resp.BadRequest(ctx, resp.CodeInvalidFormat, "Password must be 64 hexadecimal characters")
		return
	}

	// Validate salt format: 16 characters from allowed set
	if !regexp.MustCompile("^[0-9a-zA-Z@#$%&*_]{16}$").MatchString(form.Salt) {
		resp.BadRequest(ctx, resp.CodeInvalidFormat, "Salt must be 16 characters from [0-9a-zA-Z@#$%&*_]")
		return
	}

	// Validate username format: 3-15 lowercase letters, numbers, underscores
	if !regexp.MustCompile("^[a-z0-9_]{3,15}$").MatchString(form.Username) {
		resp.BadRequest(ctx, resp.CodeInvalidFormat, "Username must be 3-15 lowercase letters, numbers, or underscores")
		return
	}

	// Check if username or email already exists
	var count int
	var err error
	if form.Email == "" {
		err = config.DB.QueryRow(context.TODO(),
			"SELECT COUNT(*) FROM users WHERE username = $1",
			form.Username).Scan(&count)
	} else {
		err = config.DB.QueryRow(context.TODO(),
			"SELECT COUNT(*) FROM users WHERE username = $1 OR email = $2",
			form.Username, form.Email).Scan(&count)
	}
	if err != nil {
		resp.InternalError(ctx, resp.CodeDatabaseError, err.Error())
		return
	}
	if count > 0 {
		resp.BadRequest(ctx, resp.CodeAlreadyExists, "Username or email already registered")
		return
	}

	// Insert new user and get the generated UID
	var uid uint64
	err = config.DB.QueryRow(context.TODO(),
		"INSERT INTO users (username, password, salt, email, nickname) VALUES ($1, $2, $3, $4, $5) RETURNING uid",
		form.Username, form.Password, form.Salt, form.Email, form.Nickname).Scan(&uid)
	if err != nil {
		resp.InternalError(ctx, resp.CodeDatabaseError, err.Error())
		return
	}

	resp.Success(ctx, gin.H{
		"uid":      uid,
		"username": form.Username,
	})
}
