package gorm2

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"testing"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/traPtitech/trap-collection-server/src/config"
)

var testDB *DB

const (
	mysqlRootPassword = "pass"
	mysqlDatabase     = "trap_collection"
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

	if err := pool.Retry(func() error {
		testDB, err = NewDB(&testAppConfig{}, &testRepositoryConfig{resource.GetPort("3306/tcp")})
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		panic(fmt.Sprintf("Could not connect to database: %s", err))
	}

	code := m.Run()

	if err = pool.Purge(resource); err != nil {
		panic(fmt.Sprintf("Could not remove the container: %s", err))
	}

	os.Exit(code)
}

// テスト用のAppConfig
// config.Appを実装
type testAppConfig struct{}

func (*testAppConfig) Status() (config.AppStatus, error) {
	return config.AppStatusDevelopment, nil
}
func (*testAppConfig) FeatureV2() bool {
	return true
}
func (*testAppConfig) FeatureV1Write() bool {
	return true
}

// テスト用のRepositoryConfig
// config.RepositoryGorm2を実装
type testRepositoryConfig struct {
	port string
}

func (*testRepositoryConfig) User() (string, error) {
	return "root", nil
}
func (*testRepositoryConfig) Password() (string, error) {
	return "pass", nil
}
func (t *testRepositoryConfig) Host() (string, error) {
	return "localhost", nil
}
func (t *testRepositoryConfig) Port() (int, error) {
	port, err := strconv.Atoi(t.port)
	if err != nil {
		return 0, fmt.Errorf("failed to convert port to int: %w", err)
	}
	return port, nil
}
func (*testRepositoryConfig) Database() (string, error) {
	return "trap_collection", nil
}
