// Package resp provides standardized response formats and error codes for the API.
package resp

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Error codes definition
// Format: HTTP status code (1 digit) + category (2 digits) + sequence (1 digit)
const (
	// 400x - Client errors (Bad Request)
	CodeInvalidForm     = 4001 // Invalid form data
	CodeInvalidFormat   = 4002 // Invalid format (password, salt, username)
	CodeAlreadyExists   = 4003 // Username or email already exists
	CodeValidationError = 4004 // Validation error

	// 401x - Authentication errors (Unauthorized)
	CodeUnauthorized      = 4010 // General unauthorized
	CodeInvalidToken      = 4011 // Invalid or expired token
	CodeInvalidTokenType  = 4012 // Wrong token type (e.g., refresh instead of access)
	CodeSessionInvalid    = 4013 // Session invalid or expired
	CodeWrongPassword     = 4014 // Wrong password
	CodeMissingAuthHeader = 4015 // Missing Authorization header
	CodeTokenExpired       = 4016 // Token expired
	CodeInvalidAuthHeader  = 4017 // Invalid Authorization header format

	// 404x - Not Found
	CodeUserNotFound       = 4041 // User not found
	CodeContestNotFound    = 4042 // Contest not found
	CodeProblemNotFound    = 4043 // Problem not found
	CodeSubmissionNotFound = 4044 // Submission not found

	// 500x - Server errors (Internal Server Error)
	CodeInternalError     = 5000 // General internal error
	CodeGenerateSaltError = 5001 // Failed to generate salt
	CodeTokenGenError     = 5002 // Failed to generate token
	CodeDatabaseError     = 5003 // Database operation failed
)

// Error messages mapping
var errorMessages = map[int]string{
	CodeInvalidForm:       "Invalid form data",
	CodeInvalidFormat:     "Invalid format",
	CodeAlreadyExists:     "Resource already exists",
	CodeValidationError:   "Validation error",
	CodeUnauthorized:      "Unauthorized",
	CodeInvalidToken:      "Invalid or expired token",
	CodeInvalidTokenType:  "Invalid token type",
	CodeSessionInvalid:    "Session invalid or expired",
	CodeWrongPassword:     "Wrong password",
	CodeMissingAuthHeader: "Missing Authorization header",
	CodeUserNotFound:       "User not found",
	CodeContestNotFound:    "Contest not found",
	CodeProblemNotFound:    "Problem not found",
	CodeSubmissionNotFound: "Submission not found",
	CodeInternalError:     "Internal server error",
	CodeGenerateSaltError: "Failed to generate salt",
	CodeTokenGenError:     "Failed to generate token",
	CodeDatabaseError:     "Database operation failed",
	CodeInvalidAuthHeader:  "Invalid Authorization header format",
	CodeTokenExpired:       "Token expired, refresh your token",
}

// Response represents a standard API response
type Response struct {
	Code    int         `json:"code"`            // Application error code
	Message string      `json:"message"`         // Human-readable message
	Data    interface{} `json:"data,omitempty"`  // Response data (omitted if empty)
	Error   string      `json:"error,omitempty"` // Error details (omitted if empty)
}

// Success returns a successful response
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "Success",
		Data:    data,
	})
}

// Error returns an error response
func Error(c *gin.Context, httpStatus int, code int, errDetail string) {
	msg := errorMessages[code]
	if msg == "" {
		msg = "Unknown error"
	}

	resp := Response{
		Code:    code,
		Message: msg,
	}

	if errDetail != "" {
		resp.Error = errDetail
	}

	c.JSON(httpStatus, resp)
}

// ErrorWithMessage returns an error response with custom message
func ErrorWithMessage(c *gin.Context, httpStatus int, code int, message string, errDetail string) {
	resp := Response{
		Code:    code,
		Message: message,
	}

	if errDetail != "" {
		resp.Error = errDetail
	}

	c.JSON(httpStatus, resp)
}

// Common error response helpers

// BadRequest returns a 400 Bad Request response
func BadRequest(c *gin.Context, code int, errDetail string) {
	Error(c, http.StatusBadRequest, code, errDetail)
}

// Unauthorized returns a 401 Unauthorized response
func Unauthorized(c *gin.Context, code int, errDetail string) {
	Error(c, http.StatusUnauthorized, code, errDetail)
}

// NotFound returns a 404 Not Found response
func NotFound(c *gin.Context, code int, errDetail string) {
	Error(c, http.StatusNotFound, code, errDetail)
}

// InternalError returns a 500 Internal Server Error response
func InternalError(c *gin.Context, code int, errDetail string) {
	Error(c, http.StatusInternalServerError, code, errDetail)
}
