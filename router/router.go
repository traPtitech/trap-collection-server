package router

import (

	"github.com/labstack/echo"
)

//SetRouting ルーティング
func SetRouting(e *echo.Echo, client Traq) {
	api := e.Group("/api")
	{
		//check := api.Group("/check")
		//{
		//	check.POST("/:version/other", CheckHandler)
		//}

		//game := api.Group("/game")
		//{
		//	game.GET("/:id", DownloadHandler)
		//}
	
		responses := api.Group("/responses")
		{
			responses.POST("", PostResponse)
		}

		questionnaires := api.Group("/questionnaires")
		{
			questionnaires.GET("/:id/questions", GetQuestions)
		}

		time := api.Group("/time")
		{
			time.POST("", PostTimeHandler)
		}

		seat := api.Group("/seat")
		{
			seat.POST("", PostSeatHandler)
			seat.GET("", GetSeatHandler)
		}

		trap := api.Group("/trap", client.MiddlewareAuthUser)
		{
			//trapGame := trap.Group("/game")
			//{
			//	trapGame.POST("", PostGameHandler)
			//	trapGame.PUT("", PutGameHandler)
			//	trapGame.DELETE("/:id", DeleteGameHandler)
			//	trapGame.GET("", GetGameNameListHandler)
			//}

			trapResponses := trap.Group("/responses")
			{
				trapResponses.GET("/:id", GetResponse)
			}

			trapUsers := trap.Group("/users")
			{
				trapUsers.GET("/me", GetUsersMe)
			}

			trapResults := trap.Group("/results")
			{
				trapResults.GET("/:questionnaireID", GetResponsesByID)
			}

			//trapVersion := trap.Group("/version")
			//{
			//	trapVersion.GET("/sale", GetVersionForSaleListHandler)
			//	trapVersion.GET("/fes", GetVersionNotForSaleListHandler)
			//	trapVersion.GET("/:id/game", GetGameListHandler)
			//	trapVersion.GET("/:id/nongame", GetNonGameListHandler)
			//	trapVersion.GET("/:id/questionnaire", GetQuestionnaireHandler)
			//}
		}

		admin := api.Group("/admin", client.MiddlewareAuthUser)
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
}
