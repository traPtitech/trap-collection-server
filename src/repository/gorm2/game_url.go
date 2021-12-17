package gorm2

type GameURL struct {
	db *DB
}

func NewGameURL(db *DB) *GameURL {
	return &GameURL{
		db: db,
	}
}
