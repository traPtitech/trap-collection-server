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

//SetRouting ルーティング
func SetRouting(e *echo.Echo) {

	Questions := e.Group("/questions")
	{
		Questions.POST("", PostQuestion)
		Questions.PATCH("/:id", EditQuestion)
		Questions.DELETE("/:id", DeleteQuestion)
	}

	check := e.Group("/check")
	{
		check.POST("", CheckHandler)
	}

	download := e.Group("/download")
	{
		download.GET("/:name", DownloadHandler)
	}

	api := e.Group("/api", UserAuthenticate())
	{
		apiQuestionnnaires := api.Group("/questionnaires")
		{
			apiQuestionnnaires.GET("", GetQuestionnaires)
			apiQuestionnnaires.POST("", PostQuestionnaire)
			apiQuestionnnaires.GET("/:id", GetQuestionnaire)
			apiQuestionnnaires.PATCH("/:id", EditQuestionnaire)
			apiQuestionnnaires.DELETE("/:id", DeleteQuestionnaire)
			apiQuestionnnaires.GET("/:id/questions", GetQuestions)
		}

		apiResponses := api.Group("/responses")
		{
			apiResponses.POST("", PostResponse)
			apiResponses.GET("/:id", GetResponse)
		}

		apiUsers := api.Group("/users")
		{
			/*
				TODO
				apiUsers.GET("")
			*/
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
