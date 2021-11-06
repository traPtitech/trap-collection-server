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
	"github.com/traPtitech/trap-collection-server/pkg/common"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/storage"
)

func TestSaveGameVideo(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	rootPath := common.FilePath("./save_game_video_test")

	directoryManager := NewDirectoryManager(rootPath)
	defer func() {
		err := os.RemoveAll(string(rootPath))
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
		video       *domain.GameVideo
		isFileExist bool
		isErr       bool
		err         error
	}

	testCases := []test{
		{
			description: "ファイルが存在しないので保存できる",
			video: domain.NewGameVideo(
				values.NewGameVideoID(),
				values.GameVideoTypeMp4,
			),
		},
		{
			description: "ファイルが存在するので保存できない",
			video: domain.NewGameVideo(
				values.NewGameVideoID(),
				values.GameVideoTypeMp4,
			),
			isFileExist: true,
			isErr:       true,
			err:         storage.ErrAlreadyExists,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			if testCase.isFileExist {
				f, err := os.Create(filepath.Join(videoRootPath, uuid.UUID(testCase.video.GetID()).String()))
				if err != nil {
					t.Fatalf("failed to write file: %v", err)
				}
				f.Close()
			}

			videoBuf := bytes.NewBuffer(nil)

			switch testCase.video.GetType() {
			case values.GameVideoTypeMp4:
				err := func() error {
					f, err := os.Open("../../../testdata/1.mp4")
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
			default:
				t.Fatalf("invalid video type: %v\n", testCase.video.GetType())
			}

			expectBytes := videoBuf.Bytes()

			err := gameVideo.SaveGameVideo(ctx, videoBuf, testCase.video)

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

			f, err := os.Open(filepath.Join(videoRootPath, uuid.UUID(testCase.video.GetID()).String()))
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

func TestGetGameVideo(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	rootPath := common.FilePath("./get_game_video_test")

	directoryManager := NewDirectoryManager(rootPath)
	defer func() {
		err := os.RemoveAll(string(rootPath))
		if err != nil {
			t.Fatalf("failed to remove directory: %v", err)
		}
	}()

	gameVideo, err := NewGameVideo(directoryManager)
	if err != nil {
		t.Fatalf("failed to create game image: %v", err)
	}

	videoRootPath := filepath.Join(string(rootPath), "videos")

	type test struct {
		description string
		video       *domain.GameVideo
		isFileExist bool
		isErr       bool
		err         error
	}

	testCases := []test{
		{
			description: "ファイルが存在するので読み込める",
			video: domain.NewGameVideo(
				values.NewGameVideoID(),
				values.GameVideoTypeMp4,
			),
			isFileExist: true,
		},
		{
			description: "ファイルが存在しないので読み込めない",
			video: domain.NewGameVideo(
				values.NewGameVideoID(),
				values.GameVideoTypeMp4,
			),
			isErr: true,
			err:   storage.ErrNotFound,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			var expectBytes []byte
			if testCase.isFileExist {
				videoBuf := bytes.NewBuffer(nil)

				switch testCase.video.GetType() {
				case values.GameVideoTypeMp4:
					err := func() error {
						f, err := os.Open("../../../testdata/1.mp4")
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
				default:
					t.Fatalf("invalid video type: %v\n", testCase.video.GetType())
				}
				expectBytes = videoBuf.Bytes()

				func() {
					f, err := os.Create(filepath.Join(videoRootPath, uuid.UUID(testCase.video.GetID()).String()))
					if err != nil {
						t.Fatalf("failed to write file: %v", err)
					}
					defer f.Close()

					_, err = io.Copy(f, videoBuf)
					if err != nil {
						t.Fatalf("failed to write file: %v", err)
					}
				}()
			}

			buf := bytes.NewBuffer(nil)

			err := gameVideo.GetGameVideo(ctx, buf, testCase.video)

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
