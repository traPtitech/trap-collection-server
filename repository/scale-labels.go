package repository

import (
	"net/http"

	"github.com/labstack/echo"
	"github.com/traPtitech/trap-collection-server/model"
)

//GetScaleLabels 目盛りの取得
func GetScaleLabels(c echo.Context, questionID string) (model.ScaleLabels, error) {
	scalelabel := model.ScaleLabels{}
	if err := Db.Get(&scalelabel, "SELECT * FROM scale_labels WHERE question_id = ?",
		questionID); err != nil {
		c.Logger().Error(err)
		return model.ScaleLabels{}, echo.NewHTTPError(http.StatusInternalServerError)
	}
	return scalelabel, nil
}

//InsertScaleLabels 目盛りの追加
func InsertScaleLabels(c echo.Context, lastID string, label model.ScaleLabels) error {
	if _, err := Db.Exec(
		"INSERT INTO scale_labels (question_id, scale_label_left, scale_label_right, scale_min, scale_max) VALUES (?, ?, ?, ?, ?)",
		lastID, label.ScaleLabelLeft, label.ScaleLabelRight, label.ScaleMin, label.ScaleMax); err != nil {
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}
	return nil
}

//UpdateScaleLabels 目盛りの変更
func UpdateScaleLabels(c echo.Context, questionID string, label model.ScaleLabels) error {
	if _, err := Db.Exec(
		`INSERT INTO scale_labels (question_id, scale_label_right, scale_label_left, scale_min, scale_max) VALUES (?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE scale_label_right = ?, scale_label_left = ?, scale_min = ?, scale_max = ?`,
		questionID,
		label.ScaleLabelRight, label.ScaleLabelLeft, label.ScaleMin, label.ScaleMax,
		label.ScaleLabelRight, label.ScaleLabelLeft, label.ScaleMin, label.ScaleMax); err != nil {
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}
	return nil
}

//DeleteScaleLabels 目盛りの削除
func DeleteScaleLabels(c echo.Context, questionID string) error {
	if _, err := Db.Exec(
		"DELETE FROM scale_labels WHERE question_id= ?",
		questionID); err != nil {
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}
	return nil
}
