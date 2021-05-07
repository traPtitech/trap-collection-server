package domain

import "github.com/traPtitech/trap-collection-server/domain/values"

type GameURL struct {
	id values.GameAssetID
	url values.GameURL
}

func NewGameURL(id values.GameAssetID, url values.GameURL) *GameURL {
	return &GameURL{
		id: id,
		url: url,
	}
}

func (gurl *GameURL) GetID() values.GameAssetID {
	return gurl.id
}

func (gurl *GameURL) GetURL() values.GameURL {
	return gurl.url
}
