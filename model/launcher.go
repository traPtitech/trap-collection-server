package model

import (
	"fmt"

	"github.com/jinzhu/gorm"
	"github.com/traPtitech/trap-collection-server/openapi"
)

func getKeyIDByKey(key string) (uint, error) {
	productKey := ProductKey{}
	err := db.Where("`key` = ?", key).First(&productKey).Error
	if err != nil {
		return 0, fmt.Errorf("Failed In Getting Key ID: %w", err)
	}
	return productKey.ID, nil
}

// CheckProductKey プロダクトキーが正しいか確認
func CheckProductKey(key string) bool {
	productKey := ProductKey{}
	isNotThere := db.Where("`key` = ?", key).First(&productKey).RecordNotFound()
	return !isNotThere
}

// GetLauncherVersionDetailsByID ランチャーのバージョンをIDから取得
func GetLauncherVersionDetailsByID(id uint) (versionDetails openapi.VersionDetails, err error) {
	rows, err := db.Table("launcher_versions").
		Select("launcher_versions.id,launcher_versions.name,launcher_versions.created_at,game_version_relations.game_id").
		Joins("LEFT OUTER JOIN game_version_relations ON launcher_versions.id = game_version_relations.launcher_version_id").
		Where("launcher_versions.id = ?", id).
		Rows()
	if err != nil {
		return openapi.VersionDetails{}, fmt.Errorf("Failed In Getting Launcher Versions:%w", err)
	}
	for rows.Next() {
		var gameID interface{}
		err = rows.Scan(&versionDetails.Id, &versionDetails.Name, &versionDetails.CreatedAt, &gameID)
		if err != nil {
			return openapi.VersionDetails{}, fmt.Errorf("Failed In Scaning Launcher Version:%w", err)
		}
		if gameID != nil {
			versionDetails.Games = append(versionDetails.Games, gameID.(string))
		}
	}

	rows, err = db.Table("questions").
		Select("questions.id,questions.type,questions.content,questions.required,questions.created_at,question_options.id,question_options.label").
		Joins("LEFT OUTER JOIN question_options ON questions.id = question_options.question_id").
		Where("questions.launcher_version_id = ?", id).
		Rows()
	if err != nil && !gorm.IsRecordNotFoundError(err) {
		return openapi.VersionDetails{}, fmt.Errorf("Failed In Getting Questions:%w", err)
	}
	questionMap := make(map[int32]openapi.Question)
	for rows.Next() {
		var question openapi.Question
		var option openapi.QuestionOption
		err = rows.Scan(&question.Id, &question.Type, &question.Content, &question.Required, &question.CreatedAt, &option.Id, &option.Label)
		if err != nil {
			return openapi.VersionDetails{}, fmt.Errorf("Failed In Scaning Question:%w", err)
		}
		if _, ok := questionMap[question.Id]; ok {
			question = questionMap[question.Id]
		}
		if len(option.Label) != 0 {
			question.Options = append(question.Options, option)
		}
		questionMap[question.Id] = question
	}
	for _, v := range questionMap {
		versionDetails.Questions = append(versionDetails.Questions, v)
	}

	return
}
