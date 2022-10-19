package domain

import (
	"errors"
	"time"

	"github.com/traPtitech/trap-collection-server/src/domain/values"
)

// LauncherVersion
// ランチャーのバージョンを表すドメイン。
// 現在の仕様では、バージョン名、アンケートURLの変更はできないため、
// SetName、SetQuestionnaireURLは使われない。
// 工大祭などのイベント用ランチャーバージョンではアンケートを持つが、
// コミケでの販売用ランチャーバージョンではアンケートを持たない。
type LauncherVersion struct {
	id                values.LauncherVersionID
	name              values.LauncherVersionName
	haveQuestionnaire bool
	questionnaireURL  values.LauncherVersionQuestionnaireURL
	createdAt         time.Time
}

func NewLauncherVersionWithoutQuestionnaire(
	id values.LauncherVersionID,
	name values.LauncherVersionName,
	createdAt time.Time,
) *LauncherVersion {
	return &LauncherVersion{
		id:                id,
		name:              name,
		haveQuestionnaire: false,
		createdAt:         createdAt,
	}
}

func NewLauncherVersionWithQuestionnaire(
	id values.LauncherVersionID,
	name values.LauncherVersionName,
	questionnaireURL values.LauncherVersionQuestionnaireURL,
	createdAt time.Time,
) *LauncherVersion {
	return &LauncherVersion{
		id:                id,
		name:              name,
		haveQuestionnaire: true,
		questionnaireURL:  questionnaireURL,
		createdAt:         createdAt,
	}
}

func (lv *LauncherVersion) GetID() values.LauncherVersionID {
	return lv.id
}

func (lv *LauncherVersion) GetName() values.LauncherVersionName {
	return lv.name
}

func (lv *LauncherVersion) SetName(name values.LauncherVersionName) {
	lv.name = name
}

var (
	ErrNoQuestionnaire = errors.New("no questionnaire")
)

func (lv *LauncherVersion) GetQuestionnaireURL() (values.LauncherVersionQuestionnaireURL, error) {
	if !lv.haveQuestionnaire {
		return nil, ErrNoQuestionnaire
	}

	return lv.questionnaireURL, nil
}

func (lv *LauncherVersion) SetQuestionnaireURL(questionnaireURL values.LauncherVersionQuestionnaireURL) {
	lv.questionnaireURL = questionnaireURL
	lv.haveQuestionnaire = true
}

func (lv *LauncherVersion) UnsetQuestionnaireURL() {
	lv.questionnaireURL = nil
	lv.haveQuestionnaire = false
}

func (lv *LauncherVersion) GetCreatedAt() time.Time {
	return lv.createdAt
}
