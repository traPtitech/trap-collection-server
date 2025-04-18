package swift

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/storage"
	"go.uber.org/mock/gomock"
)

func TestSaveGameFile(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)

	client, err := newTestClient(ctx, ctrl, "save_game_file")
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	type test struct {
		description string
		fileID      values.GameFileID
		reader      *bytes.Buffer
		isFileExist bool
		isErr       bool
		err         error
	}

	testCases := []test{
		{
			description: "特に問題ないのでエラーなし",
			fileID:      values.NewGameFileID(),
			reader:      bytes.NewBufferString("test"),
		},
		{
			description: "ファイルが存在するのでErrAlreadyExists",
			fileID:      values.NewGameFileID(),
			reader:      bytes.NewBufferString("test"),
			isFileExist: true,
			isErr:       true,
			err:         storage.ErrAlreadyExists,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			gameFileStorage := NewGameFile(client)

			if testCase.isFileExist {
				err := client.saveFile(
					ctx,
					fmt.Sprintf("files/%s", uuid.UUID(testCase.fileID).String()),
					"text/plain",
					"",
					strings.NewReader(""),
				)
				if err != nil {
					t.Fatalf("failed to create file: %v", err)
				}
			}

			expectBytes := testCase.reader.Bytes()

			err := gameFileStorage.SaveGameFile(ctx, testCase.reader, testCase.fileID)

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

			buf := bytes.NewBuffer(nil)
			err = client.loadFile(ctx, fmt.Sprintf("files/%s", uuid.UUID(testCase.fileID).String()), buf)
			if err != nil {
				t.Fatalf("failed to load file: %v", err)
			}

			assert.Equal(t, expectBytes, buf.Bytes())
		})
	}
}

func TestFileGetTempURL(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)

	client, err := newTestClient(ctx, ctrl, "get_game_file")
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	gameFileStorage := NewGameFile(client)

	type test struct {
		description string
		file        *domain.GameFile
		buf         *bytes.Buffer
		isFileExist bool
		isErr       bool
		err         error
	}

	testCases := []test{
		{
			description: "特に問題ないのでエラーなし",
			file: domain.NewGameFile(
				values.NewGameFileID(),
				values.GameFileTypeJar,
				"path/to/game.jar",
				values.NewGameFileHashFromBytes([]byte{0x09, 0x8f, 0x6b, 0xcd, 0x46, 0x21, 0xd3, 0x73, 0xca, 0xde, 0x4e, 0x83, 0x26, 0x27, 0xb4, 0xf6}),
				time.Now(),
			),
			buf:         bytes.NewBufferString("test"),
			isFileExist: true,
		},
		{
			description: "ファイルが存在しないのでErrNotFound",
			file: domain.NewGameFile(
				values.NewGameFileID(),
				values.GameFileTypeJar,
				"path/to/game.jar",
				values.NewGameFileHashFromBytes([]byte{0x09, 0x8f, 0x6b, 0xcd, 0x46, 0x21, 0xd3, 0x73, 0xca, 0xde, 0x4e, 0x83, 0x26, 0x27, 0xb4, 0xf6}),
				time.Now(),
			),
			buf:   bytes.NewBufferString("test"),
			isErr: true,
			err:   storage.ErrNotFound,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			expectBytes := testCase.buf.Bytes()

			if testCase.isFileExist {
				err := client.saveFile(
					ctx,
					fmt.Sprintf("files/%s", uuid.UUID(testCase.file.GetID()).String()),
					"",
					"",
					testCase.buf,
				)
				if err != nil {
					t.Fatalf("failed to create file: %v", err)
				}
			}

			buf := bytes.NewBuffer(nil)
			tmpURL, err := gameFileStorage.GetTempURL(ctx, testCase.file, time.Second)

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

			res, err := http.Get((*url.URL)(tmpURL).String())
			if err != nil {
				t.Fatalf("failed to get file: %v", err)
			}

			_, err = buf.ReadFrom(res.Body)
			if err != nil {
				t.Fatalf("failed to read file: %v", err)
			}

			assert.Equal(t, expectBytes, buf.Bytes())
		})
	}
}

func TestFileKey(t *testing.T) {
	t.Parallel()

	// clientは使わないのでnilでOK
	gameFileStorage := NewGameFile(nil)

	loopNum := 100

	for i := 0; i < loopNum; i++ {
		fileID := values.NewGameFileID()

		key := gameFileStorage.fileKey(fileID)

		assert.Equal(t, fmt.Sprintf("files/%s", uuid.UUID(fileID).String()), key)
	}
}
