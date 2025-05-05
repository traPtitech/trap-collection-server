package v1

import (
	"errors"
	"os"
	"strconv"

	"github.com/traPtitech/trap-collection-server/src/config"
)

type Migration struct{}

var _ config.Migration = (*Migration)(nil)

func NewMigration() *Migration {
	return &Migration{}
}

func (m *Migration) EmptyDB() (bool, error) {
	emptyDBStr, ok := os.LookupEnv(envKeyMigrationEmptyDB)
	if !ok {
		return false, errors.New("MIGRATION_EMPTY_DB is not set")
	}

	emptyDB, err := strconv.ParseBool(emptyDBStr)
	if err != nil {
		return false, errors.New("MIGRATION_EMPTY_DB is not a boolean")
	}
	return emptyDB, nil
}

func (m *Migration) Baseline() (string, error) {
	baseline, ok := os.LookupEnv(envKeyMigrationBaseline)
	if !ok {
		return "", errors.New("MIGRATION_BASELINE is not set")
	}

	return baseline, nil
}
