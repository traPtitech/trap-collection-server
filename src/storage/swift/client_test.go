package swift

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"testing"

	"github.com/ncw/swift/v2/swifttest"
	"github.com/traPtitech/trap-collection-server/pkg/common"
)

var testServer *swifttest.SwiftServer

func TestMain(m *testing.M) {
	var err error
	testServer, err = swifttest.NewSwiftServer("")
	if err != nil {
		panic(err)
	}

	code := m.Run()

	testServer.Close()

	os.Exit(code)
}

func newTestClient(ctx context.Context, containerName common.SwiftContainer, cacheDirectory common.FilePath) (*Client, error) {
	authURL, err := url.Parse(testServer.AuthURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse auth url: %w", err)
	}

	client, err := NewClient(
		common.SwiftAuthURL(authURL),
		common.SwiftUserName(swifttest.TEST_ACCOUNT),
		common.SwiftPassword(swifttest.TEST_ACCOUNT),
		// テスト用サーバーはv1での認証なので、tenantIDは必要ない
		common.SwiftTenantID(""),
		common.SwiftContainer(containerName),
		cacheDirectory,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	err = client.connection.ContainerCreate(ctx, string(containerName), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create container: %w", err)
	}

	return client, err
}
