package mock

import (
	"bytes"
	context "context"
	io "io"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	domain "github.com/traPtitech/trap-collection-server/src/domain"
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

// GetGameImage mocks base method.
func (m *GameImage) GetGameImage(ctx context.Context, writer io.Writer, image *domain.GameImage) error {
	ret0 := m.getGameImage(ctx, image)

	_, err := io.Copy(writer, m.buf)
	if err != nil {
		m.ctrl.T.Fatalf("unexpected error copying from buffer: %v", err)
	}

	return ret0
}

// GetGameImage mocks base method.
func (m *GameImage) getGameImage(ctx context.Context, image *domain.GameImage) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetGameImage", ctx, image)
	ret0, _ := ret[0].(error)
	return ret0
}

// GetGameImage indicates an expected call of GetGameImage.
func (mr *GameImageMockRecorder) GetGameImage(ctx, image interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetGameImage", reflect.TypeOf((*GameImage)(nil).GetGameImage), ctx, image)
}

// SaveGameImage mocks base method.
func (m *GameImage) SaveGameImage(ctx context.Context, reader io.Reader, image *domain.GameImage) error {
	ret0 := m.saveGameImage(ctx, image)

	_, err := io.Copy(m.buf, reader)
	if err != nil {
		m.ctrl.T.Fatalf("unexpected error copying to buffer: %v", err)
	}

	return ret0
}

func (m *GameImage) saveGameImage(ctx context.Context, image *domain.GameImage) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SaveGameImage", ctx, image)
	ret0, _ := ret[0].(error)

	return ret0
}

// SaveGameImage indicates an expected call of SaveGameImage.
func (mr *GameImageMockRecorder) SaveGameImage(ctx, image interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SaveGameImage", reflect.TypeOf((*GameImage)(nil).saveGameImage), ctx, image)
}
