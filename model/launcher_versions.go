package model

//go:generate mockgen -source=$GOFILE -destination=mock_${GOFILE} -package=$GOPACKAGE

import (
	"fmt"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/traPtitech/trap-collection-server/openapi"
)

// LauncherVersion ランチャーのバージョンの構造体
type LauncherVersion struct {
	ID                   uint                  `json:"id" gorm:"type:int(11) unsigned auto_increment;PRIMARY_KEY;"`
	Name                 string                `json:"name,omitempty" gorm:"type:varchar(32);NOT NULL;UNIQUE;"`
	GameVersionRelations []GameVersionRelation `json:"games" gorm:"foreignkey:LauncherVersionID;"`
	Questions            []Question            `json:"questions" gorm:"foreignkey:LauncherVersionID;"`
	CreatedAt            time.Time             `json:"created_at,omitempty" gorm:"type:datetime;NOT NULL;default:CURRENT_TIMESTAMP;"`
	DeletedAt            time.Time             `json:"deleted_at,omitempty" gorm:"type:datetime;default:NULL;"`
}

// LauncherVersionMeta launcher_versionテーブルのリポジトリ
type LauncherVersionMeta interface {
	GetLauncherVersionDetailsByID(id uint) (versionDetails *openapi.VersionDetails, err error)
	InsertLauncherVersion(name string) (*openapi.VersionMeta, error)
}

// GetLauncherVersionDetailsByID ランチャーのバージョンをIDから取得
func (*DB) GetLauncherVersionDetailsByID(id uint) (versionDetails *openapi.VersionDetails, err error) {
	versionDetails = &openapi.VersionDetails{}

	rows, err := db.Table("launcher_versions").
		Select("launcher_versions.id,launcher_versions.name,launcher_versions.created_at,game_version_relations.game_id").
		Joins("LEFT OUTER JOIN game_version_relations ON launcher_versions.id = game_version_relations.launcher_version_id").
		Where("launcher_versions.id = ?", id).
		Rows()
	if err != nil {
		return &openapi.VersionDetails{}, fmt.Errorf("Failed In Getting Launcher Versions:%w", err)
	}
	for rows.Next() {
		var gameID interface{}
		err = rows.Scan(&versionDetails.Id, &versionDetails.Name, &versionDetails.CreatedAt, &gameID)
		if err != nil {
			return &openapi.VersionDetails{}, fmt.Errorf("Failed In Scaning Launcher Version:%w", err)
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
		return &openapi.VersionDetails{}, fmt.Errorf("Failed In Getting Questions:%w", err)
	}
	questionMap := make(map[int32]openapi.Question)
	for rows.Next() {
		var question openapi.Question
		var option openapi.QuestionOption
		err = rows.Scan(&question.Id, &question.Type, &question.Content, &question.Required, &question.CreatedAt, &option.Id, &option.Label)
		if err != nil {
			return &openapi.VersionDetails{}, fmt.Errorf("Failed In Scaning Question:%w", err)
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

func (*DB) InsertLauncherVersion(name string) (*openapi.VersionMeta, error) {
	var apiVersion openapi.VersionMeta
	err := db.Transaction(func(tx *gorm.DB) error {
		launcherVersion := LauncherVersion{
			Name: name,
		}

		err := tx.Create(&launcherVersion).Error
		if err != nil {
			return fmt.Errorf("failed to insert a lancher version record: %w", err)
		}

		err = tx.Last(&launcherVersion).Error
		if err != nil {
			return fmt.Errorf("failed to get the last launcher version record: %w", err)
		}
		apiVersion = openapi.VersionMeta{
			Id: int32(launcherVersion.ID),
			Name: launcherVersion.Name,
			CreatedAt: launcherVersion.CreatedAt,
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed in transaction: %w", err)
	}

	return &apiVersion, nil
}
