package gorm2

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	pkgContext "github.com/traPtitech/trap-collection-server/pkg/context"
	"github.com/traPtitech/trap-collection-server/src/config"
	"github.com/traPtitech/trap-collection-server/src/repository"
	"github.com/traPtitech/trap-collection-server/src/repository/gorm2/schema"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
	"gorm.io/plugin/prometheus"
)

type DB struct {
	db *gorm.DB
}

func NewDB(appConf config.App, conf config.RepositoryGorm2, migrationConf config.Migration) (*DB, error) {
	appStatus, err := appConf.Status()
	if err != nil {
		return nil, fmt.Errorf("failed to get app status: %w", err)
	}

	var logLevel logger.LogLevel
	switch appStatus {
	case config.AppStatusProduction:
		logLevel = logger.Silent
	case config.AppStatusDevelopment:
		logLevel = logger.Info
	default:
		return nil, errors.New("invalid app status")
	}

	user, err := conf.User()
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	pass, err := conf.Password()
	if err != nil {
		return nil, fmt.Errorf("failed to get password: %w", err)
	}

	host, err := conf.Host()
	if err != nil {
		return nil, fmt.Errorf("failed to get host: %w", err)
	}

	port, err := conf.Port()
	if err != nil {
		return nil, fmt.Errorf("failed to get port: %w", err)
	}

	database, err := conf.Database()
	if err != nil {
		return nil, fmt.Errorf("failed to get database: %w", err)
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&loc=Asia%%2FTokyo&charset=utf8mb4",
		user,
		pass,
		host,
		port,
		database,
	)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	db = db.Set("gorm:table_options", "ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci")

	err = schema.Migrate(context.Background(), conf, migrationConf, db)
	if err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}

	// err = migrate.Migrate(db, appConf.FeatureV2())
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to migrate: %w", err)
	// }

	var collector prometheus.MetricsCollector
	if appConf.FeatureV2() {
		collector = &MetricsCollectorV2{}
	} else {
		return nil, fmt.Errorf("only v2 is allowed")
	}

	err = db.Use(prometheus.New(prometheus.Config{
		DBName:          "trap_collection",
		RefreshInterval: 15,
		MetricsCollector: []prometheus.MetricsCollector{
			collector,
		},
	}))
	if err != nil {
		return nil, fmt.Errorf("failed to use prometheus plugin: %w", err)
	}

	return &DB{
		db: db,
	}, nil
}

func (db *DB) Close() error {
	sqldb, err := db.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB: %w", err)
	}

	err = sqldb.Close()
	if err != nil {
		return fmt.Errorf("failed to close sql.DB: %w", err)
	}

	return nil
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

func (db *DB) setLock(gormDB *gorm.DB, lockType repository.LockType) (*gorm.DB, error) {
	switch lockType {
	case repository.LockTypeRecord:
		gormDB = gormDB.Clauses(clause.Locking{Strength: "UPDATE"})
	case repository.LockTypeNone:
	default:
		return nil, errors.New("invalid lock type")
	}

	return gormDB, nil
}
