package gorm2

type GameV2 struct {
	db *DB
}

func NewGameV2(db *DB) *GameV2 {
	return &GameV2{
		db: db,
	}
}
