package mock

import (
	"bytes"
	context "context"
	io "io"
	reflect "reflect"
	time "time"

	gomock "github.com/golang/mock/gomock"
	domain "github.com/traPtitech/trap-collection-server/src/domain"
	values "github.com/traPtitech/trap-collection-server/src/domain/values"
)

// GameImage is a mock of GameImage interface.
type GameImage struct {
	ctrl     *gomock.Controller
	recorder *GameImageMockRecorder
	buf      *bytes.Buffer
}

// GameImageMockRecorder is the mock recorder for MockGameImage.
type GameImageMockRecorder struct {
	mock *GameImage
}

// NewGameImage creates a new mock instance.
func NewGameImage(ctrl *gomock.Controller, buf *bytes.Buffer) *GameImage {
	mock := &GameImage{ctrl: ctrl}
	mock.recorder = &GameImageMockRecorder{mock}
	mock.buf = buf
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *GameImage) EXPECT() *GameImageMockRecorder {
	return m.recorder
}

// GetTempURL mocks base method.
func (m *GameImage) GetTempURL(ctx context.Context, image *domain.GameImage, expires time.Duration) (values.GameImageTmpURL, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetTempURL", ctx, image, expires)
	ret0, _ := ret[0].(values.GameImageTmpURL)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetTempURL indicates an expected call of GetTempURL.
func (mr *GameImageMockRecorder) GetTempURL(ctx, image, expires interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetTempURL", reflect.TypeOf((*GameImage)(nil).GetTempURL), ctx, image, expires)
}

// SaveGameImage mocks base method.
func (m *GameImage) SaveGameImage(ctx context.Context, reader io.Reader, imageID values.GameImageID) error {
	ret0 := m.saveGameImage(ctx, imageID)

	_, err := io.Copy(m.buf, reader)
	if err != nil {
		m.ctrl.T.Fatalf("unexpected error copying to buffer: %v", err)
	}

	return ret0
}

func (m *GameImage) saveGameImage(ctx context.Context, imageID values.GameImageID) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SaveGameImage", ctx, imageID)
	ret0, _ := ret[0].(error)

	return ret0
}

// SaveGameImage indicates an expected call of SaveGameImage.
func (mr *GameImageMockRecorder) SaveGameImage(ctx, image interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SaveGameImage", reflect.TypeOf((*GameImage)(nil).saveGameImage), ctx, image)
}
