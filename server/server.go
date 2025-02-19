// package server contains basic setup functions to start up the web server
package server

import (
	"github.com/Projeto-USPY/uspy-backend/config"
	"github.com/Projeto-USPY/uspy-backend/db"
	"github.com/Projeto-USPY/uspy-backend/entity"
	"github.com/Projeto-USPY/uspy-backend/entity/validation"
	"github.com/Projeto-USPY/uspy-backend/server/controllers/account"
	"github.com/Projeto-USPY/uspy-backend/server/controllers/private"
	"github.com/Projeto-USPY/uspy-backend/server/controllers/public"
	"github.com/Projeto-USPY/uspy-backend/server/controllers/restricted"
	"github.com/Projeto-USPY/uspy-backend/server/middleware"
	"github.com/gin-gonic/gin"
)

func setupAccount(DB db.Env, accountGroup *gin.RouterGroup) {
	accountGroup.DELETE("", middleware.JWT(), account.Delete(DB))
	accountGroup.GET("/captcha", account.SignupCaptcha())
	accountGroup.GET("/logout", middleware.JWT(), account.Logout())
	accountGroup.GET("/profile", middleware.JWT(), account.Profile(DB))
	accountGroup.POST("/login", account.Login(DB))
	accountGroup.POST("/create", account.Signup(DB))
	accountGroup.PUT("/password_change", middleware.JWT(), account.ChangePassword(DB))
	accountGroup.PUT("/password_reset", account.ResetPassword(DB))
	accountGroup.GET("/verify", account.VerifyAccount(DB))

	emailGroup := accountGroup.Group("/email")
	{
		emailGroup.POST("/verification", account.VerifyEmail(DB))
		emailGroup.POST("/password_reset", account.RequestPasswordReset(DB))
	}
}

func setupPublic(DB db.Env, apiGroup *gin.RouterGroup) {
	apiGroup.GET("/subject/all", public.GetSubjects(DB))
	subjectAPI := apiGroup.Group("/subject", entity.SubjectBinder)
	{
		subjectAPI.GET("", public.GetSubjectByCode(DB))
		subjectAPI.GET("/relations", public.GetRelations(DB))
		subjectAPI.GET("/offerings", public.GetOfferings(DB))
	}
}

func setupRestricted(DB db.Env, restrictedGroup *gin.RouterGroup) {
	subjectAPI := restrictedGroup.Group("/subject", entity.SubjectBinder)
	{
		subjectAPI.GET("/grades", restricted.GetGrades(DB))
		subjectAPI.GET("/offerings", restricted.GetOfferingsWithStats(DB))

		offeringsAPI := subjectAPI.Group("/offerings", entity.OfferingBinder)
		{
			offeringsAPI.GET("/comments", restricted.GetOfferingComments(DB))
		}
	}
}

func setupPrivate(DB db.Env, privateGroup *gin.RouterGroup) {
	subjectAPI := privateGroup.Group("/subject", entity.SubjectBinder)
	{
		subjectAPI.GET("/grade", private.GetSubjectGrade(DB))
		subjectAPI.GET("/review", private.GetSubjectReview(DB))
		subjectAPI.POST("/review", private.UpdateSubjectReview(DB))

		offeringsAPI := subjectAPI.Group("/offerings", entity.OfferingBinder)
		{
			offeringsAPI.GET("/comments", private.GetComment(DB))
			offeringsAPI.PUT("/comments", private.PublishComment(DB))

			commentsAPI := offeringsAPI.Group("/comments", entity.CommentRatingBinder)
			{
				commentsAPI.GET("/rating", private.GetCommentRating(DB))
				commentsAPI.PUT("/rating", private.RateComment(DB))
				commentsAPI.PUT("/report", private.ReportComment(DB))
			}
		}
	}
}

func SetupRouter(DB db.Env) (*gin.Engine, error) {
	r := gin.Default() // Create web-server object

	err := validation.SetupValidators()
	if err != nil {
		return nil, err
	}

	r.Use(gin.Recovery(), middleware.DefineDomain(), middleware.DumpErrors())

	if config.Env.IsLocal() {
		r.Use(middleware.AllowAnyOrigin())
	} else {
		if limiter := middleware.RateLimiter(config.Env.RateLimit); limiter != nil {
			r.Use(limiter)
		}
		r.Use(middleware.AllowUSPYOrigin())
	}

	// Login, Logout, Sign-in and other account related operations
	setupAccount(DB, r.Group("/account"))

	// Public endpoints: available for all users, including guests
	setupPublic(DB, r.Group("/api"))

	// Restricted endpoints: available only for registered users
	setupRestricted(DB, r.Group("/api/restricted", middleware.JWT()))

	// Private endpoints: every endpoint related to operations that the user utilizes their own data
	setupPrivate(DB, r.Group("/private", middleware.JWT()))

	return r, nil
}
