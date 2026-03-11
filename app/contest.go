package app

import (
	"context"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"soj_core/config"
	"soj_core/utils/resp"
)

// Contest represents a contest in the system
type Contest struct {
	ContestID   uint64    `json:"contest_id" db:"contest_id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	StartTime   time.Time `json:"start_time" db:"start_time"`
	EndTime     time.Time `json:"end_time" db:"end_time"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// CreateContestForm represents the form for creating a contest
type CreateContestForm struct {
	Name        string `json:"name" binding:"required,max=100"`
	Description string `json:"description" binding:"omitempty"`
	StartTime   string `json:"start_time" binding:"required"` // RFC3339 format
	EndTime     string `json:"end_time" binding:"required"`   // RFC3339 format
}

// UpdateContestForm represents the form for updating a contest
type UpdateContestForm struct {
	Name        string `json:"name" binding:"omitempty,max=100"`
	Description string `json:"description" binding:"omitempty"`
	StartTime   string `json:"start_time" binding:"omitempty"` // RFC3339 format
	EndTime     string `json:"end_time" binding:"omitempty"`   // RFC3339 format
}

// parseTime parses RFC3339 formatted time string
func parseTime(timeStr string) (time.Time, error) {
	return time.Parse(time.RFC3339, timeStr)
}

// GetContests returns a list of all contests
// GET /contests
func GetContests(c *gin.Context) {
	rows, err := config.DB.Query(context.TODO(),
		"SELECT contest_id, name, description, start_time, end_time, created_at FROM contests ORDER BY start_time DESC")
	if err != nil {
		resp.InternalError(c, resp.CodeDatabaseError, err.Error())
		return
	}
	defer rows.Close()

	var contests []Contest
	for rows.Next() {
		var contest Contest
		err := rows.Scan(&contest.ContestID, &contest.Name, &contest.Description,
			&contest.StartTime, &contest.EndTime, &contest.CreatedAt)
		if err != nil {
			resp.InternalError(c, resp.CodeDatabaseError, err.Error())
			return
		}
		contests = append(contests, contest)
	}

	resp.Success(c, gin.H{
		"contests": contests,
		"total":    len(contests),
	})
}

// GetContest returns a single contest by ID
// GET /contests/:id
func GetContest(c *gin.Context) {
	idStr := c.Param("id")
	contestID, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		resp.BadRequest(c, resp.CodeInvalidFormat, "Invalid contest ID")
		return
	}

	var contest Contest
	err = config.DB.QueryRow(context.TODO(),
		"SELECT contest_id, name, description, start_time, end_time, created_at FROM contests WHERE contest_id = $1",
		contestID).Scan(&contest.ContestID, &contest.Name, &contest.Description,
		&contest.StartTime, &contest.EndTime, &contest.CreatedAt)
	if err != nil {
		resp.NotFound(c, resp.CodeContestNotFound, "Contest not found")
		return
	}

	resp.Success(c, contest)
}

// CreateContest creates a new contest
// POST /contests
func CreateContest(c *gin.Context) {
	var form CreateContestForm
	if err := c.ShouldBindJSON(&form); err != nil {
		resp.BadRequest(c, resp.CodeInvalidForm, err.Error())
		return
	}

	// Parse start time
	startTime, err := parseTime(form.StartTime)
	if err != nil {
		resp.BadRequest(c, resp.CodeInvalidFormat, "Invalid start_time format, expected RFC3339")
		return
	}

	// Parse end time
	endTime, err := parseTime(form.EndTime)
	if err != nil {
		resp.BadRequest(c, resp.CodeInvalidFormat, "Invalid end_time format, expected RFC3339")
		return
	}

	// Validate time range
	if endTime.Before(startTime) || endTime.Equal(startTime) {
		resp.BadRequest(c, resp.CodeValidationError, "End time must be after start time")
		return
	}

	// Insert contest
	var contestID uint64
	err = config.DB.QueryRow(context.TODO(),
		"INSERT INTO contests (name, description, start_time, end_time) VALUES ($1, $2, $3, $4) RETURNING contest_id",
		form.Name, form.Description, startTime, endTime).Scan(&contestID)
	if err != nil {
		resp.InternalError(c, resp.CodeDatabaseError, err.Error())
		return
	}

	resp.Success(c, gin.H{
		"contest_id": contestID,
		"message":    "Contest created successfully",
	})
}

// UpdateContest updates an existing contest
// PUT /contests/:id
func UpdateContest(c *gin.Context) {
	idStr := c.Param("id")
	contestID, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		resp.BadRequest(c, resp.CodeInvalidFormat, "Invalid contest ID")
		return
	}

	var form UpdateContestForm
	if err := c.ShouldBindJSON(&form); err != nil {
		resp.BadRequest(c, resp.CodeInvalidForm, err.Error())
		return
	}

	// Check if contest exists
	var exists bool
	err = config.DB.QueryRow(context.TODO(),
		"SELECT EXISTS(SELECT 1 FROM contests WHERE contest_id = $1)", contestID).Scan(&exists)
	if err != nil {
		resp.InternalError(c, resp.CodeDatabaseError, err.Error())
		return
	}
	if !exists {
		resp.NotFound(c, resp.CodeContestNotFound, "Contest not found")
		return
	}

	// Build dynamic update query
	var updates []interface{}
	var queryParts []string
	paramIdx := 1

	if form.Name != "" {
		queryParts = append(queryParts, "name = $"+strconv.Itoa(paramIdx))
		updates = append(updates, form.Name)
		paramIdx++
	}
	if form.Description != "" {
		queryParts = append(queryParts, "description = $"+strconv.Itoa(paramIdx))
		updates = append(updates, form.Description)
		paramIdx++
	}

	var startTime, endTime time.Time
	if form.StartTime != "" {
		startTime, err = parseTime(form.StartTime)
		if err != nil {
			resp.BadRequest(c, resp.CodeInvalidFormat, "Invalid start_time format, expected RFC3339")
			return
		}
		queryParts = append(queryParts, "start_time = $"+strconv.Itoa(paramIdx))
		updates = append(updates, startTime)
		paramIdx++
	}
	if form.EndTime != "" {
		endTime, err = parseTime(form.EndTime)
		if err != nil {
			resp.BadRequest(c, resp.CodeInvalidFormat, "Invalid end_time format, expected RFC3339")
			return
		}
		queryParts = append(queryParts, "end_time = $"+strconv.Itoa(paramIdx))
		updates = append(updates, endTime)
		paramIdx++
	}

	if len(queryParts) == 0 {
		resp.BadRequest(c, resp.CodeInvalidForm, "No fields to update")
		return
	}

	// Validate time range if both times are being updated
	if form.StartTime != "" && form.EndTime != "" {
		if endTime.Before(startTime) || endTime.Equal(startTime) {
			resp.BadRequest(c, resp.CodeValidationError, "End time must be after start time")
			return
		}
	}

	// Execute update
	query := "UPDATE contests SET " + joinStrings(queryParts, ", ") + " WHERE contest_id = $" + strconv.Itoa(paramIdx)
	updates = append(updates, contestID)

	_, err = config.DB.Exec(context.TODO(), query, updates...)
	if err != nil {
		resp.InternalError(c, resp.CodeDatabaseError, err.Error())
		return
	}

	resp.Success(c, gin.H{
		"contest_id": contestID,
		"message":    "Contest updated successfully",
	})
}

// DeleteContest deletes a contest
// DELETE /contests/:id
func DeleteContest(c *gin.Context) {
	idStr := c.Param("id")
	contestID, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		resp.BadRequest(c, resp.CodeInvalidFormat, "Invalid contest ID")
		return
	}

	// Check if contest exists
	var exists bool
	err = config.DB.QueryRow(context.TODO(),
		"SELECT EXISTS(SELECT 1 FROM contests WHERE contest_id = $1)", contestID).Scan(&exists)
	if err != nil {
		resp.InternalError(c, resp.CodeDatabaseError, err.Error())
		return
	}
	if !exists {
		resp.NotFound(c, resp.CodeContestNotFound, "Contest not found")
		return
	}

	// Delete contest (cascade will handle related records)
	_, err = config.DB.Exec(context.TODO(),
		"DELETE FROM contests WHERE contest_id = $1", contestID)
	if err != nil {
		resp.InternalError(c, resp.CodeDatabaseError, err.Error())
		return
	}

	resp.Success(c, gin.H{
		"contest_id": contestID,
		"message":    "Contest deleted successfully",
	})
}

// joinStrings joins a slice of strings with a separator
func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}
