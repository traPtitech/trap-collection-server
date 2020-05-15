package model

import (
	"fmt"

	"github.com/traPtitech/trap-collection-server/openapi"
)

var typeMap = map[uint8]string{
	0: "radio",
	1: "checkbox",
	2: "text",
}

// GetQuestions 質問の取得
func GetQuestions(versionID uint) ([]openapi.Question, error) {
	rows, err := db.Table("questions").
		Select("questions.id, questions.type, questions.content, question.required, questions.created_at, question_options.id, question_options.label").
		Joins("LEFT OUTER JOIN question_options ON questions.id = question_options.question_id").
		Where("questions.launcher_version_id = ?", versionID).Rows()
	if err != nil {
		return []openapi.Question{}, fmt.Errorf("Failed In Getting Questions: %w", err)
	}
	questionMap := make(map[int32]openapi.Question)
	for rows.Next() {
		var question openapi.Question
		var option openapi.QuestionOption
		err = rows.Scan(&question.Id, &question.Type, &question.Content, &question.CreatedAt, &option.Id, &option.Label)
		if err != nil {
			return []openapi.Question{}, fmt.Errorf("Failed In Scanning Questions: %w", err)
		}
		if _, ok := questionMap[question.Id]; ok {
			question = questionMap[question.Id]
		}
		if len(option.Label) != 0 {
			question.Options = append(question.Options, option)
		}
		questionMap[question.Id] = question
	}
	questions := make([]openapi.Question, 0, len(questionMap))
	for _, v := range questionMap {
		questions = append(questions, v)
	}
	return questions, nil
}
