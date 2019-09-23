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
			b,err := repository.CheckAdmin(c)
			if err != nil {
				return err
			}
			// 管理者でないユーザはアクセスできない
			if !b {
				return echo.NewHTTPError(http.StatusUnauthorized, "You are not logged in")
			}
			return next(c)
		}
	}
}

//SetRouting ルーティング
func SetRouting(e *echo.Echo) {

	Questions := e.Group("/questions")
	{
		Questions.POST("", PostQuestion)
		Questions.PATCH("/:id", EditQuestion)
		Questions.DELETE("/:id", DeleteQuestion)
	}

	responses := e.Group("/responses")
	{
		responses.POST("", PostResponse)
	}

	check := e.Group("/check")
	{
		check.POST("/:version", CheckHandler)
	}

	download := e.Group("/download")
	{
		download.GET("/:name", DownloadHandler)
	}

	api := e.Group("/api", UserAuthenticate())
	{
		admin := e.Group("/admin", AdminAuthenticate())
		{
			adminQuestionnaires := admin.Group("/questionnaires")
			{
				adminQuestionnaires.POST("", PostQuestionnaire)
				adminQuestionnaires.PATCH("/:id", EditQuestionnaire)
				adminQuestionnaires.DELETE("/:id", DeleteQuestionnaire)
			}
		}
		apiQuestionnnaires := api.Group("/questionnaires")
		{
			apiQuestionnnaires.GET("", GetQuestionnaires)
			apiQuestionnnaires.GET("/:id", GetQuestionnaire)
			apiQuestionnnaires.GET("/:id/questions", GetQuestions)
		}

		apiResponses := api.Group("/responses")
		{
			apiResponses.GET("/:id", GetResponse)
		}

		apiUsers := api.Group("/users")
		{
			apiUsersMe := apiUsers.Group("/me")
			{
				apiUsersMe.GET("", GetUsersMe)
			}
		}

		apiResults := api.Group("/results")
		{
			apiResults.GET("/:questionnaireID", GetResponsesByID)
		}

		game := api.Group("/game")
		{
			game.POST("", PostGameHandler)
			game.PUT("", PutGameHandler)
			game.DELETE("", DeleteGameHandler)
			game.GET("", GetGameNameListHandler)
		}
	}
}
