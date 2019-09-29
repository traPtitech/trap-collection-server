package router

import (
	"net/http"

	"github.com/labstack/echo"

	"github.com/traPtitech/trap-collection-server/repository"
)

//UserAuthenticate traQ認証
func UserAuthenticate() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// トークンを持たないユーザはアクセスできない
			if repository.GetUserID(c) == "-" {
				return echo.NewHTTPError(http.StatusUnauthorized, "You are not logged in")
			}
			return next(c)
		}
	}
}

//AdminAuthenticate AdminかをtraQ認証
func AdminAuthenticate() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			b, err := repository.CheckAdmin(c)
			if err != nil {
				return err
			}
			// 管理者でないユーザはアクセスできない
			if !b {
				return echo.NewHTTPError(http.StatusUnauthorized, "You are not admin or logged in")
			}
			return next(c)
		}
	}
}

//SetRouting ルーティング
func SetRouting(e *echo.Echo) {
	check := e.Group("/check")
	{
		check.POST("/:version/other", CheckHandler)
	}

	game := e.Group("/game")
	{
		game.GET("/:id", DownloadHandler)
	}

	responses := e.Group("/responses")
	{
		responses.POST("", PostResponse)
	}

	questionnaires := e.Group("/questionnaires")
	{
		questionnaires.GET("/:id/questions", GetQuestions)
	}

	time := e.Group("/time")
	{
		time.POST("", PostTimeHandler)
	}

	seat := e.Group("/seat")
	{
		seat.POST("", PostSeatHandler)
		seat.GET("", GetSeatHandler)
	}

	trap := e.Group("/trap", UserAuthenticate())
	{
		trapGame := trap.Group("/game")
		{
			trapGame.POST("", PostGameHandler)
			trapGame.PUT("", PutGameHandler)
			trapGame.DELETE("/:id", DeleteGameHandler)
			trapGame.GET("", GetGameNameListHandler)
		}

		trapResponses := trap.Group("/responses")
		{
			trapResponses.GET("/:id", GetResponse)
		}

		trapUsers := trap.Group("/users")
		{
			trapUsersMe := trapUsers.Group("/me")
			{
				trapUsersMe.GET("", GetUsersMe)
			}
		}

		trapResults := trap.Group("/results")
		{
			trapResults.GET("/:questionnaireID", GetResponsesByID)
		}

		trapVersion := trap.Group("/version")
		{
			trapVersion.GET("/sale", GetVersionForSaleListHandler)
			trapVersion.GET("/fes", GetVersionNotForSaleListHandler)
			trapVersion.GET("/:id/game", GetGameListHandler)
			trapVersion.GET("/:id/nongame", GetNonGameListHandler)
			trapVersion.GET("/:id/questionnaire", GetQuestionnaireHandler)
		}
	}

	admin := e.Group("/admin", AdminAuthenticate())
	{
		adminQuestions := admin.Group("/questions")
		{
			adminQuestions.POST("", PostQuestion)
			adminQuestions.PATCH("/:id", EditQuestion)
			adminQuestions.DELETE("/:id", DeleteQuestion)
		}

		adminQuestionnaires := admin.Group("/questionnaires")
		{
			adminQuestionnaires.POST("", PostQuestionnaire)
			adminQuestionnaires.PATCH("/:id", EditQuestionnaire)
			adminQuestionnaires.DELETE("/:id", DeleteQuestionnaire)
			adminQuestionnaires.GET("", GetQuestionnaires)
			adminQuestionnaires.GET("/:id", GetQuestionnaire)
		}

		adminVersion := admin.Group("/version")
		{
			adminVersion.POST("", PostVersionHandler)
			adminVersion.PUT("/:id", PutVersionHandler)
			adminVersion.DELETE("/:id", DeleteVersionHandler)
		}

		adminSpecial := admin.Group("/special")
		{
			adminSpecial.POST("/:id", PostSpecialHandler)
			adminSpecial.DELETE("/:id", DeleteSpecialHandler)
		}

		adminAdmin := admin.Group("/admin")
		{
			adminAdmin.GET("", GetAdminsHandler)
			adminAdmin.POST("", PostAdminsHandler)
			adminAdmin.DELETE("", DeleteAdminHandler)
		}
	}
}
