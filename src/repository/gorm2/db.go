package gorm2

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"

	"github.com/traPtitech/trap-collection-server/pkg/common"
	pkgContext "github.com/traPtitech/trap-collection-server/pkg/context"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type DB struct {
	db *gorm.DB
}

func NewDB(isProduction common.IsProduction) (*DB, error) {
	user, ok := os.LookupEnv("DB_USERNAME")
	if !ok {
		return nil, errors.New("DB_USERNAME is not set")
	}

	pass, ok := os.LookupEnv("DB_PASSWORD")
	if !ok {
		return nil, errors.New("DB_PASSWORD is not set")
	}

	host, ok := os.LookupEnv("DB_HOSTNAME")
	if !ok {
		return nil, errors.New("DB_HOSTNAME is not set")
	}

	port, ok := os.LookupEnv("DB_PORT")
	if !ok {
		return nil, errors.New("DB_PORT is not set")
	}

	database, ok := os.LookupEnv("DB_DATABASE")
	if !ok {
		return nil, errors.New("DB_DATABASE is not set")
	}

	var logLevel logger.LogLevel
	if isProduction {
		logLevel = logger.Silent
	} else {
		logLevel = logger.Info
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&loc=Asia%%2FTokyo&charset=utf8mb4",
		user,
		pass,
		host,
		port,
		database,
	)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	db = db.Set("gorm:table_options", "ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci")
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	err = db.AutoMigrate(tables...)
	if err != nil {
		return nil, fmt.Errorf("failed to auto migrate: %w", err)
	}

	return &DB{
		db: db,
	}, nil
}

func (db *DB) Get() (*sql.DB, error) {
	return db.db.DB()
}

func (db *DB) Transaction(ctx context.Context, txOption *sql.TxOptions, fn func(ctx context.Context) error) error {
	fc := func(tx *gorm.DB) error {
		ctx = context.WithValue(ctx, pkgContext.DBKey, tx)

		err := fn(ctx)
		if err != nil {
			return err
		}

		return nil
	}

	if txOption == nil {
		err := db.db.Transaction(fc)
		if err != nil {
			return fmt.Errorf("failed in transaction: %w", err)
		}
	} else {
		err := db.db.Transaction(fc, txOption)
		if err != nil {
			return fmt.Errorf("failed in transaction: %w", err)
		}
	}

	return nil
}

func (db *DB) getDB(ctx context.Context) (*gorm.DB, error) {
	iDB := ctx.Value(pkgContext.DBKey)
	if iDB == nil {
		return db.db.WithContext(ctx), nil
	}

	gormDB, ok := iDB.(*gorm.DB)
	if !ok {
		return nil, errors.New("failed to get gorm.DB")
	}

	return gormDB.WithContext(ctx), nil
}
