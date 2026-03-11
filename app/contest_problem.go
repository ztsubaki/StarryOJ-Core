package app

import (
	"context"
	"strconv"

	"github.com/gin-gonic/gin"

	"soj_core/config"
	"soj_core/utils/resp"
)

// ContestProblem represents a problem in a contest
type ContestProblem struct {
	ContestID uint64 `json:"contest_id" db:"contest_id"`
	ProblemID uint64 `json:"problem_id" db:"problem_id"`
}

// AddContestProblemForm represents the form for adding a problem to a contest
type AddContestProblemForm struct {
	ProblemID uint64 `json:"problem_id" binding:"required"`
}

// GetContestProblems returns all problems in a contest
// GET /contests/:id/problems
func GetContestProblems(c *gin.Context) {
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

	// Get problems with their details
	rows, err := config.DB.Query(context.TODO(),
		`SELECT p.problem_id, p.title, p.description, p.input_format, p.output_format,
		        p.sample_input, p.sample_output, p.created_at
		 FROM problems p
		 INNER JOIN contest_problems cp ON p.problem_id = cp.problem_id
		 WHERE cp.contest_id = $1
		 ORDER BY p.problem_id ASC`,
		contestID)
	if err != nil {
		resp.InternalError(c, resp.CodeDatabaseError, err.Error())
		return
	}
	defer rows.Close()

	var problems []Problem
	for rows.Next() {
		var p Problem
		err := rows.Scan(&p.ProblemID, &p.Title, &p.Description, &p.InputFormat,
			&p.OutputFormat, &p.SampleInput, &p.SampleOutput, &p.CreatedAt)
		if err != nil {
			resp.InternalError(c, resp.CodeDatabaseError, err.Error())
			return
		}
		problems = append(problems, p)
	}

	resp.Success(c, gin.H{
		"contest_id": contestID,
		"problems":   problems,
		"total":      len(problems),
	})
}

// AddContestProblem adds a problem to a contest
// POST /contests/:id/problems
func AddContestProblem(c *gin.Context) {
	idStr := c.Param("id")
	contestID, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		resp.BadRequest(c, resp.CodeInvalidFormat, "Invalid contest ID")
		return
	}

	var form AddContestProblemForm
	if err := c.ShouldBindJSON(&form); err != nil {
		resp.BadRequest(c, resp.CodeInvalidForm, err.Error())
		return
	}

	// Check if contest exists
	var existsContest bool
	err = config.DB.QueryRow(context.TODO(),
		"SELECT EXISTS(SELECT 1 FROM contests WHERE contest_id = $1)", contestID).Scan(&existsContest)
	if err != nil {
		resp.InternalError(c, resp.CodeDatabaseError, err.Error())
		return
	}
	if !existsContest {
		resp.NotFound(c, resp.CodeContestNotFound, "Contest not found")
		return
	}

	// Check if problem exists
	var existsProblem bool
	err = config.DB.QueryRow(context.TODO(),
		"SELECT EXISTS(SELECT 1 FROM problems WHERE problem_id = $1)", form.ProblemID).Scan(&existsProblem)
	if err != nil {
		resp.InternalError(c, resp.CodeDatabaseError, err.Error())
		return
	}
	if !existsProblem {
		resp.NotFound(c, resp.CodeProblemNotFound, "Problem not found")
		return
	}

	// Check if problem already in contest
	var alreadyAdded bool
	err = config.DB.QueryRow(context.TODO(),
		"SELECT EXISTS(SELECT 1 FROM contest_problems WHERE contest_id = $1 AND problem_id = $2)",
		contestID, form.ProblemID).Scan(&alreadyAdded)
	if err != nil {
		resp.InternalError(c, resp.CodeDatabaseError, err.Error())
		return
	}
	if alreadyAdded {
		resp.BadRequest(c, resp.CodeAlreadyExists, "Problem already in this contest")
		return
	}

	// Insert contest problem
	_, err = config.DB.Exec(context.TODO(),
		"INSERT INTO contest_problems (contest_id, problem_id) VALUES ($1, $2)",
		contestID, form.ProblemID)
	if err != nil {
		resp.InternalError(c, resp.CodeDatabaseError, err.Error())
		return
	}

	resp.Success(c, gin.H{
		"contest_id": contestID,
		"problem_id": form.ProblemID,
		"message":    "Problem added to contest successfully",
	})
}

// RemoveContestProblem removes a problem from a contest
// DELETE /contests/:id/problems/:problem_id
func RemoveContestProblem(c *gin.Context) {
	idStr := c.Param("id")
	contestID, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		resp.BadRequest(c, resp.CodeInvalidFormat, "Invalid contest ID")
		return
	}

	problemIDStr := c.Param("problem_id")
	problemID, err := strconv.ParseUint(problemIDStr, 10, 64)
	if err != nil {
		resp.BadRequest(c, resp.CodeInvalidFormat, "Invalid problem ID")
		return
	}

	// Check if contest exists
	var existsContest bool
	err = config.DB.QueryRow(context.TODO(),
		"SELECT EXISTS(SELECT 1 FROM contests WHERE contest_id = $1)", contestID).Scan(&existsContest)
	if err != nil {
		resp.InternalError(c, resp.CodeDatabaseError, err.Error())
		return
	}
	if !existsContest {
		resp.NotFound(c, resp.CodeContestNotFound, "Contest not found")
		return
	}

	// Check if problem exists in contest
	var existsInContest bool
	err = config.DB.QueryRow(context.TODO(),
		"SELECT EXISTS(SELECT 1 FROM contest_problems WHERE contest_id = $1 AND problem_id = $2)",
		contestID, problemID).Scan(&existsInContest)
	if err != nil {
		resp.InternalError(c, resp.CodeDatabaseError, err.Error())
		return
	}
	if !existsInContest {
		resp.NotFound(c, resp.CodeProblemNotFound, "Problem not found in this contest")
		return
	}

	// Delete contest problem
	_, err = config.DB.Exec(context.TODO(),
		"DELETE FROM contest_problems WHERE contest_id = $1 AND problem_id = $2",
		contestID, problemID)
	if err != nil {
		resp.InternalError(c, resp.CodeDatabaseError, err.Error())
		return
	}

	resp.Success(c, gin.H{
		"contest_id": contestID,
		"problem_id": problemID,
		"message":    "Problem removed from contest successfully",
	})
}
