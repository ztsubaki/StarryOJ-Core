package app

import (
	"context"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"soj_core/config"
	"soj_core/utils/resp"
)

// ContestParticipant represents a participant in a contest
type ContestParticipant struct {
	ContestID uint64    `json:"contest_id" db:"contest_id"`
	UID       uint64    `json:"uid" db:"uid"`
	JoinedAt  time.Time `json:"joined_at" db:"joined_at"`
}

// JoinContestForm represents the form for joining a contest
type JoinContestForm struct {
	ContestID uint64 `json:"contest_id" binding:"required"`
}

// GetContestParticipants returns all participants of a contest
// GET /contests/:id/participants
func GetContestParticipants(c *gin.Context) {
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

	rows, err := config.DB.Query(context.TODO(),
		"SELECT contest_id, uid, joined_at FROM contest_participants WHERE contest_id = $1 ORDER BY joined_at ASC",
		contestID)
	if err != nil {
		resp.InternalError(c, resp.CodeDatabaseError, err.Error())
		return
	}
	defer rows.Close()

	var participants []ContestParticipant
	for rows.Next() {
		var p ContestParticipant
		err := rows.Scan(&p.ContestID, &p.UID, &p.JoinedAt)
		if err != nil {
			resp.InternalError(c, resp.CodeDatabaseError, err.Error())
			return
		}
		participants = append(participants, p)
	}

	resp.Success(c, gin.H{
		"contest_id":   contestID,
		"participants": participants,
		"total":        len(participants),
	})
}

// JoinContest adds current user to a contest
// POST /contests/:id/join
func JoinContest(c *gin.Context) {
	idStr := c.Param("id")
	contestID, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		resp.BadRequest(c, resp.CodeInvalidFormat, "Invalid contest ID")
		return
	}

	// Get uid from context
	uid, exists := c.Get("uid")
	if !exists {
		resp.Unauthorized(c, resp.CodeUnauthorized, "User not authenticated")
		return
	}
	userID := uid.(uint64)

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

	// Check if already joined
	var alreadyJoined bool
	err = config.DB.QueryRow(context.TODO(),
		"SELECT EXISTS(SELECT 1 FROM contest_participants WHERE contest_id = $1 AND uid = $2)",
		contestID, userID).Scan(&alreadyJoined)
	if err != nil {
		resp.InternalError(c, resp.CodeDatabaseError, err.Error())
		return
	}
	if alreadyJoined {
		resp.BadRequest(c, resp.CodeAlreadyExists, "Already joined this contest")
		return
	}

	// Insert participant
	_, err = config.DB.Exec(context.TODO(),
		"INSERT INTO contest_participants (contest_id, uid) VALUES ($1, $2)",
		contestID, userID)
	if err != nil {
		resp.InternalError(c, resp.CodeDatabaseError, err.Error())
		return
	}

	resp.Success(c, gin.H{
		"contest_id": contestID,
		"uid":        userID,
		"message":    "Successfully joined contest",
	})
}

// LeaveContest removes current user from a contest
// DELETE /contests/:id/leave
func LeaveContest(c *gin.Context) {
	idStr := c.Param("id")
	contestID, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		resp.BadRequest(c, resp.CodeInvalidFormat, "Invalid contest ID")
		return
	}

	// Get uid from context
	uid, exists := c.Get("uid")
	if !exists {
		resp.Unauthorized(c, resp.CodeUnauthorized, "User not authenticated")
		return
	}
	userID := uid.(uint64)

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

	// Check if joined
	var joined bool
	err = config.DB.QueryRow(context.TODO(),
		"SELECT EXISTS(SELECT 1 FROM contest_participants WHERE contest_id = $1 AND uid = $2)",
		contestID, userID).Scan(&joined)
	if err != nil {
		resp.InternalError(c, resp.CodeDatabaseError, err.Error())
		return
	}
	if !joined {
		resp.BadRequest(c, resp.CodeValidationError, "Not joined this contest")
		return
	}

	// Delete participant
	_, err = config.DB.Exec(context.TODO(),
		"DELETE FROM contest_participants WHERE contest_id = $1 AND uid = $2",
		contestID, userID)
	if err != nil {
		resp.InternalError(c, resp.CodeDatabaseError, err.Error())
		return
	}

	resp.Success(c, gin.H{
		"contest_id": contestID,
		"uid":        userID,
		"message":    "Successfully left contest",
	})
}

// RemoveContestParticipant removes a participant from a contest (admin only)
// DELETE /contests/:id/participants/:uid
func RemoveContestParticipant(c *gin.Context) {
	idStr := c.Param("id")
	contestID, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		resp.BadRequest(c, resp.CodeInvalidFormat, "Invalid contest ID")
		return
	}

	uidStr := c.Param("uid")
	targetUID, err := strconv.ParseUint(uidStr, 10, 64)
	if err != nil {
		resp.BadRequest(c, resp.CodeInvalidFormat, "Invalid user ID")
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

	// Check if user is a participant
	var joined bool
	err = config.DB.QueryRow(context.TODO(),
		"SELECT EXISTS(SELECT 1 FROM contest_participants WHERE contest_id = $1 AND uid = $2)",
		contestID, targetUID).Scan(&joined)
	if err != nil {
		resp.InternalError(c, resp.CodeDatabaseError, err.Error())
		return
	}
	if !joined {
		resp.NotFound(c, resp.CodeUserNotFound, "User is not a participant of this contest")
		return
	}

	// Delete participant
	_, err = config.DB.Exec(context.TODO(),
		"DELETE FROM contest_participants WHERE contest_id = $1 AND uid = $2",
		contestID, targetUID)
	if err != nil {
		resp.InternalError(c, resp.CodeDatabaseError, err.Error())
		return
	}

	resp.Success(c, gin.H{
		"contest_id": contestID,
		"uid":        targetUID,
		"message":    "Participant removed successfully",
	})
}
