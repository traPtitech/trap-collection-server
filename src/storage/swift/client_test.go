package swift

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/ncw/swift/v2"
	"github.com/ncw/swift/v2/swifttest"
	"github.com/stretchr/testify/assert"
	"github.com/traPtitech/trap-collection-server/src/config/mock"
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

func newTestClient(ctx context.Context, ctrl *gomock.Controller, containerName string) (*Client, error) {
	authURL, err := url.Parse(testServer.AuthURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse auth url: %w", err)
	}

	mockConf := mock.NewMockStorageSwift(ctrl)
	mockConf.
		EXPECT().
		AuthURL().
		Return(authURL, nil)
	mockConf.
		EXPECT().
		UserName().
		Return(swifttest.TEST_ACCOUNT, nil)
	mockConf.
		EXPECT().
		Password().
		Return(swifttest.TEST_ACCOUNT, nil)
	mockConf.
		EXPECT().
		TenantName().
		Return("", nil)
	mockConf.
		EXPECT().
		TenantID().
		Return("", nil)
	mockConf.
		EXPECT().
		Container().
		Return(containerName, nil)
	mockConf.
		EXPECT().
		TmpURLKey().
		Return("", nil)

	client, err := NewClient(mockConf)
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
	ctx := context.Background()
	ctrl := gomock.NewController(t)

	client, err := newTestClient(ctx, ctrl, "save_file")
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

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
		})
	}
}

func TestLoadFile(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)

	client, err := newTestClient(ctx, ctrl, "load_file")
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	type test struct {
		description string
		name        string
		isFileExist bool
		content     *bytes.Buffer
		isErr       bool
		err         error
	}

	testCases := []test{
		{
			description: "特に問題ないので取得できる",
			isFileExist: true,
			name:        "a",
			content:     bytes.NewBufferString("a"),
		},
		{
			description: "ファイルが存在しないのでErrNotFound",
			name:        "c",
			isFileExist: false,
			isErr:       true,
			err:         ErrNotFound,
		},
		{
			description: "サイズが大きくても取得できる",
			name:        "d",
			isFileExist: true,
			content:     bytes.NewBufferString(strings.Repeat("d", 1024*1024*10)),
		},
		{
			description: "名前に/が含まれていても取得できる",
			isFileExist: true,
			name:        "f/g",
			content:     bytes.NewBufferString("f"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			var expectBytes []byte
			if testCase.isFileExist {
				expectBytes = testCase.content.Bytes()

				_, err := client.connection.ObjectPut(
					ctx,
					client.containerName,
					testCase.name,
					testCase.content,
					true,
					"",
					"",
					nil,
				)
				if err != nil {
					t.Fatalf("failed to put object: %v", err)
				}
			}

			buf := bytes.NewBuffer(nil)

			err := client.loadFile(ctx, testCase.name, buf)

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

			assert.Equal(t, expectBytes, buf.Bytes())
		})
	}
}
