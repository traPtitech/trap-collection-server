package domain

import "github.com/traPtitech/trap-collection-server/domain/values"

type GameFile struct {
	id values.GameAssetID
	fileType values.GameFileType
	md5 values.GameFileMd5
}

func NewGameFile(id values.GameAssetID, fileType values.GameFileType, md5 values.GameFileMd5) *GameFile {
	return &GameFile{
		id: id,
		fileType: fileType,
		md5: md5,
	}
}

func (gf *GameFile) GetID() values.GameAssetID {
	return gf.id
}

func (gf *GameFile) GetType() values.GameFileType {
	return gf.fileType
}

func (gf * GameFile) GetMd5() values.GameFileMd5 {
	return gf.md5
}
