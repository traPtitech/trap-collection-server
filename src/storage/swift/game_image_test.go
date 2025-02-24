package swift

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"math/rand"
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
	"github.com/traPtitech/trap-collection-server/testdata"
	"go.uber.org/mock/gomock"
)

func TestSaveGameImage(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)

	client, err := newTestClient(ctx, ctrl, "save_game_image")
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	type test struct {
		description string
		imageID     values.GameImageID
		isFileExist bool
		isErr       bool
		err         error
	}

	testCases := []test{
		{
			description: "特に問題ないのでエラーなし",
			imageID:     values.NewGameImageID(),
		},
		{
			description: "ファイルが存在するのでErrAlreadyExists",
			imageID:     values.NewGameImageID(),
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
					fmt.Sprintf("images/%s", uuid.UUID(testCase.imageID).String()),
					"text/plain",
					"",
					strings.NewReader(""),
				)
				if err != nil {
					t.Fatalf("failed to create file: %v", err)
				}
			}

			imgBuf := bytes.NewBufferString("hoge")
			expectBytes := imgBuf.Bytes()

			err := gameImageStorage.SaveGameImage(ctx, imgBuf, testCase.imageID)

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
			err = client.loadFile(ctx, fmt.Sprintf("images/%s", uuid.UUID(testCase.imageID).String()), buf)
			if err != nil {
				t.Fatalf("failed to load file: %v", err)
			}

			assert.Equal(t, expectBytes, buf.Bytes())
		})
	}
}

func TestImageGetTempURL(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)

	client, err := newTestClient(ctx, ctrl, "get_game_image")
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

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
			imgBuf := bytes.NewBuffer(nil)

			err := func() error {
				var path string
				switch testCase.image.GetType() {
				case values.GameImageTypeJpeg:
					path = "1.jpg"
				case values.GameImageTypePng:
					path = "1.png"
				case values.GameImageTypeGif:
					path = "1.gif"
				default:
					return fmt.Errorf("invalid image type: %d", testCase.image.GetType())
				}
				f, err := testdata.FS.Open(path)
				if err != nil {
					return fmt.Errorf("failed to open testdata: %w", err)
				}
				defer f.Close()

				_, err = io.Copy(imgBuf, f)
				if err != nil {
					return fmt.Errorf("failed to copy testdata: %w", err)
				}

				return nil
			}()
			if err != nil {
				t.Fatalf("failed to prepare testdata: %v", err)
			}

			expectBytes := imgBuf.Bytes()

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
			tmpURL, err := gameImageStorage.GetTempURL(ctx, testCase.image, time.Second)

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

		key := gameImageStorage.imageKey(imageID)

		assert.Equal(t, fmt.Sprintf("images/%s", uuid.UUID(image.GetID()).String()), key)
	}
}
