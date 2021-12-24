package swift

import (
	"context"
	"errors"
	"fmt"
	"io"
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
	tennantName common.SwiftTenantName,
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
		string(tennantName),
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
	tennantName string,
	tennantID string,
) (*swift.Connection, error) {
	c := &swift.Connection{
		UserName: userName,
		ApiKey:   password,
		AuthUrl:  authURL.String(),
		Tenant:   tennantName,
		TenantId: tennantID,
	}

	err := c.Authenticate(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate: %w", err)
	}

	return c, nil
}

var (
	ErrAlreadyExists = fmt.Errorf("already exists")
)

func (c *Client) saveFile(
	ctx context.Context,
	name string,
	contentType string,
	hash string,
	content io.Reader,
) error {
	_, _, err := c.connection.Object(ctx, c.containerName, name)
	if err == nil {
		return ErrAlreadyExists
	}
	if err != nil && !errors.Is(err, swift.ObjectNotFound) {
		return fmt.Errorf("failed to get object: %w", err)
	}

	var checksum string
	if len(hash) == 0 {
		checksum = ""
	} else {
		checksum = hash
	}

	f, err := c.connection.ObjectCreate(
		ctx,
		c.containerName,
		name,
		true,
		checksum,
		contentType,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to create object: %w", err)
	}
	defer f.Close()

	r, w, err := c.cache.Get(name)
	if err != nil {
		return fmt.Errorf("failed to get cache: %w", err)
	}
	defer w.Close()

	err = r.Close()
	if err != nil {
		return fmt.Errorf("failed to close cache: %w", err)
	}

	mw := io.MultiWriter(f, w)

	_, err = io.Copy(mw, content)
	if err != nil {
		return fmt.Errorf("failed to copy content: %w", err)
	}

	return nil
}

var (
	ErrNotFound = fmt.Errorf("not found")
)

func (c *Client) loadFile(ctx context.Context, name string, w io.Writer) error {
	if c.cache.Exists(name) {
		r, _, err := c.cache.Get(name)
		if err != nil {
			return fmt.Errorf("failed to get cache: %w", err)
		}
		defer r.Close()

		_, err = io.Copy(w, r)
		if err != nil {
			return fmt.Errorf("failed to copy cache: %w", err)
		}

		return nil
	}

	_, _, err := c.connection.Object(ctx, c.containerName, name)
	if errors.Is(err, swift.ObjectNotFound) {
		return ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("failed to get object: %w", err)
	}

	_, err = c.connection.ObjectGet(
		ctx,
		c.containerName,
		name,
		w,
		true,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to get object: %w", err)
	}

	return nil
}
