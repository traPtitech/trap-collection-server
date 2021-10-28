package domain

import "github.com/traPtitech/trap-collection-server/src/domain/values"

type GameURL struct {
	id   values.GameURLID
	link values.GameURLLink
}

func NewGameURL(
	id values.GameURLID,
	link values.GameURLLink,
) *GameURL {
	return &GameURL{
		id:   id,
		link: link,
	}
}

func (a *GameURL) GetID() values.GameURLID {
	return a.id
}

func (a *GameURL) GetLink() values.GameURLLink {
	return a.link
}
