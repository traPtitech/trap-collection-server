package swift

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/traPtitech/trap-collection-server/pkg/common"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/storage"
)

func TestSaveGameVideo(t *testing.T) {
	ctx := context.Background()

	client, err := newTestClient(
		ctx,
		common.SwiftContainer("save_game_video"),
		common.FilePath("save_game_video"),
	)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	defer func() {
		err := os.RemoveAll("save_game_video")
		if err != nil {
			t.Fatalf("failed to remove directory: %v", err)
		}
	}()

	gameVideoStorage := NewGameVideo(client)

	type test struct {
		description string
		videoID     values.GameVideoID
		isFileExist bool
		isErr       bool
		err         error
	}

	testCases := []test{
		{
			description: "特に問題ないのでエラーなし",
			videoID:     values.NewGameVideoID(),
		},
		{
			description: "ファイルが存在するのでErrAlreadyExists",
			videoID:     values.NewGameVideoID(),
			isFileExist: true,
			isErr:       true,
			err:         storage.ErrAlreadyExists,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			if testCase.isFileExist {
				err := client.saveFile(
					ctx,
					fmt.Sprintf("videos/%s", uuid.UUID(testCase.videoID).String()),
					"text/plain",
					"",
					strings.NewReader(""),
				)
				if err != nil {
					t.Fatalf("failed to create file: %v", err)
				}
			}

			videoBuf := bytes.NewBuffer(nil)
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

			expectBytes := videoBuf.Bytes()

			err = gameVideoStorage.SaveGameVideo(ctx, videoBuf, testCase.videoID)

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
			err = client.loadFile(ctx, fmt.Sprintf("videos/%s", uuid.UUID(testCase.videoID).String()), buf)
			if err != nil {
				t.Fatalf("failed to load file: %v", err)
			}

			assert.Equal(t, expectBytes, buf.Bytes())
		})
	}
}

func TestGetGameVideo(t *testing.T) {
	ctx := context.Background()

	client, err := newTestClient(
		ctx,
		common.SwiftContainer("get_game_video"),
		common.FilePath("get_game_video"),
	)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	defer func() {
		err := os.RemoveAll("get_game_video")
		if err != nil {
			t.Fatalf("failed to remove directory: %v", err)
		}
	}()

	gameVideoStorage := NewGameVideo(client)

	type test struct {
		description string
		video       *domain.GameVideo
		isFileExist bool
		isErr       bool
		err         error
	}

	testCases := []test{
		{
			description: "特に問題ないのでエラーなし",
			video: domain.NewGameVideo(
				values.NewGameVideoID(),
				values.GameVideoTypeMp4,
				time.Now(),
			),
			isFileExist: true,
		},
		{
			description: "ファイルが存在しないのでErrNotFound",
			video: domain.NewGameVideo(
				values.NewGameVideoID(),
				values.GameVideoTypeMp4,
				time.Now(),
			),
			isErr: true,
			err:   storage.ErrNotFound,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
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
				videoBuf = bytes.NewBufferString("hoge")
			}

			expectBytes := videoBuf.Bytes()

			if testCase.isFileExist {
				err := client.saveFile(
					ctx,
					fmt.Sprintf("videos/%s", uuid.UUID(testCase.video.GetID()).String()),
					"",
					"",
					videoBuf,
				)
				if err != nil {
					t.Fatalf("failed to create file: %v", err)
				}
			}

			buf := bytes.NewBuffer(nil)
			err := gameVideoStorage.GetGameVideo(ctx, buf, testCase.video)

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

func TestVideoKey(t *testing.T) {
	t.Parallel()

	// clientは使わないのでnilでOK
	gameVideoStorage := NewGameVideo(nil)

	loopNum := 100

	for i := 0; i < loopNum; i++ {
		videoID := values.NewGameVideoID()

		key := gameVideoStorage.videoKey(videoID)

		assert.Equal(t, fmt.Sprintf("videos/%s", uuid.UUID(videoID).String()), key)
	}
}
