package swift

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/ncw/swift/v2"
	"github.com/ncw/swift/v2/swifttest"
	"github.com/stretchr/testify/assert"
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

func TestSaveFile(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	client, err := newTestClient(
		ctx,
		common.SwiftContainer("save_file"),
		common.FilePath("save_file"),
	)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	defer func() {
		err := os.RemoveAll("save_file")
		if err != nil {
			t.Fatalf("failed to remove directory: %v", err)
		}
	}()

	type test struct {
		description        string
		name               string
		contentType        string
		hash               string
		content            *bytes.Buffer
		isAlreadyFileExist bool
		isErr              bool
		err                error
	}

	testCases := []test{
		{
			description: "特に問題ないので保存できる",
			name:        "a",
			contentType: "text/plain",
			content:     bytes.NewBufferString("a"),
		},
		{
			description:        "ファイルが既に存在するのでErrAlreadyExists",
			name:               "b",
			contentType:        "text/plain",
			content:            bytes.NewBufferString("b"),
			isAlreadyFileExist: true,
			isErr:              true,
			err:                ErrAlreadyExists,
		},
		{
			description: "hashが設定されていても保存できる",
			name:        "c",
			contentType: "text/plain",
			content:     bytes.NewBufferString("c"),
			hash:        "4a8a08f09d37b73795649038408b5f33",
		},
		{
			description: "hashが誤っているのでエラー",
			name:        "d",
			contentType: "text/plain",
			content:     bytes.NewBufferString("d"),
			// 正しい値: 8277e0910d750195b448797616e091ad
			hash:  "invalid",
			isErr: true,
		},
		{
			description: "サイズが大きくても保存できる",
			name:        "e",
			contentType: "text/plain",
			content:     bytes.NewBufferString(strings.Repeat("e", 1024*1024*10)),
		},
		{
			description: "ファイル名に/が含まれていても保存できる",
			name:        "f/g",
			contentType: "text/plain",
			content:     bytes.NewBufferString("f"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			defer func() {
				// 確実にキャッシュが残るように、キャッシュを消す
				err := client.cache.Clean()
				if err != nil {
					t.Fatalf("failed to clean cache: %v", err)
				}
			}()

			expectBytes := testCase.content.Bytes()

			if testCase.isAlreadyFileExist {
				err := client.connection.ObjectPutBytes(ctx, client.containerName, testCase.name, []byte{0}, testCase.contentType)
				if err != nil {
					t.Fatalf("failed to put object: %v", err)
				}
			}

			err := client.saveFile(
				ctx,
				testCase.name,
				testCase.contentType,
				testCase.hash,
				testCase.content,
			)

			if testCase.isErr {
				if testCase.err == nil {
					assert.Error(t, err)
				} else if !errors.Is(err, testCase.err) {
					t.Errorf("error must be %v, but actual is %v", testCase.err, err)
				}
			} else {
				assert.NoError(t, err)
			}
			if err != nil {
				return
			}

			_, _, err = client.connection.Object(ctx, client.containerName, testCase.name)
			if errors.Is(err, swift.ObjectNotFound) {
				t.Fatalf("object not found: %v", err)
			}
			if err != nil {
				t.Fatalf("failed to get object: %v", err)
			}

			actualBytes, err := client.connection.ObjectGetBytes(ctx, client.containerName, testCase.name)
			if err != nil {
				t.Fatalf("failed to get object: %v", err)
			}

			assert.Equal(t, expectBytes, actualBytes)

			assert.True(t, client.cache.Exists(testCase.name))

			r, _, err := client.cache.Get(testCase.name)
			if err != nil {
				t.Fatalf("failed to get cache: %v", err)
			}
			defer r.Close()

			actualBytes, err = io.ReadAll(r)
			if err != nil {
				t.Fatalf("failed to read cache: %v", err)
			}

			assert.Equal(t, expectBytes, actualBytes)
		})
	}
}
