package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"soj_core/app"
	"soj_core/config"
)

func main() {
	// Initialize database connection
	pool, err := config.InitDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer config.CloseDB()
	_ = pool

	// Create Gin router with default middleware (logger and recovery)
	router := gin.Default()
	router.Use(Cors())

	// Public routes (no authentication required)
	public := router.Group("/")
	{
		public.GET("/prelogin/:username", app.GetPreLogin)
		public.POST("/login", app.PostLoginV2)
		public.POST("/register", app.PostRegister)
	}

	// Token refresh route (requires refresh token)
	refreshGroup := router.Group("/")
	refreshGroup.Use(app.OptionalAuth)
	{
		refreshGroup.GET("/refresh", app.GetFresh)
	}

	// Protected routes (requires valid access token)
	protected := router.Group("/")
	protected.Use(app.RequireAuth)
	{
		protected.GET("/logout", app.GetLogout)
	}

	// Contest routes
	contestGroup := router.Group("/contests")
	contestOptionalGroup := router.Group("/contests")
	contestProtectedGroup := router.Group("/contests")
	contestOptionalGroup.Use(app.OptionalAuth)
	contestProtectedGroup.Use(app.RequireAuth)
	{
		contestGroup.GET("", app.GetContests)                   // GET /contests - List all contests
		contestGroup.POST("", app.CreateContest)                // POST /contests - Create a contest
		contestGroup.GET("/:id", app.GetContest)                // GET /contests/:id - Get a contest
		contestProtectedGroup.PUT("/:id", app.UpdateContest)    // PUT /contests/:id - Update a contest
		contestProtectedGroup.DELETE("/:id", app.DeleteContest) // DELETE /contests/:id - Delete a contest

		// Contest participants
		contestGroup.GET("/:id/participants", app.GetContestParticipants)                    // GET /contests/:id/participants - List participants
		contestProtectedGroup.POST("/:id/join", app.JoinContest)                             // POST /contests/:id/join - Join contest
		contestProtectedGroup.DELETE("/:id/leave", app.LeaveContest)                         // DELETE /contests/:id/leave - Leave contest
		contestProtectedGroup.DELETE("/:id/participants/:uid", app.RemoveContestParticipant) // DELETE /contests/:id/participants/:uid - Remove participant (admin)

		// Contest problems
		contestOptionalGroup.GET("/:id/problems", app.GetContestProblems)                   // GET /contests/:id/problems - List problems in contest
		contestProtectedGroup.POST("/:id/problems", app.AddContestProblem)                  // POST /contests/:id/problems - Add problem to contest
		contestProtectedGroup.DELETE("/:id/problems/:problem_id", app.RemoveContestProblem) // DELETE /contests/:id/problems/:problem_id - Remove problem from contest
	}

	// Problem routes
	//problemGroup := router.Group("/problems")
	problemOptionalGroup := router.Group("/problems")
	problemProtectedGroup := router.Group("/problems")
	problemOptionalGroup.Use(app.OptionalAuth)
	problemProtectedGroup.Use(app.RequireAuth)
	{
		problemOptionalGroup.GET("", app.GetProblems)           // GET /problems - List all problems
		problemProtectedGroup.POST("", app.CreateProblem)       // POST /problems - Create a problem
		problemOptionalGroup.GET("/:id", app.GetProblem)        // GET /problems/:id - Get a problem
		problemProtectedGroup.PUT("/:id", app.UpdateProblem)    // PUT /problems/:id - Update a problem
		problemProtectedGroup.DELETE("/:id", app.DeleteProblem) // DELETE /problems/:id - Delete a problem
	}

	// Submission routes
	submissionGroup := router.Group("/submissions")
	submissionOptionalGroup := router.Group("/submissions")
	submissionProtectedGroup := router.Group("/submissions")
	submissionOptionalGroup.Use(app.OptionalAuth)
	submissionProtectedGroup.Use(app.RequireAuth)
	{
		submissionGroup.GET("", app.GetSubmissions)                             // GET /submissions - List submissions (with filters)
		submissionProtectedGroup.POST("", app.CreateSubmission)                 // POST /submissions - Create a submission
		submissionGroup.GET("/:id", app.GetSubmission)                          // GET /submissions/:id - Get a submission
		submissionProtectedGroup.PUT("/:id/status", app.UpdateSubmissionStatus) // PUT /submissions/:id/status - Update submission status only
		submissionProtectedGroup.DELETE("/:id", app.DeleteSubmission)           // DELETE /submissions/:id - Delete a submission
	}

	log.Println("Server starting on :8080")
	if err := router.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// Cors returns a middleware that handles CORS headers
func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		if origin != "" {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, UPDATE")
			c.Header("Access-Control-Allow-Headers", "Authorization, Content-Length, X-CSRF-Token, Token, session, Content-Type")
			c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers")
			c.Header("Access-Control-Max-Age", "172800")
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		if c.Request.Method == "OPTIONS" {
			c.Status(http.StatusOK)
			return
		}

		c.Next()
	}
}
