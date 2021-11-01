package gorm2

type Game struct {
	db *DB
}

func NewGame(db *DB) *Game {
	return &Game{
		db: db,
	}
}
