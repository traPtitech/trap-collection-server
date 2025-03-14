package local

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/traPtitech/trap-collection-server/src/config/mock"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/storage"
	"github.com/traPtitech/trap-collection-server/testdata"
	"go.uber.org/mock/gomock"
)

func TestSaveGameVideo(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	ctrl := gomock.NewController(t)

	rootPath := "./save_game_video_test"
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
			t.Fatalf("failed to remove directory: %v\n", err)
		}
	}()

	gameVideo, err := NewGameVideo(directoryManager)
	if err != nil {
		t.Fatalf("failed to create game video: %v\n", err)
	}

	videoRootPath := filepath.Join(string(rootPath), "videos")

	type test struct {
		description string
		videoID     values.GameVideoID
		isFileExist bool
		isErr       bool
		err         error
	}

	testCases := []test{
		{
			description: "ファイルが存在しないので保存できる",
			videoID:     values.NewGameVideoID(),
		},
		{
			description: "ファイルが存在するので保存できない",
			videoID:     values.NewGameVideoID(),
			isFileExist: true,
			isErr:       true,
			err:         storage.ErrAlreadyExists,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			if testCase.isFileExist {
				f, err := os.Create(filepath.Join(videoRootPath, uuid.UUID(testCase.videoID).String()))
				if err != nil {
					t.Fatalf("failed to write file: %v", err)
				}
				f.Close()
			}

			videoBuf := bytes.NewBuffer(nil)
			err := func() error {
				f, err := testdata.FS.Open("1.mp4")
				if err != nil {
					return fmt.Errorf("failed to open file: %w", err)
				}
				defer f.Close()

				_, err = io.Copy(videoBuf, f)
				if err != nil {
					return fmt.Errorf("failed to copy file: %w", err)
				}

				return nil
			}()
			if err != nil {
				t.Fatalf("failed to encode image: %s", err)
			}

			expectBytes := videoBuf.Bytes()

			err = gameVideo.SaveGameVideo(ctx, videoBuf, testCase.videoID)

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

			f, err := os.Open(filepath.Join(videoRootPath, uuid.UUID(testCase.videoID).String()))
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
