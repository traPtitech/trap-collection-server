package gorm2

type GameVideoV2 struct {
	db *DB
}

func NewGameVideoV2(db *DB) *GameVideoV2 {
	return &GameVideoV2{
		db: db,
	}
}
