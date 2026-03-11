package app

import (
	"context"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"soj_core/config"
	"soj_core/utils/resp"
)

// Problem represents a problem in the system
type Problem struct {
	ProblemID    uint64    `json:"problem_id" db:"problem_id"`
	Title        string    `json:"title" db:"title"`
	Description  string    `json:"description" db:"description"`
	InputFormat  string    `json:"input_format" db:"input_format"`
	OutputFormat string    `json:"output_format" db:"output_format"`
	SampleInput  string    `json:"sample_input" db:"sample_input"`
	SampleOutput string    `json:"sample_output" db:"sample_output"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

// CreateProblemForm represents the form for creating a problem
type CreateProblemForm struct {
	Title        string `json:"title" binding:"required,max=100"`
	Description  string `json:"description" binding:"required"`
	InputFormat  string `json:"input_format" binding:"required"`
	OutputFormat string `json:"output_format" binding:"required"`
	SampleInput  string `json:"sample_input" binding:"required"`
	SampleOutput string `json:"sample_output" binding:"required"`
}

// UpdateProblemForm represents the form for updating a problem
type UpdateProblemForm struct {
	Title        string `json:"title" binding:"omitempty,max=100"`
	Description  string `json:"description" binding:"omitempty"`
	InputFormat  string `json:"input_format" binding:"omitempty"`
	OutputFormat string `json:"output_format" binding:"omitempty"`
	SampleInput  string `json:"sample_input" binding:"omitempty"`
	SampleOutput string `json:"sample_output" binding:"omitempty"`
}

// GetProblems returns a list of all problems
// GET /problems
func GetProblems(c *gin.Context) {
	rows, err := config.DB.Query(context.TODO(),
		"SELECT problem_id, title, description, input_format, output_format, sample_input, sample_output, created_at FROM problems ORDER BY created_at DESC")
	if err != nil {
		resp.InternalError(c, resp.CodeDatabaseError, err.Error())
		return
	}
	defer rows.Close()

	var problems []Problem
	for rows.Next() {
		var problem Problem
		err := rows.Scan(&problem.ProblemID, &problem.Title, &problem.Description,
			&problem.InputFormat, &problem.OutputFormat, &problem.SampleInput,
			&problem.SampleOutput, &problem.CreatedAt)
		if err != nil {
			resp.InternalError(c, resp.CodeDatabaseError, err.Error())
			return
		}
		problems = append(problems, problem)
	}

	resp.Success(c, gin.H{
		"problems": problems,
		"total":    len(problems),
	})
}

// GetProblem returns a single problem by ID
// GET /problems/:id
func GetProblem(c *gin.Context) {
	idStr := c.Param("id")
	problemID, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		resp.BadRequest(c, resp.CodeInvalidFormat, "Invalid problem ID")
		return
	}

	var problem Problem
	err = config.DB.QueryRow(context.TODO(),
		"SELECT problem_id, title, description, input_format, output_format, sample_input, sample_output, created_at FROM problems WHERE problem_id = $1",
		problemID).Scan(&problem.ProblemID, &problem.Title, &problem.Description,
		&problem.InputFormat, &problem.OutputFormat, &problem.SampleInput,
		&problem.SampleOutput, &problem.CreatedAt)
	if err != nil {
		resp.NotFound(c, resp.CodeProblemNotFound, "Problem not found")
		return
	}

	resp.Success(c, problem)
}

// CreateProblem creates a new problem
// POST /problems
func CreateProblem(c *gin.Context) {
	var form CreateProblemForm
	if err := c.ShouldBindJSON(&form); err != nil {
		resp.BadRequest(c, resp.CodeInvalidForm, err.Error())
		return
	}

	// Insert problem
	var problemID uint64
	err := config.DB.QueryRow(context.TODO(),
		"INSERT INTO problems (title, description, input_format, output_format, sample_input, sample_output) VALUES ($1, $2, $3, $4, $5, $6) RETURNING problem_id",
		form.Title, form.Description, form.InputFormat, form.OutputFormat, form.SampleInput, form.SampleOutput).Scan(&problemID)
	if err != nil {
		resp.InternalError(c, resp.CodeDatabaseError, err.Error())
		return
	}

	resp.Success(c, gin.H{
		"problem_id": problemID,
		"message":    "Problem created successfully",
	})
}

// UpdateProblem updates an existing problem
// PUT /problems/:id
func UpdateProblem(c *gin.Context) {
	idStr := c.Param("id")
	problemID, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		resp.BadRequest(c, resp.CodeInvalidFormat, "Invalid problem ID")
		return
	}

	var form UpdateProblemForm
	if err := c.ShouldBindJSON(&form); err != nil {
		resp.BadRequest(c, resp.CodeInvalidForm, err.Error())
		return
	}

	// Check if problem exists
	var exists bool
	err = config.DB.QueryRow(context.TODO(),
		"SELECT EXISTS(SELECT 1 FROM problems WHERE problem_id = $1)", problemID).Scan(&exists)
	if err != nil {
		resp.InternalError(c, resp.CodeDatabaseError, err.Error())
		return
	}
	if !exists {
		resp.NotFound(c, resp.CodeProblemNotFound, "Problem not found")
		return
	}

	// Build dynamic update query
	var updates []interface{}
	var queryParts []string
	paramIdx := 1

	if form.Title != "" {
		queryParts = append(queryParts, "title = $"+strconv.Itoa(paramIdx))
		updates = append(updates, form.Title)
		paramIdx++
	}
	if form.Description != "" {
		queryParts = append(queryParts, "description = $"+strconv.Itoa(paramIdx))
		updates = append(updates, form.Description)
		paramIdx++
	}
	if form.InputFormat != "" {
		queryParts = append(queryParts, "input_format = $"+strconv.Itoa(paramIdx))
		updates = append(updates, form.InputFormat)
		paramIdx++
	}
	if form.OutputFormat != "" {
		queryParts = append(queryParts, "output_format = $"+strconv.Itoa(paramIdx))
		updates = append(updates, form.OutputFormat)
		paramIdx++
	}
	if form.SampleInput != "" {
		queryParts = append(queryParts, "sample_input = $"+strconv.Itoa(paramIdx))
		updates = append(updates, form.SampleInput)
		paramIdx++
	}
	if form.SampleOutput != "" {
		queryParts = append(queryParts, "sample_output = $"+strconv.Itoa(paramIdx))
		updates = append(updates, form.SampleOutput)
		paramIdx++
	}

	if len(queryParts) == 0 {
		resp.BadRequest(c, resp.CodeInvalidForm, "No fields to update")
		return
	}

	// Execute update
	query := "UPDATE problems SET " + joinStrings(queryParts, ", ") + " WHERE problem_id = $" + strconv.Itoa(paramIdx)
	updates = append(updates, problemID)

	_, err = config.DB.Exec(context.TODO(), query, updates...)
	if err != nil {
		resp.InternalError(c, resp.CodeDatabaseError, err.Error())
		return
	}

	resp.Success(c, gin.H{
		"problem_id": problemID,
		"message":    "Problem updated successfully",
	})
}

// DeleteProblem deletes a problem
// DELETE /problems/:id
func DeleteProblem(c *gin.Context) {
	idStr := c.Param("id")
	problemID, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		resp.BadRequest(c, resp.CodeInvalidFormat, "Invalid problem ID")
		return
	}

	// Check if problem exists
	var exists bool
	err = config.DB.QueryRow(context.TODO(),
		"SELECT EXISTS(SELECT 1 FROM problems WHERE problem_id = $1)", problemID).Scan(&exists)
	if err != nil {
		resp.InternalError(c, resp.CodeDatabaseError, err.Error())
		return
	}
	if !exists {
		resp.NotFound(c, resp.CodeProblemNotFound, "Problem not found")
		return
	}

	// Delete problem (cascade will handle related records)
	_, err = config.DB.Exec(context.TODO(),
		"DELETE FROM problems WHERE problem_id = $1", problemID)
	if err != nil {
		resp.InternalError(c, resp.CodeDatabaseError, err.Error())
		return
	}

	resp.Success(c, gin.H{
		"problem_id": problemID,
		"message":    "Problem deleted successfully",
	})
}
