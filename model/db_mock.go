// +build !main

package model

import gomock "github.com/golang/mock/gomock"


// DBMock DB関連の構造体のMock
type DBMock struct {
	*MockGameIntroductionMeta
	*MockGameVersionRelationMeta
	*MockGameVersionMeta
	*MockGameMeta
	*MockLauncherVersionMeta
	*MockMaintainerMeta
	*MockPlayerMeta
	*MockProductKeyMeta
	*MockQuestionMeta
	*MockResponseMeta
}

// NewDBMock DBのMockのコンストラクタ
func NewDBMock(ctrl *gomock.Controller) *DBMock {
	dbMock := new(DBMock)

	dbMock.MockGameIntroductionMeta = NewMockGameIntroductionMeta(ctrl)
	dbMock.MockGameVersionRelationMeta = NewMockGameVersionRelationMeta(ctrl)
	dbMock.MockGameVersionMeta = NewMockGameVersionMeta(ctrl)
	dbMock.MockGameMeta = NewMockGameMeta(ctrl)
	dbMock.MockLauncherVersionMeta = NewMockLauncherVersionMeta(ctrl)
	dbMock.MockMaintainerMeta = NewMockMaintainerMeta(ctrl)
	dbMock.MockPlayerMeta = NewMockPlayerMeta(ctrl)
	dbMock.MockProductKeyMeta = NewMockProductKeyMeta(ctrl)
	dbMock.MockQuestionMeta = NewMockQuestionMeta(ctrl)
	dbMock.MockResponseMeta = NewMockResponseMeta(ctrl)

	return dbMock
}
