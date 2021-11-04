package local

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/traPtitech/trap-collection-server/pkg/common"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/storage"
)

func TestSaveGameImage(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	rootPath := common.FilePath("./save_game_image_test")

	directoryManager := NewDirectoryManager(rootPath)
	defer func() {
		err := os.RemoveAll(string(rootPath))
		if err != nil {
			t.Fatalf("failed to remove directory: %v", err)
		}
	}()

	gameImage, err := NewGameImage(directoryManager)
	if err != nil {
		t.Fatalf("failed to create game image: %v", err)
	}

	imageRootPath := filepath.Join(string(rootPath), "images")

	type test struct {
		description string
		image       *domain.GameImage
		reader      *bytes.Buffer
		isFileExist bool
		isErr       bool
		err         error
	}

	testCases := []test{
		{
			description: "ファイルが存在しないので保存できる",
			image: domain.NewGameImage(
				values.NewGameImageID(),
				values.GameImageTypeJpeg,
			),
			reader: bytes.NewBufferString("a"),
		},
		{
			description: "ファイルが存在するので保存できない",
			image: domain.NewGameImage(
				values.NewGameImageID(),
				values.GameImageTypeJpeg,
			),
			reader:      bytes.NewBufferString("b"),
			isFileExist: true,
			isErr:       true,
			err:         storage.ErrAlreadyExists,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			if testCase.isFileExist {
				f, err := os.Create(filepath.Join(imageRootPath, uuid.UUID(testCase.image.GetID()).String()))
				if err != nil {
					t.Fatalf("failed to write file: %v", err)
				}
				f.Close()
			}

			expectBytes := testCase.reader.Bytes()

			err := gameImage.SaveGameImage(ctx, testCase.reader, testCase.image)

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

			f, err := os.Open(filepath.Join(imageRootPath, uuid.UUID(testCase.image.GetID()).String()))
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
