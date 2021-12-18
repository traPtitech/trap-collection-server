package swift

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"math/rand"
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

func TestSaveGameImage(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	client, err := newTestClient(
		ctx,
		common.SwiftContainer("save_game_image"),
		common.FilePath("save_game_image"),
	)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	defer func() {
		err := os.RemoveAll("save_game_image")
		if err != nil {
			t.Fatalf("failed to remove directory: %v", err)
		}
	}()

	type test struct {
		description string
		image       *domain.GameImage
		isFileExist bool
		isErr       bool
		err         error
	}

	testCases := []test{
		{
			description: "特に問題ないのでエラーなし",
			image: domain.NewGameImage(
				values.NewGameImageID(),
				values.GameImageTypeJpeg,
				time.Now(),
			),
		},
		{
			description: "pngでもエラーなし",
			image: domain.NewGameImage(
				values.NewGameImageID(),
				values.GameImageTypePng,
				time.Now(),
			),
		},
		{
			description: "gifでもエラーなし",
			image: domain.NewGameImage(
				values.NewGameImageID(),
				values.GameImageTypeGif,
				time.Now(),
			),
		},
		{
			description: "想定外のファイルタイプなのでエラー",
			image: domain.NewGameImage(
				values.NewGameImageID(),
				100,
				time.Now(),
			),
			isErr: true,
		},
		{
			description: "ファイルが存在するのでErrAlreadyExists",
			image: domain.NewGameImage(
				values.NewGameImageID(),
				values.GameImageTypeJpeg,
				time.Now(),
			),
			isFileExist: true,
			isErr:       true,
			err:         storage.ErrAlreadyExists,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			gameImageStorage := NewGameImage(client)

			if testCase.isFileExist {
				err := client.saveFile(
					ctx,
					fmt.Sprintf("images/%s", uuid.UUID(testCase.image.GetID()).String()),
					"text/plain",
					"",
					strings.NewReader(""),
				)
				if err != nil {
					t.Fatalf("failed to create file: %v", err)
				}
			}

			var expectBytes []byte
			img := image.NewRGBA(image.Rect(0, 0, 3000, 3000))
			imgBuf := bytes.NewBuffer(nil)

			switch testCase.image.GetType() {
			case values.GameImageTypeJpeg:
				err := jpeg.Encode(imgBuf, img, nil)
				if err != nil {
					t.Fatalf("failed to encode image: %v\n", err)
				}
			case values.GameImageTypePng:
				err := png.Encode(imgBuf, img)
				if err != nil {
					t.Fatalf("failed to encode image: %v\n", err)
				}
			case values.GameImageTypeGif:
				err := gif.Encode(imgBuf, img, nil)
				if err != nil {
					t.Fatalf("failed to encode image: %v\n", err)
				}
			default:
				imgBuf = bytes.NewBufferString("hoge")
			}
			expectBytes = imgBuf.Bytes()

			err := gameImageStorage.SaveGameImage(ctx, imgBuf, testCase.image)

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
			err = client.loadFile(ctx, fmt.Sprintf("images/%s", uuid.UUID(testCase.image.GetID()).String()), buf)
			if err != nil {
				t.Fatalf("failed to load file: %v", err)
			}

			assert.Equal(t, expectBytes, buf.Bytes())
		})
	}
}

func TestGetGameImage(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	client, err := newTestClient(
		ctx,
		common.SwiftContainer("get_game_image"),
		common.FilePath("get_game_image"),
	)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	defer func() {
		err := os.RemoveAll("get_game_image")
		if err != nil {
			t.Fatalf("failed to remove directory: %v", err)
		}
	}()

	gameImageStorage := NewGameImage(client)

	type test struct {
		description string
		image       *domain.GameImage
		isFileExist bool
		isErr       bool
		err         error
	}

	testCases := []test{
		{
			description: "特に問題ないのでエラーなし",
			image: domain.NewGameImage(
				values.NewGameImageID(),
				values.GameImageTypeJpeg,
				time.Now(),
			),
			isFileExist: true,
		},
		{
			description: "pngでもエラーなし",
			image: domain.NewGameImage(
				values.NewGameImageID(),
				values.GameImageTypePng,
				time.Now(),
			),
			isFileExist: true,
		},
		{
			description: "gifでもエラーなし",
			image: domain.NewGameImage(
				values.NewGameImageID(),
				values.GameImageTypeGif,
				time.Now(),
			),
			isFileExist: true,
		},
		{
			description: "ファイルが存在しないのでErrNotFound",
			image: domain.NewGameImage(
				values.NewGameImageID(),
				values.GameImageTypeJpeg,
				time.Now(),
			),
			isErr: true,
			err:   storage.ErrNotFound,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			var expectBytes []byte
			img := image.NewRGBA(image.Rect(0, 0, 3000, 3000))
			imgBuf := bytes.NewBuffer(nil)

			switch testCase.image.GetType() {
			case values.GameImageTypeJpeg:
				err := jpeg.Encode(imgBuf, img, nil)
				if err != nil {
					t.Fatalf("failed to encode image: %v\n", err)
				}
			case values.GameImageTypePng:
				err := png.Encode(imgBuf, img)
				if err != nil {
					t.Fatalf("failed to encode image: %v\n", err)
				}
			case values.GameImageTypeGif:
				err := gif.Encode(imgBuf, img, nil)
				if err != nil {
					t.Fatalf("failed to encode image: %v\n", err)
				}
			default:
				imgBuf = bytes.NewBufferString("hoge")
			}
			expectBytes = imgBuf.Bytes()

			if testCase.isFileExist {
				err := client.saveFile(
					ctx,
					fmt.Sprintf("images/%s", uuid.UUID(testCase.image.GetID()).String()),
					"",
					"",
					imgBuf,
				)
				if err != nil {
					t.Fatalf("failed to create file: %v", err)
				}
			}

			buf := bytes.NewBuffer(nil)
			err := gameImageStorage.GetGameImage(ctx, buf, testCase.image)

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

func TestImageKey(t *testing.T) {
	t.Parallel()

	// clientは使わないのでnilでOK
	gameImageStorage := NewGameImage(nil)

	loopNum := 100

	for i := 0; i < loopNum; i++ {
		imageID := values.NewGameImageID()

		image := domain.NewGameImage(
			imageID,
			values.GameImageType(rand.Intn(3)),
			time.Now(),
		)

		key := gameImageStorage.imageKey(image)

		assert.Equal(t, fmt.Sprintf("images/%s", uuid.UUID(image.GetID()).String()), key)
	}
}
