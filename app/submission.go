package app

import (
	"context"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"soj_core/config"
	"soj_core/utils/resp"
)

// Submission represents a submission in the system
type Submission struct {
	SubmissionID uint64    `json:"submission_id" db:"submission_id"`
	ProblemID    uint64    `json:"problem_id" db:"problem_id"`
	ContestID    *uint64   `json:"contest_id" db:"contest_id"`
	UID          uint64    `json:"uid" db:"uid"`
	Code         string    `json:"code" db:"code"`
	Language     string    `json:"language" db:"language"`
	Status       string    `json:"status" db:"status"`
	SubmittedAt  time.Time `json:"submitted_at" db:"submitted_at"`
}

// CreateSubmissionForm represents the form for creating a submission
type CreateSubmissionForm struct {
	ProblemID uint64 `json:"problem_id" binding:"required"`
	ContestID uint64 `json:"contest_id" binding:"omitempty"`
	Code      string `json:"code" binding:"required"`
	Language  string `json:"language" binding:"required,max=20"`
}

// UpdateSubmissionStatusForm represents the form for updating submission status
type UpdateSubmissionStatusForm struct {
	Status string `json:"status" binding:"required,max=20"`
}

// Valid submission statuses
var validSubmissionStatuses = []string{"Pending", "Running", "Accepted", "Wrong Answer", "Time Limit Exceeded", "Memory Limit Exceeded", "Runtime Error", "Compilation Error"}

// isValidStatus checks if the status is valid
func isValidStatus(status string) bool {
	for _, s := range validSubmissionStatuses {
		if s == status {
			return true
		}
	}
	return false
}

// GetSubmissions returns a list of submissions
// GET /submissions
// Query params: problem_id, contest_id, uid (optional filters)
func GetSubmissions(c *gin.Context) {
	// Build query based on filters
	query := "SELECT submission_id, problem_id, contest_id, uid, code, language, status, submitted_at FROM submissions WHERE 1=1"
	var args []interface{}
	argIdx := 1

	if problemID := c.Query("problem_id"); problemID != "" {
		pid, err := strconv.ParseUint(problemID, 10, 64)
		if err == nil {
			query += " AND problem_id = $" + strconv.Itoa(argIdx)
			args = append(args, pid)
			argIdx++
		}
	}

	if contestID := c.Query("contest_id"); contestID != "" {
		cid, err := strconv.ParseUint(contestID, 10, 64)
		if err == nil {
			query += " AND contest_id = $" + strconv.Itoa(argIdx)
			args = append(args, cid)
			argIdx++
		}
	}

	if uid := c.Query("uid"); uid != "" {
		userID, err := strconv.ParseUint(uid, 10, 64)
		if err == nil {
			query += " AND uid = $" + strconv.Itoa(argIdx)
			args = append(args, userID)
			argIdx++
		}
	}

	query += " ORDER BY submitted_at DESC"

	rows, err := config.DB.Query(context.TODO(), query, args...)
	if err != nil {
		resp.InternalError(c, resp.CodeDatabaseError, err.Error())
		return
	}
	defer rows.Close()

	var submissions []Submission
	for rows.Next() {
		var sub Submission
		err := rows.Scan(&sub.SubmissionID, &sub.ProblemID, &sub.ContestID, &sub.UID,
			&sub.Code, &sub.Language, &sub.Status, &sub.SubmittedAt)
		if err != nil {
			resp.InternalError(c, resp.CodeDatabaseError, err.Error())
			return
		}
		submissions = append(submissions, sub)
	}

	resp.Success(c, gin.H{
		"submissions": submissions,
		"total":       len(submissions),
	})
}

// GetSubmission returns a single submission by ID
// GET /submissions/:id
func GetSubmission(c *gin.Context) {
	idStr := c.Param("id")
	submissionID, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		resp.BadRequest(c, resp.CodeInvalidFormat, "Invalid submission ID")
		return
	}

	var sub Submission
	err = config.DB.QueryRow(context.TODO(),
		"SELECT submission_id, problem_id, contest_id, uid, code, language, status, submitted_at FROM submissions WHERE submission_id = $1",
		submissionID).Scan(&sub.SubmissionID, &sub.ProblemID, &sub.ContestID, &sub.UID,
		&sub.Code, &sub.Language, &sub.Status, &sub.SubmittedAt)
	if err != nil {
		resp.NotFound(c, resp.CodeSubmissionNotFound, "Submission not found")
		return
	}

	resp.Success(c, sub)
}

// CreateSubmission creates a new submission
// POST /submissions
func CreateSubmission(c *gin.Context) {
	var form CreateSubmissionForm
	if err := c.ShouldBindJSON(&form); err != nil {
		resp.BadRequest(c, resp.CodeInvalidForm, err.Error())
		return
	}

	// Get uid from context (set by AuthenV2 middleware)
	uid, exists := c.Get("uid")
	if !exists {
		resp.Unauthorized(c, resp.CodeUnauthorized, "User not authenticated")
		return
	}
	userID := uid.(uint64)

	// Verify problem exists
	var problemExists bool
	err := config.DB.QueryRow(context.TODO(),
		"SELECT EXISTS(SELECT 1 FROM problems WHERE problem_id = $1)", form.ProblemID).Scan(&problemExists)
	if err != nil {
		resp.InternalError(c, resp.CodeDatabaseError, err.Error())
		return
	}
	if !problemExists {
		resp.NotFound(c, resp.CodeProblemNotFound, "Problem not found")
		return
	}

	// If contest_id is provided, verify contest exists
	var contestID interface{}
	if form.ContestID > 0 {
		var contestExists bool
		err = config.DB.QueryRow(context.TODO(),
			"SELECT EXISTS(SELECT 1 FROM contests WHERE contest_id = $1)", form.ContestID).Scan(&contestExists)
		if err != nil {
			resp.InternalError(c, resp.CodeDatabaseError, err.Error())
			return
		}
		if !contestExists {
			resp.NotFound(c, resp.CodeContestNotFound, "Contest not found")
			return
		}
		contestID = form.ContestID
	} else {
		contestID = nil
	}

	// Insert submission with default status "Pending"
	var submissionID uint64
	err = config.DB.QueryRow(context.TODO(),
		"INSERT INTO submissions (problem_id, contest_id, uid, code, language, status) VALUES ($1, $2, $3, $4, $5, $6) RETURNING submission_id",
		form.ProblemID, contestID, userID, form.Code, form.Language, "Pending").Scan(&submissionID)
	if err != nil {
		resp.InternalError(c, resp.CodeDatabaseError, err.Error())
		return
	}

	resp.Success(c, gin.H{
		"submission_id": submissionID,
		"status":        "Pending",
		"message":       "Submission created successfully",
	})
}

// UpdateSubmissionStatus updates only the status of a submission
// PUT /submissions/:id/status
func UpdateSubmissionStatus(c *gin.Context) {
	idStr := c.Param("id")
	submissionID, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		resp.BadRequest(c, resp.CodeInvalidFormat, "Invalid submission ID")
		return
	}

	var form UpdateSubmissionStatusForm
	if err := c.ShouldBindJSON(&form); err != nil {
		resp.BadRequest(c, resp.CodeInvalidForm, err.Error())
		return
	}

	// Validate status
	if !isValidStatus(form.Status) {
		resp.BadRequest(c, resp.CodeValidationError, "Invalid status. Valid statuses: "+joinStrings(validSubmissionStatuses, ", "))
		return
	}

	// Check if submission exists
	var exists bool
	err = config.DB.QueryRow(context.TODO(),
		"SELECT EXISTS(SELECT 1 FROM submissions WHERE submission_id = $1)", submissionID).Scan(&exists)
	if err != nil {
		resp.InternalError(c, resp.CodeDatabaseError, err.Error())
		return
	}
	if !exists {
		resp.NotFound(c, resp.CodeSubmissionNotFound, "Submission not found")
		return
	}

	// Update only status
	_, err = config.DB.Exec(context.TODO(),
		"UPDATE submissions SET status = $1 WHERE submission_id = $2",
		form.Status, submissionID)
	if err != nil {
		resp.InternalError(c, resp.CodeDatabaseError, err.Error())
		return
	}

	resp.Success(c, gin.H{
		"submission_id": submissionID,
		"status":        form.Status,
		"message":       "Submission status updated successfully",
	})
}

// DeleteSubmission deletes a submission
// DELETE /submissions/:id
func DeleteSubmission(c *gin.Context) {
	idStr := c.Param("id")
	submissionID, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		resp.BadRequest(c, resp.CodeInvalidFormat, "Invalid submission ID")
		return
	}

	// Check if submission exists
	var exists bool
	err = config.DB.QueryRow(context.TODO(),
		"SELECT EXISTS(SELECT 1 FROM submissions WHERE submission_id = $1)", submissionID).Scan(&exists)
	if err != nil {
		resp.InternalError(c, resp.CodeDatabaseError, err.Error())
		return
	}
	if !exists {
		resp.NotFound(c, resp.CodeSubmissionNotFound, "Submission not found")
		return
	}

	// Delete submission
	_, err = config.DB.Exec(context.TODO(),
		"DELETE FROM submissions WHERE submission_id = $1", submissionID)
	if err != nil {
		resp.InternalError(c, resp.CodeDatabaseError, err.Error())
		return
	}

	resp.Success(c, gin.H{
		"submission_id": submissionID,
		"message":       "Submission deleted successfully",
	})
}
