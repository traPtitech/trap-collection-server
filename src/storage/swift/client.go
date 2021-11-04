package swift

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/ncw/swift/v2"
	"github.com/traPtitech/trap-collection-server/pkg/common"
	"gopkg.in/djherbis/fscache.v0"
)

const (
	cacheDuration = 7 * 24 * time.Hour
)

type Client struct {
	connection    *swift.Connection
	containerName string
	cache         fscache.Cache
}

func NewClient(
	authURL common.SwiftAuthURL,
	userName common.SwiftUserName,
	password common.SwiftPassword,
	tennantID common.SwiftTenantID,
	containerName common.SwiftContainer,
	cacheDirectory common.FilePath,
) (*Client, error) {
	ctx := context.Background()

	connection, err := setupSwift(
		ctx,
		(*url.URL)(authURL),
		string(userName),
		string(password),
		string(tennantID),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to setup swift: %w", err)
	}

	cache, err := fscache.New(string(cacheDirectory), 0755, cacheDuration)
	if err != nil {
		return nil, fmt.Errorf("failed to setup cache: %w", err)
	}

	return &Client{
		connection:    connection,
		containerName: string(containerName),
		cache:         cache,
	}, nil
}

func setupSwift(
	ctx context.Context,
	authURL *url.URL,
	userName string,
	password string,
	tennantID string,
) (*swift.Connection, error) {
	c := &swift.Connection{
		UserName: userName,
		ApiKey:   password,
		AuthUrl:  authURL.String(),
		Tenant:   tennantID,
	}

	err := c.Authenticate(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate: %w", err)
	}

	return c, nil
}
