package router

import (
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/labstack/echo"

	"github.com/traPtitech/trap-collection-server/model"
	"github.com/traPtitech/trap-collection-server/repository"
)

// PostResponse POST /responses
func PostResponse(c echo.Context) error {

	req := model.Responses{}

	if err := c.Bind(&req); err != nil {
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusBadRequest)
	}

	limit, err := repository.GetQuestionnaireLimit(c, req.ID)
	if err != nil {
		return err
	}

	// 回答期限を過ぎた回答は許可しない
	if limit != "NULL" && limit < time.Now().Format(time.RFC3339) {
		return echo.NewHTTPError(http.StatusMethodNotAllowed)
	}

	responseID, err := repository.InsertRespondents(c, req)
	if err != nil {
		return err
	}

	for _, body := range req.Body {
		switch body.QuestionType {
		case "MultipleChoice", "Checkbox", "Dropdown":
			for _, option := range body.OptionResponse {
				if err := repository.InsertResponse(c, responseID, req, body, option); err != nil {
					return err
				}
			}
		default:
			if err := repository.InsertResponse(c, responseID, req, body, body.Response); err != nil {
				return err
			}
		}
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"responseID":      responseID,
		"questionnaireID": req.ID,
		"submitted_at":    req.SubmittedAt,
		"body":            req.Body,
	})
}

// GetResponsesByID GET /results/:questionnaireID
func GetResponsesByID(c echo.Context) error {
	questionnaireID, err := strconv.Atoi(c.Param("questionnaireID"))
	if err != nil {
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusBadRequest)
	}

	resSharedTo, err := repository.GetResShared(c, questionnaireID)
	if err != nil {
		return err
	}

	// アンケートの回答を確認する権限が無ければエラーを返す
	if err := repository.CheckResponseConfirmable(c, resSharedTo, questionnaireID); err != nil {
		return err
	}

	sortQuery := c.QueryParam("sort")
	// sortされた回答者の情報
	respondents, sortNum, err := repository.GetSortedRespondents(c, questionnaireID, sortQuery)
	if err != nil {
		return err
	}

	// 必要な回答を一気に持ってくる
	responses, err := repository.GetResponsesByID(questionnaireID)
	if err != nil {
		return err
	}

	// 各回答者のアンケートIDと回答
	resMap := map[int][]model.QIDandResponse{}
	for _, resp := range responses {
		resMap[resp.ResponseID] = append(resMap[resp.ResponseID],
			model.QIDandResponse{
				QuestionID: resp.QuestionID,
				Response:   resp.Body,
			})
	}

	// 質問IDと種類を取ってくる
	questionTypeList, err := repository.GetQuestionsType(c, questionnaireID)
	if err != nil {
		return err
	}

	// 返す構造体
	type ReturnInfo struct {
		ResponseID  int                  `json:"responseID"`
		UserID      string               `json:"traqID"`
		SubmittedAt string               `json:"submitted_at"`
		ModifiedAt  string               `json:"modified_at"`
		Body        []model.ResponseBody `json:"response_body"`
	}
	returnInfo := []ReturnInfo{}

	for _, respondent := range respondents {
		bodyList := repository.GetResponseBodyList(c, questionTypeList, resMap[respondent.ResponseID])
		// 回答の配列に追加
		returnInfo = append(returnInfo,
			ReturnInfo{
				ResponseID:  respondent.ResponseID,
				SubmittedAt: repository.NullTimeToString(respondent.SubmittedAt),
				ModifiedAt:  respondent.ModifiedAt.Format(time.RFC3339),
				Body:        bodyList,
			})
	}

	// 昇順ソート
	if sortNum > 0 {
		sort.Slice(returnInfo, func(i, j int) bool {
			bodyI := returnInfo[i].Body[sortNum-1]
			bodyJ := returnInfo[j].Body[sortNum-1]
			if bodyI.QuestionType == "Number" {
				numi, err := strconv.Atoi(bodyI.Response)
				if err != nil {
					return true
				}
				numj, err := strconv.Atoi(bodyJ.Response)
				if err != nil {
					return true
				}
				return numi < numj
			}
			return bodyI.Response < bodyJ.Response
		})
	}
	// 降順ソート
	if sortNum < 0 {
		sort.Slice(returnInfo, func(i, j int) bool {
			bodyI := returnInfo[i].Body[-sortNum-1]
			bodyJ := returnInfo[j].Body[-sortNum-1]
			if bodyI.QuestionType == "Number" {
				numi, err := strconv.Atoi(bodyI.Response)
				if err != nil {
					return true
				}
				numj, err := strconv.Atoi(bodyJ.Response)
				if err != nil {
					return true
				}
				return numi > numj
			}
			return bodyI.Response > bodyJ.Response
		})
	}

	return c.JSON(http.StatusOK, returnInfo)
}

// GetResponse GET /responses
func GetResponse(c echo.Context) error {
	responseID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest)
	}

	respondentInfo, err := repository.GetRespondentByID(c, responseID)
	if err != nil {
		return nil
	}

	responses := struct {
		QuestionnaireID int                  `json:"questionnaireID"`
		SubmittedAt     string               `json:"submitted_at"`
		ModifiedAt      string               `json:"modified_at"`
		Body            []model.ResponseBody `json:"body"`
	}{
		respondentInfo.QuestionnaireID,
		repository.NullTimeToString(respondentInfo.SubmittedAt),
		repository.NullTimeToString(respondentInfo.ModifiedAt),
		[]model.ResponseBody{},
	}

	questionTypeList, err := repository.GetQuestionsType(c, responses.QuestionnaireID)
	if err != nil {
		return err
	}

	bodyList := []model.ResponseBody{}
	for _, questionType := range questionTypeList {
		body, err := repository.GetResponseBody(c, responseID, questionType.ID, questionType.Type)
		if err != nil {
			return err
		}
		bodyList = append(bodyList, body)
	}
	responses.Body = bodyList
	return c.JSON(http.StatusOK, responses)
}
