package schema

import (
	"context"
	"fmt"

	"ariga.io/atlas-go-sdk/atlasexec"
	"github.com/traPtitech/trap-collection-server/migrations"
	"github.com/traPtitech/trap-collection-server/src/config"
	"gorm.io/gorm"
)

func buildAtlasURL(conf config.RepositoryGorm2) (string, error) {
	user, err := conf.User()
	if err != nil {
		return "", fmt.Errorf("failed to get user: %w", err)
	}

	pass, err := conf.Password()
	if err != nil {
		return "", fmt.Errorf("failed to get password: %w", err)
	}

	host, err := conf.Host()
	if err != nil {
		return "", fmt.Errorf("failed to get host: %w", err)
	}

	port, err := conf.Port()
	if err != nil {
		return "", fmt.Errorf("failed to get port: %w", err)
	}

	database, err := conf.Database()
	if err != nil {
		return "", fmt.Errorf("failed to get database: %w", err)
	}

	url := fmt.Sprintf("maria://%s:%s@%s:%d/%s",
		user,
		pass,
		host,
		port,
		database,
	)

	return url, nil
}

func Migrate(ctx context.Context, dbConf config.RepositoryGorm2, migrationConf config.Migration, db *gorm.DB) error {
	workdir, err := atlasexec.NewWorkingDir(
		atlasexec.WithMigrations(migrations.MigrationDir),
	)
	if err != nil {
		return fmt.Errorf("create atlas working dir: %w", err)
	}
	defer workdir.Close()

	client, err := atlasexec.NewClient(workdir.Path(), "atlas")
	if err != nil {
		return fmt.Errorf("create atlas client: %w", err)
	}

	url, err := buildAtlasURL(dbConf)
	if err != nil {
		return fmt.Errorf("build URL: %w", err)
	}

	var baseline string
	emptyDB, err := migrationConf.EmptyDB()
	if err != nil {
		return fmt.Errorf("get emptyDB: %w", err)
	}
	if !emptyDB {
		baseline, err = migrationConf.Baseline()
		if err != nil {
			return fmt.Errorf("get baseline: %w", err)
		}
	}

	params := &atlasexec.MigrateApplyParams{
		URL:             url,
		BaselineVersion: baseline,
	}

	res, err := client.MigrateApply(ctx, params)
	if err != nil {
		return fmt.Errorf("apply migration: %w", err)
	}

	db.Logger.Info(ctx, "migrate apply result: %+v", res)

	return nil
}
