package swift

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/ncw/swift/v2"
	"github.com/traPtitech/trap-collection-server/pkg/common"
)

const (
	cacheDuration = 7 * 24 * time.Hour
)

type Client struct {
	connection    *swift.Connection
	containerName string
	tmpURLKey     string
}

func NewClient(
	authURL common.SwiftAuthURL,
	userName common.SwiftUserName,
	password common.SwiftPassword,
	tennantName common.SwiftTenantName,
	tennantID common.SwiftTenantID,
	containerName common.SwiftContainer,
	tmpURLKey common.SwiftTmpURLKey,
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
		tmpURLKey:     string(tmpURLKey),
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

	_, err = io.Copy(f, content)
	if err != nil {
		return fmt.Errorf("failed to copy content: %w", err)
	}

	return nil
}

var (
	ErrNotFound = fmt.Errorf("not found")
)

func (c *Client) loadFile(ctx context.Context, name string, w io.Writer) error {
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

func (c *Client) createTempURL(ctx context.Context, name string, expires time.Duration) (*url.URL, error) {
	_, _, err := c.connection.Object(ctx, c.containerName, name)
	if errors.Is(err, swift.ObjectNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get object: %w", err)
	}

	strURL := c.connection.ObjectTempUrl(c.containerName, name, c.tmpURLKey, http.MethodGet, time.Now().Add(expires))

	tmpURL, err := url.Parse(strURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse url: %w", err)
	}

	return tmpURL, nil
}
