package gorm2

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/traPtitech/trap-collection-server/src/config"
	"github.com/traPtitech/trap-collection-server/src/config/mock"
)

var testDB *DB

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
			config.AutoRemove = true
			config.RestartPolicy = docker.RestartPolicy{
				Name: "no",
			}
		},
	)
	if err != nil {
		panic(fmt.Sprintf("Could not create container: %s", err))
	}

	portStr := resource.GetPort("3306/tcp")
	port, err := strconv.Atoi(portStr)
	if err != nil {
		panic(fmt.Sprintf("port is invalid: %s", err))
	}

	// 他のテストでは*testing.Tを使っているが、*testing.Mは使えないので、勝手に実装
	ctrl := gomock.NewController(&reporter{})
	mockAppConf := mock.NewMockApp(ctrl)
	mockRepositoryConf := mock.NewMockRepositoryGorm2(ctrl)

	// pool.Retryで繰り返すため、AnyTimesをつける
	mockAppConf.EXPECT().FeatureV2().Return(true).AnyTimes()
	mockAppConf.EXPECT().Status().Return(config.AppStatusDevelopment, nil).AnyTimes()

	mockRepositoryConf.EXPECT().Database().Return(mysqlDatabase, nil).AnyTimes()
	mockRepositoryConf.EXPECT().Host().Return(mysqlHost, nil).AnyTimes()
	mockRepositoryConf.EXPECT().Password().Return(mysqlRootPassword, nil).AnyTimes()
	mockRepositoryConf.EXPECT().User().Return(mysqlUser, nil).AnyTimes()
	mockRepositoryConf.EXPECT().Port().Return(port, nil).AnyTimes()

	if err := pool.Retry(func() error {
		testDB, err = NewDB(mockAppConf, mockRepositoryConf)
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		panic(fmt.Sprintf("Could not connect to database: %s", err))
	}

	code := m.Run()

	// os.Exitはdeferを使わないのでここでやる。
	ctrl.Finish()
	if err = pool.Purge(resource); err != nil {
		panic(fmt.Sprintf("Could not remove the container: %s", err))
	}

	os.Exit(code)
}

// gomock.TestReporterを実装
type reporter struct{}

func (*reporter) Errorf(format string, args ...interface{}) {
	log.Println(fmt.Errorf(format, args...))
}

func (*reporter) Fatalf(format string, args ...interface{}) {
	log.Fatalf(format, args...)
}
