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
			),
		},
		{
			// 実際には発生しないが、念のため確認
			description: "想定外のファイルタイプなのでエラー",
			video: domain.NewGameVideo(
				values.NewGameVideoID(),
				100,
			),
			isErr: true,
		},
		{
			description: "ファイルが存在するのでErrAlreadyExists",
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
				err := client.saveFile(
					ctx,
					fmt.Sprintf("videos/%s", uuid.UUID(testCase.video.GetID()).String()),
					"text/plain",
					"",
					strings.NewReader(""),
				)
				if err != nil {
					t.Fatalf("failed to create file: %v", err)
				}
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
				videoBuf = bytes.NewBufferString("hoge")
			}

			expectBytes := videoBuf.Bytes()

			err := gameVideoStorage.SaveGameVideo(ctx, videoBuf, testCase.video)

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
			err = client.loadFile(ctx, fmt.Sprintf("videos/%s", uuid.UUID(testCase.video.GetID()).String()), buf)
			if err != nil {
				t.Fatalf("failed to load file: %v", err)
			}

			assert.Equal(t, expectBytes, buf.Bytes())
		})
	}
}

func TestGetGameVideo(t *testing.T) {
	t.Parallel()

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
			),
			isFileExist: true,
		},
		{
			description: "ファイルが存在しないのでErrNotFound",
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
