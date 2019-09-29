package router

import (
	"net/http"
	"time"

	"github.com/labstack/echo"

	"github.com/traPtitech/trap-collection-server/model"
	"github.com/traPtitech/trap-collection-server/repository"
)

// GetQuestions GET /questions
func GetQuestions(c echo.Context) error {
	questionnaireID := c.Param("id")

	allquestions, err := repository.GetQuestions(c, questionnaireID)
	if err != nil {
		return err
	}

	if len(allquestions) == 0 {
		return echo.NewHTTPError(http.StatusNotFound)
	}

	type questionInfo struct {
		QuestionID      string   `json:"questionID"`
		PageNum         int      `json:"page_num"`
		QuestionNum     int      `json:"question_num"`
		QuestionType    string   `json:"question_type"`
		Body            string   `json:"body"`
		IsRequrired     bool     `json:"is_required"`
		CreatedAt       string   `json:"created_at"`
		Options         []string `json:"options"`
		ScaleLabelRight string   `json:"scale_label_right"`
		ScaleLabelLeft  string   `json:"scale_label_left"`
		ScaleMin        int      `json:"scale_min"`
		ScaleMax        int      `json:"scale_max"`
	}
	var ret []questionInfo

	for _, v := range allquestions {
		options := []string{}
		scalelabel := model.ScaleLabels{}
		var err error
		switch v.Type {
		case "MultipleChoice", "Checkbox", "Dropdown":
			options, err = repository.GetOptions(c, v.ID)
		case "LinearScale":
			scalelabel, err = repository.GetScaleLabels(c, v.ID)
		}
		if err != nil {
			return err
		}

		ret = append(ret,
			questionInfo{
				QuestionID:      v.ID,
				PageNum:         v.PageNum,
				QuestionNum:     v.QuestionNum,
				QuestionType:    v.Type,
				Body:            v.Body,
				IsRequrired:     v.IsRequrired,
				CreatedAt:       v.CreatedAt.Format(time.RFC3339),
				Options:         options,
				ScaleLabelRight: scalelabel.ScaleLabelRight,
				ScaleLabelLeft:  scalelabel.ScaleLabelLeft,
				ScaleMin:        scalelabel.ScaleMin,
				ScaleMax:        scalelabel.ScaleMax,
			})
	}

	return c.JSON(http.StatusOK, ret)
}

// PostQuestion POST /questions
func PostQuestion(c echo.Context) error {
	req := struct {
		QuestionnaireID string   `json:"questionnaireID"`
		QuestionType    string   `json:"question_type"`
		QuestionNum     int      `json:"question_num"`
		PageNum         int      `json:"page_num"`
		Body            string   `json:"body"`
		IsRequired      bool     `json:"is_required"`
		Options         []string `json:"options"`
		ScaleLabelRight string   `json:"scale_label_right"`
		ScaleLabelLeft  string   `json:"scale_label_left"`
		ScaleMin        int      `json:"scale_min"`
		ScaleMax        int      `json:"scale_max"`
	}{}

	if err := c.Bind(&req); err != nil {
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusBadRequest)
	}

	lastID, err := repository.InsertQuestion(
		c, req.QuestionnaireID, req.PageNum, req.QuestionNum, req.QuestionType, req.Body, req.IsRequired)
	if err != nil {
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	switch req.QuestionType {
	case "MultipleChoice", "Checkbox", "Dropdown":
		for i, v := range req.Options {
			if err := repository.InsertOption(c, lastID, i+1, v); err != nil {
				return err
			}
		}
	case "LinearScale":
		if err := repository.InsertScaleLabels(c, lastID,
			model.ScaleLabels{
				ScaleLabelLeft:  req.ScaleLabelLeft,
				ScaleLabelRight: req.ScaleLabelRight,
				ScaleMax:        req.ScaleMax,
				ScaleMin:        req.ScaleMin,
			}); err != nil {
			return err
		}
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"questionID":        lastID,
		"questionnaireID":   req.QuestionnaireID,
		"question_type":     req.QuestionType,
		"question_num":      req.QuestionNum,
		"page_num":          req.PageNum,
		"body":              req.Body,
		"is_required":       req.IsRequired,
		"options":           req.Options,
		"scale_label_right": req.ScaleLabelRight,
		"scale_label_left":  req.ScaleLabelLeft,
		"scale_max":         req.ScaleMax,
		"scale_min":         req.ScaleMin,
	})
}

// EditQuestion PATCH /questions/:id
func EditQuestion(c echo.Context) error {
	questionID := c.Param("id")

	req := struct {
		QuestionnaireID string   `json:"questionnaireID"`
		QuestionType    string   `json:"question_type"`
		QuestionNum     int      `json:"question_num"`
		PageNum         int      `json:"page_num"`
		Body            string   `json:"body"`
		IsRequired      bool     `json:"is_required"`
		Options         []string `json:"options"`
		ScaleLabelRight string   `json:"scale_label_right"`
		ScaleLabelLeft  string   `json:"scale_label_left"`
		ScaleMax        int      `json:"scale_max"`
		ScaleMin        int      `json:"scale_min"`
	}{}

	if err := c.Bind(&req); err != nil {
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusBadRequest)
	}

	if err := repository.UpdateQuestion(
		c, req.QuestionnaireID, req.PageNum, req.QuestionNum, req.QuestionType, req.Body,
		req.IsRequired, questionID); err != nil {
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	switch req.QuestionType {
	case "MultipleChoice", "Checkbox", "Dropdown":
		if err := repository.UpdateOptions(c, req.Options, questionID); err != nil {
			return err
		}
	case "LinearScale":
		if err := repository.UpdateScaleLabels(c, questionID,
			model.ScaleLabels{
				ScaleLabelLeft:  req.ScaleLabelLeft,
				ScaleLabelRight: req.ScaleLabelRight,
				ScaleMax:        req.ScaleMax,
				ScaleMin:        req.ScaleMin,
			}); err != nil {
			return err
		}
	}

	return c.NoContent(http.StatusOK)
}

// DeleteQuestion DELETE /questions/:id
func DeleteQuestion(c echo.Context) error {
	questionID := c.Param("id")

	if err := repository.DeleteQuestion(c, questionID); err != nil {
		return err
	}

	if err := repository.DeleteOptions(c, questionID); err != nil {
		return err
	}

	if err := repository.DeleteScaleLabels(c, questionID); err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}
