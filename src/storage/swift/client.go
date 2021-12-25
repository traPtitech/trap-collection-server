package swift

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/ncw/swift/v2"
	"github.com/traPtitech/trap-collection-server/pkg/common"
	"golang.org/x/sync/errgroup"
)

type Client struct {
	connection    *swift.Connection
	containerName string
	cache         *Cache
}

func NewClient(
	authURL common.SwiftAuthURL,
	userName common.SwiftUserName,
	password common.SwiftPassword,
	tennantName common.SwiftTenantName,
	tennantID common.SwiftTenantID,
	containerName common.SwiftContainer,
	cache *Cache,
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

	eg, _ := errgroup.WithContext(ctx)
	pr, pw := io.Pipe()

	mw := io.MultiWriter(f, pw)

	eg.Go(func() error {
		defer pr.Close()

		cacheName := strings.ReplaceAll(name, "/", "_")

		err := c.cache.save(cacheName, pr)
		if err != nil {
			return fmt.Errorf("failed to copy content: %w", err)
		}

		return nil
	})

	eg.Go(func() error {
		defer pw.Close()

		_, err = io.Copy(mw, content)
		if err != nil {
			return fmt.Errorf("failed to copy content: %w", err)
		}

		return nil
	})

	err = eg.Wait()
	if err != nil {
		return fmt.Errorf("failed to save file: %w", err)
	}

	return nil
}

var (
	ErrNotFound = fmt.Errorf("not found")
)

func (c *Client) loadFile(ctx context.Context, name string, w io.Writer) (bool, error) {
	cacheName := strings.ReplaceAll(name, "/", "_")

	hit, err := c.cache.load(cacheName, w)
	if err != nil {
		return false, fmt.Errorf("failed to load cache: %w", err)
	}

	if hit {
		return true, nil
	}

	_, _, err = c.connection.Object(ctx, c.containerName, name)
	if errors.Is(err, swift.ObjectNotFound) {
		return false, ErrNotFound
	}
	if err != nil {
		return false, fmt.Errorf("failed to get object: %w", err)
	}

	eg, ctx := errgroup.WithContext(ctx)
	pr, pw := io.Pipe()

	mw := io.MultiWriter(w, pw)

	eg.Go(func() error {
		defer pr.Close()

		err := c.cache.save(cacheName, pr)
		if err != nil {
			return fmt.Errorf("failed to save cache: %w", err)
		}

		return nil
	})

	eg.Go(func() error {
		defer pw.Close()

		_, err = c.connection.ObjectGet(
			ctx,
			c.containerName,
			name,
			mw,
			true,
			nil,
		)
		if err != nil {
			return fmt.Errorf("failed to get object: %w", err)
		}

		return nil
	})

	err = eg.Wait()
	if err != nil {
		return false, fmt.Errorf("failed to get object: %w", err)
	}

	return false, nil
}
