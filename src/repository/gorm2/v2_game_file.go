package gorm2

type GameFileV2 struct {
	db *DB
}

func NewGameFileV2(db *DB) *GameFileV2 {
	return &GameFileV2{
		db: db,
	}
}
