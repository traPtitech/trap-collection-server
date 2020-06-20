package model
//go:generate mockgen -source=$GOFILE -destination=mock_${GOFILE} -package=$GOPACKAGE

import (
	"fmt"
	"time"

	"github.com/traPtitech/trap-collection-server/openapi"
)

// Question 質問の構造体
type Question struct {
	ID                uint             `gorm:"type:int(11) unsigned auto_increment;primary_key;"`
	LauncherVersionID uint             `gorm:"type:int(11) unsigned;not null;"`
	QuestionNum       uint             `gorm:"type:int(11) unsigned;not null;"`
	Type              uint8            `gorm:"type:tinyint unsigned;not null;"`
	Content           string           `gorm:"type:text;not null;"`
	Required          bool             `gorm:"type:boolean;not null;default:true;"`
	QuestionOptions   []QuestionOption `gorm:"foreign_key:QuestionID;"`
	CreatedAt         time.Time        `gorm:"type:datetime;not null;default:current_timestamp;"`
	DeletedAt         time.Time        `gorm:"type:datetime;default:null;"`
}

// QuestionMeta questionテーブルのリポジトリ
type QuestionMeta interface {
	GetQuestions(versionID uint) ([]*openapi.Question, error)
}

// GetQuestions 質問の取得
func (*DB) GetQuestions(versionID uint) ([]*openapi.Question, error) {
	rows, err := db.Table("questions").
		Select("questions.id, questions.type, questions.content, question.required, questions.created_at, question_options.id, question_options.label").
		Joins("LEFT OUTER JOIN question_options ON questions.id = question_options.question_id").
		Where("questions.launcher_version_id = ?", versionID).Rows()
	if err != nil {
		return []*openapi.Question{}, fmt.Errorf("Failed In Getting Questions: %w", err)
	}
	questionMap := make(map[int32]*openapi.Question)
	for rows.Next() {
		var question *openapi.Question
		var option openapi.QuestionOption
		err = rows.Scan(&question.Id, &question.Type, &question.Content, &question.CreatedAt, &option.Id, &option.Label)
		if err != nil {
			return []*openapi.Question{}, fmt.Errorf("Failed In Scanning Questions: %w", err)
		}
		if _, ok := questionMap[question.Id]; ok {
			question = questionMap[question.Id]
		}
		if len(option.Label) != 0 {
			question.Options = append(question.Options, option)
		}
		questionMap[question.Id] = question
	}
	questions := make([]*openapi.Question, 0, len(questionMap))
	for _, v := range questionMap {
		questions = append(questions, v)
	}
	return questions, nil
}
