package migrations

import (
	"embed"
)

//go:embed *
var MigrationDir embed.FS
