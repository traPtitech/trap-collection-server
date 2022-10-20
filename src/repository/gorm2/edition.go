package gorm2

type Edition struct {
	db *DB
}

func NewEdition(db *DB) *Edition {
	return &Edition{
		db: db,
	}
}
