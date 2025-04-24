package schema

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"testing"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/traPtitech/trap-collection-server/src/config"
	"github.com/traPtitech/trap-collection-server/src/config/mock"
	"go.uber.org/mock/gomock"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var testDB *gorm.DB

const (
	mysqlRootPassword = "pass"
	mysqlDatabase     = "trap_collection"
	mysqlHost         = "localhost"
	mysqlUser         = "root"
	timezone          = "Asia/Tokyo"
)

func TestMain(m *testing.M) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		panic(fmt.Sprintf("Could not create pool: %s", err))
	}

	err = pool.Client.Ping()
	if err != nil {
		panic(fmt.Sprintf("Failed to ping: %s", err))
	}

	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "mariadb",
		Tag:        "10.6.4",
		Env: []string{
			"MYSQL_ROOT_PASSWORD=" + mysqlRootPassword,
			"MYSQL_DATABASE=" + mysqlDatabase,
			"TZ=" + timezone,
		},
	},
		func(config *docker.HostConfig) {
			config.AutoRemove = false
			config.RestartPolicy = docker.RestartPolicy{
				Name: "no",
			}
		},
	)
	if err != nil {
		panic(fmt.Sprintf("Could not create container: %s", err))
	}

	defer func() {
		if err = pool.Purge(resource); err != nil {
			log.Printf("Could not remove the container: %s", err)
		}
	}()

	portStr := resource.GetPort("3306/tcp")
	port, err := strconv.Atoi(portStr)
	if err != nil {
		panic(fmt.Sprintf("port is invalid: %s", err))
	}

	// 他のテストでは*testing.Tを使っているが、*testing.Mは使えないので、勝手に実装
	ctrl := gomock.NewController(&reporter{})
	defer ctrl.Finish()
	mockAppConf := mock.NewMockApp(ctrl)
	mockRepositoryConf := mock.NewMockRepositoryGorm2(ctrl)
	mockMigrationConf := mock.NewMockMigration(ctrl)

	// pool.Retryで繰り返すため、AnyTimesをつける
	mockAppConf.EXPECT().FeatureV2().Return(true).AnyTimes()
	mockAppConf.EXPECT().Status().Return(config.AppStatusDevelopment, nil).AnyTimes()

	mockRepositoryConf.EXPECT().Database().Return(mysqlDatabase, nil).AnyTimes()
	mockRepositoryConf.EXPECT().Host().Return(mysqlHost, nil).AnyTimes()
	mockRepositoryConf.EXPECT().Password().Return(mysqlRootPassword, nil).AnyTimes()
	mockRepositoryConf.EXPECT().User().Return(mysqlUser, nil).AnyTimes()
	mockRepositoryConf.EXPECT().Port().Return(port, nil).AnyTimes()

	mockMigrationConf.EXPECT().EmptyDB().Return(true, nil).AnyTimes()

	err = pool.Retry(func() error {
		if err := setupTestDB(mockAppConf, mockRepositoryConf, mockMigrationConf); err != nil {
			return fmt.Errorf("setup test database: %w", err)
		}
		return nil
	})
	if err != nil {
		panic(fmt.Sprintf("Could not connect to database: %s", err))
	}

	m.Run()
}

func setupTestDB(appConf config.App, conf config.RepositoryGorm2, migrationConf config.Migration) error {

	appStatus, err := appConf.Status()
	if err != nil {
		return fmt.Errorf("failed to get app status: %w", err)
	}

	var logLevel logger.LogLevel
	switch appStatus {
	case config.AppStatusProduction:
		logLevel = logger.Silent
	case config.AppStatusDevelopment:
		logLevel = logger.Info
	default:
		return errors.New("invalid app status")
	}

	user, err := conf.User()
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	pass, err := conf.Password()
	if err != nil {
		return fmt.Errorf("failed to get password: %w", err)
	}

	host, err := conf.Host()
	if err != nil {
		return fmt.Errorf("failed to get host: %w", err)
	}

	port, err := conf.Port()
	if err != nil {
		return fmt.Errorf("failed to get port: %w", err)
	}

	database, err := conf.Database()
	if err != nil {
		return fmt.Errorf("failed to get database: %w", err)
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
		return fmt.Errorf("connect to database: %w", err)
	}
	db = db.Set("gorm:table_options", "ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci")

	testDB = db

	err = Migrate(context.Background(), conf, migrationConf, testDB)
	if err != nil {
		return fmt.Errorf("migrate database: %w", err)
	}

	return nil
}

// gomock.TestReporterを実装
type reporter struct{}

func (*reporter) Errorf(format string, args ...interface{}) {
	log.Println(fmt.Errorf(format, args...))
}

func (*reporter) Fatalf(format string, args ...interface{}) {
	log.Fatalf(format, args...)
}
