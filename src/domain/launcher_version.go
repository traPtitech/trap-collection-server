package domain

import (
	"errors"
	"time"

	"github.com/traPtitech/trap-collection-server/src/domain/values"
)

// Edition
// ランチャーのバージョンを表すドメイン。
// 現在の仕様では、バージョン名、アンケートURLの変更はできないため、
// SetName、SetQuestionnaireURLは使われない。
// 工大祭などのイベント用ランチャーバージョンではアンケートを持つが、
// コミケでの販売用ランチャーバージョンではアンケートを持たない。
type Edition struct {
	id                values.EditionID
	name              values.EditionName
	haveQuestionnaire bool
	questionnaireURL  values.EditionQuestionnaireURL
	createdAt         time.Time
}

func NewEditionWithoutQuestionnaire(
	id values.EditionID,
	name values.EditionName,
	createdAt time.Time,
) *Edition {
	return &Edition{
		id:                id,
		name:              name,
		haveQuestionnaire: false,
		createdAt:         createdAt,
	}
}

func NewEditionWithQuestionnaire(
	id values.EditionID,
	name values.EditionName,
	questionnaireURL values.EditionQuestionnaireURL,
	createdAt time.Time,
) *Edition {
	return &Edition{
		id:                id,
		name:              name,
		haveQuestionnaire: true,
		questionnaireURL:  questionnaireURL,
		createdAt:         createdAt,
	}
}

func (lv *Edition) GetID() values.EditionID {
	return lv.id
}

func (lv *Edition) GetName() values.EditionName {
	return lv.name
}

func (lv *Edition) SetName(name values.EditionName) {
	lv.name = name
}

var (
	ErrNoQuestionnaire = errors.New("no questionnaire")
)

func (lv *Edition) GetQuestionnaireURL() (values.EditionQuestionnaireURL, error) {
	if !lv.haveQuestionnaire {
		return nil, ErrNoQuestionnaire
	}

	return lv.questionnaireURL, nil
}

func (lv *Edition) SetQuestionnaireURL(questionnaireURL values.EditionQuestionnaireURL) {
	lv.questionnaireURL = questionnaireURL
	lv.haveQuestionnaire = true
}

func (lv *Edition) UnsetQuestionnaireURL() {
	lv.questionnaireURL = nil
	lv.haveQuestionnaire = false
}

func (lv *Edition) GetCreatedAt() time.Time {
	return lv.createdAt
}
