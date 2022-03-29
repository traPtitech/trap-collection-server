package local

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/traPtitech/trap-collection-server/src/config/mock"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/storage"
)

func TestSaveGameFile(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	ctrl := gomock.NewController(t)

	rootPath := "./save_game_file_test"
	mockConf := mock.NewMockStorageLocal(ctrl)
	mockConf.
		EXPECT().
		Path().
		Return(rootPath, nil)
	directoryManager, err := NewDirectoryManager(mockConf)
	if err != nil {
		t.Fatalf("failed to create directory manager: %v", err)
		return
	}
	defer func() {
		err := os.RemoveAll(rootPath)
		if err != nil {
			t.Fatalf("failed to remove directory: %v", err)
		}
	}()

	gameFile, err := NewGameFile(directoryManager)
	if err != nil {
		t.Fatalf("failed to create game file: %v", err)
	}

	fileRootPath := filepath.Join(string(rootPath), "files")

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
			description: "ファイルが存在しないので保存できる",
			fileID:      values.NewGameFileID(),
			reader:      bytes.NewBufferString("test"),
		},
		{
			description: "ファイルが存在するので保存できない",
			fileID:      values.NewGameFileID(),
			reader:      bytes.NewBufferString("test"),
			isFileExist: true,
			isErr:       true,
			err:         storage.ErrAlreadyExists,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			if testCase.isFileExist {
				f, err := os.Create(filepath.Join(fileRootPath, uuid.UUID(testCase.fileID).String()))
				if err != nil {
					t.Fatalf("failed to write file: %v", err)
				}
				f.Close()
			}

			expectBytes := testCase.reader.Bytes()

			err := gameFile.SaveGameFile(ctx, testCase.reader, testCase.fileID)

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

			f, err := os.Open(filepath.Join(fileRootPath, uuid.UUID(testCase.fileID).String()))
			if err != nil {
				t.Fatalf("failed to read file: %v", err)
			}
			defer f.Close()

			actualBytes, err := io.ReadAll(f)
			if err != nil {
				t.Fatalf("failed to read file: %v", err)
			}

			assert.Equal(t, expectBytes, actualBytes)
		})
	}
}

func TestGetGameFile(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	ctrl := gomock.NewController(t)

	rootPath := "./get_game_file_test"
	mockConf := mock.NewMockStorageLocal(ctrl)
	mockConf.
		EXPECT().
		Path().
		Return(rootPath, nil)
	directoryManager, err := NewDirectoryManager(mockConf)
	if err != nil {
		t.Fatalf("failed to create directory manager: %v", err)
		return
	}
	defer func() {
		err := os.RemoveAll(rootPath)
		if err != nil {
			t.Fatalf("failed to remove directory: %v", err)
		}
	}()

	gameFile, err := NewGameFile(directoryManager)
	if err != nil {
		t.Fatalf("failed to create game file: %v", err)
	}

	fileRootPath := filepath.Join(string(rootPath), "files")

	type test struct {
		description string
		file        *domain.GameFile
		isFileExist bool
		fileContent *bytes.Buffer
		isErr       bool
		err         error
	}

	testCases := []test{
		{
			description: "ファイルが存在するので読み込める",
			file: domain.NewGameFile(
				values.NewGameFileID(),
				values.GameFileTypeJar,
				"/path/to/game.jar",
				values.NewGameFileHashFromBytes([]byte{0x09, 0x8f, 0x6b, 0xcd, 0x46, 0x21, 0xd3, 0x73, 0xca, 0xde, 0x4e, 0x83, 0x26, 0x27, 0xb4, 0xf6}),
				time.Now(),
			),
			isFileExist: true,
			fileContent: bytes.NewBufferString("test"),
		},
		{
			description: "ファイルが存在しないので読み込めない",
			file: domain.NewGameFile(
				values.NewGameFileID(),
				values.GameFileTypeJar,
				"/path/to/game.jar",
				values.NewGameFileHashFromBytes([]byte{0x09, 0x8f, 0x6b, 0xcd, 0x46, 0x21, 0xd3, 0x73, 0xca, 0xde, 0x4e, 0x83, 0x26, 0x27, 0xb4, 0xf6}),
				time.Now(),
			),
			isErr: true,
			err:   storage.ErrNotFound,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			var expectBytes []byte
			if testCase.isFileExist {
				expectBytes = testCase.fileContent.Bytes()

				func() {
					f, err := os.Create(filepath.Join(fileRootPath, uuid.UUID(testCase.file.GetID()).String()))
					if err != nil {
						t.Fatalf("failed to write file: %v", err)
					}
					defer f.Close()

					_, err = io.Copy(f, testCase.fileContent)
					if err != nil {
						t.Fatalf("failed to write file: %v", err)
					}
				}()
			}

			buf := bytes.NewBuffer(nil)

			err := gameFile.GetGameFile(ctx, buf, testCase.file)

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
