package domain

import (
	"github.com/traPtitech/trap-collection-server/domain/values"
)

type LauncherVersion struct {
	id values.LauncherVersionID
	name values.LauncherVersionName
	questionnaireURL values.QuestionnaireURL
	createdAt values.LauncherVersionCreatedAt
	deletedAt values.LauncherVersionDeletedAt
}

func NewLauncherVersion(id values.LauncherVersionID, name values.LauncherVersionName, questionnaireURL values.QuestionnaireURL, createdAt values.LauncherVersionCreatedAt, deletedAt values.LauncherVersionDeletedAt) *LauncherVersion {
	return &LauncherVersion{
		id: id,
		name: name,
		questionnaireURL: questionnaireURL,
		createdAt: createdAt,
		deletedAt: deletedAt,
	}
}

func (lv *LauncherVersion) GetID() values.LauncherVersionID {
	return lv.id
}

func (lv *LauncherVersion) GetName() values.LauncherVersionName {
	return lv.name
}

func (lv *LauncherVersion) GetQuestionnaireURL() values.QuestionnaireURL {
	return lv.questionnaireURL
}

func (lv *LauncherVersion) GetCreatedAt() values.LauncherVersionCreatedAt {
	return lv.createdAt
}

func (lv *LauncherVersion) GetDeletedAt() values.LauncherVersionDeletedAt {
	return lv.deletedAt
}

func (lv *LauncherVersion) SetDeletedAt(deletedAt values.LauncherVersionDeletedAt) {
	lv.deletedAt = deletedAt
}
