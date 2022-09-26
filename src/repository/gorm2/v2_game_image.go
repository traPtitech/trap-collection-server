package gorm2

type GameImageV2 struct {
	db *DB
}

func NewGameImageV2(db *DB) *GameImageV2 {
	return &GameImageV2{
		db: db,
	}
}
