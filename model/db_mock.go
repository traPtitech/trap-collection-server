// +build !main

package model

import gomock "github.com/golang/mock/gomock"

// DBMock DB関連の構造体のMock
type DBMock struct {
	*MockGameAssetMeta
	*MockGameIntroductionMeta
	*MockGameVersionRelationMeta
	*MockGameVersionMeta
	*MockGameMeta
	*MockLauncherVersionMeta
	*MockMaintainerMeta
	*MockProductKeyMeta
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
	dbMock.MockProductKeyMeta = NewMockProductKeyMeta(ctrl)

	return dbMock
}
