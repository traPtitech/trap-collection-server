package router

import (
	"net/http"
	"time"

	"github.com/labstack/echo"

	"github.com/traPtitech/trap-collection-server/repository"
)

// GetQuestionnaires GET /questionnaires
func GetQuestionnaires(c echo.Context) error {

	questionnaires, pageMax, err := repository.GetQuestionnaires(c, c.QueryParam("nontargeted") == "true")
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"page_max":       pageMax,
		"questionnaires": questionnaires,
	})
}

// GetQuestionnaire GET /questionnaires/:id
func GetQuestionnaire(c echo.Context) error {
	questionnaireID := c.Param("id")

	questionnaire, err := repository.GetQuestionnaireInfo(c, questionnaireID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"questionnaireID": questionnaire.ID,
		"title":           questionnaire.Title,
		"description":     questionnaire.Description,
		"res_time_limit":  repository.NullTimeToString(questionnaire.ResTimeLimit),
		"created_at":      questionnaire.CreatedAt.Format(time.RFC3339),
		"modified_at":     questionnaire.ModifiedAt.Format(time.RFC3339),
	})
}

// PostQuestionnaire POST /questionnaires
func PostQuestionnaire(c echo.Context) error {

	req := struct {
		Title        string `json:"title"`
		Description  string `json:"description"`
		ResTimeLimit string `json:"res_time_limit"`
	}{}

	// JSONを構造体につける
	if err := c.Bind(&req); err != nil {
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusBadRequest)
	}

	lastID, err := repository.InsertQuestionnaire(c, req.Title, req.Description, req.ResTimeLimit)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"questionnaireID": lastID,
		"title":           req.Title,
		"description":     req.Description,
		"res_time_limit":  req.ResTimeLimit,
		"deleted_at":      "NULL",
		"created_at":      time.Now().Format(time.RFC3339),
		"modified_at":     time.Now().Format(time.RFC3339),
	})
}

// EditQuestionnaire PATCH /questonnaires/:id
func EditQuestionnaire(c echo.Context) error {

	questionnaireID := c.Param("id")

	req := struct {
		Title        string   `json:"title"`
		Description  string   `json:"description"`
		ResTimeLimit string   `json:"res_time_limit"`
		ResSharedTo  string   `json:"res_shared_to"`
		Targets      []string `json:"targets"`
	}{}

	if err := c.Bind(&req); err != nil {
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusBadRequest)
	}

	if req.ResSharedTo == "" {
		req.ResSharedTo = "administrators"
	}

	if err := repository.UpdateQuestionnaire(
		c, req.Title, req.Description, req.ResTimeLimit, req.ResSharedTo, questionnaireID); err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}

// DeleteQuestionnaire DELETE /questonnaires/:id
func DeleteQuestionnaire(c echo.Context) error {
	questionnaireID := c.Param("id")

	if err := repository.DeleteQuestionnaire(c, questionnaireID); err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}
