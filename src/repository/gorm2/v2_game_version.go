package gorm2

type GameVersionV2 struct {
	db *DB
}

func NewGameVersionV2(db *DB) *GameVersionV2 {
	return &GameVersionV2{
		db: db,
	}
}
