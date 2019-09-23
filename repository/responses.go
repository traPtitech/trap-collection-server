package repository

import (
	"net/http"
	"strconv"
	"time"

	"database/sql"

	"github.com/labstack/echo"
	"github.com/traPtitech/trap-collection-server/model"
)

//InsertRespondents 回答セットの追加
func InsertRespondents(c echo.Context, req model.Responses) (int, error) {
	var result sql.Result
	var err error
	if req.SubmittedAt == "" || req.SubmittedAt == "NULL" {
		req.SubmittedAt = "NULL"
		if result, err = Db.Exec(
			`INSERT INTO respondents (questionnaire_id, user_traqid, modified_at) VALUES (?, ?, ?)`,
			req.ID, GetUserID(c), time.Now()); err != nil {
			c.Logger().Error(err)
			return 0, echo.NewHTTPError(http.StatusInternalServerError)
		}
	} else {
		if result, err = Db.Exec(
			`INSERT INTO respondents
				(questionnaire_id, user_traqid, submitted_at, modified_at) VALUES (?, ?, ?, ?)`,
			req.ID, GetUserID(c), req.SubmittedAt, time.Now()); err != nil {
			c.Logger().Error(err)
			return 0, echo.NewHTTPError(http.StatusInternalServerError)
		}
	}
	lastID, err := result.LastInsertId()
	if err != nil {
		c.Logger().Error(err)
		return 0, echo.NewHTTPError(http.StatusInternalServerError)
	}
	return int(lastID), nil
}

//InsertResponse 回答の追加
func InsertResponse(c echo.Context, responseID int, req model.Responses, body model.ResponseBody, data string) error {
	if _, err := Db.Exec(
		`INSERT INTO response (response_id, question_id, body, modified_at) VALUES (?, ?, ?, ?)`,
		responseID, body.QuestionID, data, time.Now()); err != nil {
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}
	return nil
}

//GetRespondents 回答の取得
func GetRespondents(c echo.Context, questionnaireID int) ([]string, error) {
	respondents := []string{}
	if err := Db.Select(&respondents,
		"SELECT user_traqid FROM respondents WHERE questionnaire_id = ? AND deleted_at IS NULL AND submitted_at IS NOT NULL",
		questionnaireID); err != nil {
		c.Logger().Error(err)
		return nil, echo.NewHTTPError(http.StatusInternalServerError)
	}
	return respondents, nil
}

//GetMyResponses 自分の回答の取得
func GetMyResponses(c echo.Context) ([]model.ResponseInfo, error) {
	responsesinfo := []model.ResponseInfo{}

	if err := Db.Select(&responsesinfo,
		`SELECT questionnaire_id, response_id, modified_at, submitted_at from respondents
		WHERE user_traqid = ? AND deleted_at IS NULL ORDER BY modified_at DESC`,
		GetUserID(c)); err != nil {
		c.Logger().Error(err)
		return []model.ResponseInfo{}, echo.NewHTTPError(http.StatusInternalServerError)
	}
	return responsesinfo, nil
}

//GetResponsesInfo 回答の取得
func GetResponsesInfo(c echo.Context, responsesinfo []model.ResponseInfo) ([]model.MyResponse, error) {
	myresponses := []model.MyResponse{}

	for _, response := range responsesinfo {
		title, resTimeLimit, err := GetTitleAndLimit(c, response.QuestionnaireID)
		if title == "" {
			continue
		}
		if err != nil {
			return nil, err
		}
		myresponses = append(myresponses,
			model.MyResponse{
				ResponseID:      response.ResponseID,
				QuestionnaireID: response.QuestionnaireID,
				Title:           title,
				ResTimeLimit:    resTimeLimit,
				SubmittedAt:     NullTimeToString(response.SubmittedAt),
				ModifiedAt:      response.ModifiedAt.Format(time.RFC3339),
			})
	}
	return myresponses, nil
}

//GetResponseBody 回答内容の取得
func GetResponseBody(c echo.Context, responseID int, questionID int, questionType string) (model.ResponseBody, error) {
	body := model.ResponseBody{
		QuestionID:   questionID,
		QuestionType: questionType,
	}
	switch questionType {
	case "MultipleChoice", "Checkbox", "Dropdown":
		option := []string{}
		if err := Db.Select(&option,
			`SELECT body from response
			WHERE response_id = ? AND question_id = ? AND deleted_at IS NULL`,
			responseID, body.QuestionID); err != nil {
			c.Logger().Error(err)
			return model.ResponseBody{}, echo.NewHTTPError(http.StatusInternalServerError)
		}
		body.OptionResponse = option
		// sortで比較するため
		for _, op := range option {
			if body.Response != "" {
				body.Response += ", "
			}
			body.Response += op
		}
	default:
		var response string
		if err := Db.Get(&response,
			`SELECT body from response
			WHERE response_id = ? AND question_id = ? AND deleted_at IS NULL`,
			responseID, body.QuestionID); err != nil {
			if err != sql.ErrNoRows {
				c.Logger().Error(err)
				return model.ResponseBody{}, echo.NewHTTPError(http.StatusInternalServerError)
			}
		}
		body.Response = response
	}
	return body, nil
}

//RespondedAt 回答時刻の取得
func RespondedAt(c echo.Context, questionnaireID int) (string, error) {
	respondedAt := sql.NullString{}
	if err := Db.Get(&respondedAt,
		`SELECT MAX(submitted_at) FROM respondents
		WHERE user_traqid = ? AND questionnaire_id = ? AND deleted_at IS NULL`,
		GetUserID(c), questionnaireID); err != nil {
		c.Logger().Error(err)
		return "", echo.NewHTTPError(http.StatusInternalServerError)
	}
	return NullStringConvert(respondedAt), nil
}

//GetRespondentByID IDによる回答の取得
func GetRespondentByID(c echo.Context, responseID int) (model.ResponseID, error) {
	respondentInfo := model.ResponseID{}
	if err := Db.Get(&respondentInfo,
		`SELECT questionnaire_id, modified_at, submitted_at from respondents
		WHERE response_id = ? AND deleted_at IS NULL`,
		responseID); err != nil {
		if err != sql.ErrNoRows {
			c.Logger().Error(err)
			return model.ResponseID{}, echo.NewHTTPError(http.StatusInternalServerError)
		}
	}
	return respondentInfo, nil
}

//UpdateRespondents 回答の変更
func UpdateRespondents(c echo.Context, questionnaireID int, responseID int, submittedAt string) error {
	if submittedAt == "" || submittedAt == "NULL" {
		submittedAt = "NULL"
		if _, err := Db.Exec(
			`UPDATE respondents
			SET questionnaire_id = ?, submitted_at = NULL, modified_at = ? WHERE response_id = ?`,
			questionnaireID, time.Now(), responseID); err != nil {
			c.Logger().Error(err)
			return echo.NewHTTPError(http.StatusInternalServerError)
		}
	} else {
		if _, err := Db.Exec(
			`UPDATE respondents
			SET questionnaire_id = ?, submitted_at = ?, modified_at = ? WHERE response_id = ?`,
			questionnaireID, submittedAt, time.Now(), responseID); err != nil {
			c.Logger().Error(err)
			return echo.NewHTTPError(http.StatusInternalServerError)
		}
	}
	return nil

}

//DeleteResponse 回答の削除
func DeleteResponse(c echo.Context, responseID int) error {
	if _, err := Db.Exec(
		`UPDATE response SET deleted_at = ? WHERE response_id = ?`,
		time.Now(), responseID); err != nil {
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}
	return nil
}

//GetResponsesByID IDによる回答の取得
func GetResponsesByID(questionnaireID int) ([]model.ResponseAnDBody, error) {
	responses := []model.ResponseAnDBody{}
	if err := Db.Select(&responses,
		`SELECT respondents.response_id AS response_id,
		user_traqid, 
		respondents.modified_at AS modified_at,
		respondents.submitted_at AS submitted_at,
		response.question_id,
		response.body
		FROM respondents
		RIGHT OUTER JOIN response
		ON respondents.response_id = response.response_id
		WHERE respondents.questionnaire_id = ?
		AND respondents.deleted_at IS NULL
		AND response.deleted_at IS NULL
		AND respondents.submitted_at IS NOT NULL`, questionnaireID); err != nil {
		return nil, echo.NewHTTPError(http.StatusInternalServerError)
	}
	return responses, nil
}

//GetResponseBodyList 回答一覧
func GetResponseBodyList(c echo.Context, questionTypeList []model.QuestionIDType, responses []model.QIDandResponse) []model.ResponseBody {
	bodyList := []model.ResponseBody{}

	for _, qType := range questionTypeList {
		response := ""
		optionResponse := []string{}
		for _, respInfo := range responses {
			// 質問IDが一致したら追加
			if qType.ID == respInfo.QuestionID {
				switch qType.Type {
				case "MultipleChoice", "Checkbox", "Dropdown":
					if response != "" {
						response += ","
					}
					response += respInfo.Response
					optionResponse = append(optionResponse, respInfo.Response)
				default:
					response += respInfo.Response
				}
			}
		}
		// 回答内容の配列に追加
		bodyList = append(bodyList,
			model.ResponseBody{
				QuestionID:     qType.ID,
				QuestionType:   qType.Type,
				Response:       response,
				OptionResponse: optionResponse,
			})
	}
	return bodyList
}

//GetSortedRespondents sortされた回答者の情報を返す
func GetSortedRespondents(c echo.Context, questionnaireID int, sortQuery string) ([]model.UserResponse, int, error) {
	sql := `SELECT response_id, user_traqid, modified_at, submitted_at from respondents
			WHERE deleted_at IS NULL AND questionnaire_id = ? AND submitted_at IS NOT NULL`

	sortNum := 0
	switch sortQuery {
	case "traqid":
		sql += ` ORDER BY user_traqid`
	case "-traqid":
		sql += ` ORDER BY user_traqid DESC`
	case "submitted_at":
		sql += ` ORDER BY submitted_at`
	case "-submitted_at":
		sql += ` ORDER BY submitted_at DESC`
	case "":
	default:
		var err error
		sortNum, err = strconv.Atoi(sortQuery)
		if err != nil {
			c.Logger().Error(err)
			return []model.UserResponse{}, 0, echo.NewHTTPError(http.StatusBadRequest)
		}
	}

	responsesinfo := []model.UserResponse{}
	if err := Db.Select(&responsesinfo, sql,
		questionnaireID); err != nil {
		c.Logger().Error(err)
		return []model.UserResponse{}, 0, echo.NewHTTPError(http.StatusInternalServerError)
	}
	return responsesinfo, sortNum, nil
}
