package s3

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/stretchr/testify/assert"
	v1 "github.com/traPtitech/trap-collection-server/src/config/v1"
	"github.com/traPtitech/trap-collection-server/src/storage"
)

var testClient *Client

func (c *Client) createBucket() error {
	_, err := c.client.CreateBucket(context.TODO(), &s3.CreateBucketInput{
		Bucket: &c.bucket,
	})
	if err != nil {
		return fmt.Errorf("failed to create bucket: %w", err)
	}

	return nil
}

func TestMain(m *testing.M) {
	conf := v1.NewStorageS3()

	var err error
	testClient, err = NewClient(conf)
	if err != nil {
		fmt.Printf("failed to create client: %v", err)
		os.Exit(1)
	}

	err = testClient.createBucket()
	if err != nil {
		fmt.Printf("failed to create bucket: %v", err)
		os.Exit(1)
	}

	code := m.Run()

	os.Exit(code)
}

func TestSaveFile(t *testing.T) {
	ctx := context.Background()

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
				_, err := testClient.client.PutObject(ctx, &s3.PutObjectInput{
					Bucket: &testClient.bucket,
					Key:    &testCase.name,
					Body:   bytes.NewReader([]byte{0}),
				}, s3.WithAPIOptions(
					v4.SwapComputePayloadSHA256ForUnsignedPayloadMiddleware,
				))
				if err != nil {
					t.Fatalf("failed to put object: %v", err)
				}
			}

			err := testClient.saveFile(
				ctx,
				testCase.name,
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
			if err != nil || testCase.isErr {
				return
			}

			objects, err := testClient.client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
				Bucket: &testClient.bucket,
				Prefix: &testCase.name,
			})
			if err != nil {
				t.Fatalf("failed to get object: %v", err)
			}

			exist := false
			for _, object := range objects.Contents {
				if object.Key != nil && *object.Key == testCase.name {
					exist = true
					break
				}
			}
			if !exist {
				t.Fatal("object not found")
			}

			result, err := testClient.client.GetObject(ctx, &s3.GetObjectInput{
				Bucket: &testClient.bucket,
				Key:    &testCase.name,
			})
			if err != nil {
				t.Fatalf("failed to get object: %v", err)
			}

			actualBytes, err := io.ReadAll(result.Body)
			if err != nil {
				t.Fatalf("failed to read body: %v", err)
			}

			assert.Equal(t, expectBytes, actualBytes)
		})
	}
}

func TestCreateTempURL(t *testing.T) {
	ctx := context.Background()

	type test struct {
		description string
		name        string
		isFileExist bool
		expiresIn   time.Duration
		content     *bytes.Buffer
		isErr       bool
		err         error
	}

	testCases := []test{
		{
			description: "特に問題ないので取得できる",
			isFileExist: true,
			name:        "a",
			expiresIn:   2 * time.Second,
			content:     bytes.NewBufferString("a"),
		},
		{
			description: "ファイルが存在しないのでErrNotFound",
			name:        "h",
			isFileExist: false,
			isErr:       true,
			err:         storage.ErrNotFound,
		},
		{
			description: "サイズが大きくても取得できる",
			name:        "d",
			isFileExist: true,
			expiresIn:   2 * time.Second,
			content:     bytes.NewBufferString(strings.Repeat("d", 1024*1024*10)),
		},
		{
			description: "名前に/が含まれていても取得できる",
			isFileExist: true,
			name:        "f/g",
			expiresIn:   2 * time.Second,
			content:     bytes.NewBufferString("f"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			var expectBytes []byte
			if testCase.isFileExist {
				expectBytes = testCase.content.Bytes()

				_, err := testClient.client.PutObject(ctx, &s3.PutObjectInput{
					Bucket: &testClient.bucket,
					Key:    &testCase.name,
					Body:   testCase.content,
				}, s3.WithAPIOptions(
					v4.SwapComputePayloadSHA256ForUnsignedPayloadMiddleware,
				))
				if err != nil {
					t.Fatalf("failed to put object: %v", err)
				}
			}

			buf := bytes.NewBuffer(nil)

			tmpURL, err := testClient.createTempURL(ctx, testCase.name, testCase.expiresIn)

			if testCase.isErr {
				if testCase.err == nil {
					assert.Error(t, err)
				} else if !errors.Is(err, testCase.err) {
					t.Errorf("error must be %v, but actual is %v", testCase.err, err)
				}
			} else {
				assert.NoError(t, err)
			}
			if err != nil || testCase.isErr {
				return
			}

			res, err := http.Get((*url.URL)(tmpURL).String())
			if err != nil {
				t.Fatalf("failed to get file: %v", err)
			}

			_, err = buf.ReadFrom(res.Body)
			if err != nil {
				t.Fatalf("failed to read file: %v", err)
			}

			assert.Equal(t, expectBytes, buf.Bytes())

			time.Sleep(testCase.expiresIn)

			res, err = http.Get((*url.URL)(tmpURL).String())
			if err != nil {
				t.Fatalf("failed to get file: %v", err)
			}

			assert.Equal(t, http.StatusForbidden, res.StatusCode)
		})
	}
}

func (c *Client) loadFile(ctx context.Context, name string, w io.Writer) error {
	res, err := c.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &c.bucket,
		Key:    &name,
	})
	var awsErr *types.NoSuchKey
	if err != nil && errors.As(err, &awsErr) {
		return fmt.Errorf("failed to get object: %w", storage.ErrNotFound)
	}
	if err != nil {
		return fmt.Errorf("failed to get object: %w", err)
	}

	_, err = io.Copy(w, res.Body)
	if err != nil {
		return fmt.Errorf("failed to copy: %w", err)
	}

	return nil
}
