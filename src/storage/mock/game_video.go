package mock

import (
	"bytes"
	context "context"
	io "io"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	domain "github.com/traPtitech/trap-collection-server/src/domain"
)

// GameVideo is a mock of GameVideo interface.
type GameVideo struct {
	ctrl     *gomock.Controller
	recorder *GameVideoMockRecorder
	buf      *bytes.Buffer
}

// GameVideoMockRecorder is the mock recorder for MockGameVideo.
type GameVideoMockRecorder struct {
	mock *GameVideo
}

// NewGameVideo creates a new mock instance.
func NewGameVideo(ctrl *gomock.Controller, buf *bytes.Buffer) *GameVideo {
	mock := &GameVideo{ctrl: ctrl}
	mock.recorder = &GameVideoMockRecorder{mock}
	mock.buf = buf
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *GameVideo) EXPECT() *GameVideoMockRecorder {
	return m.recorder
}

// GetGameVideo mocks base method.
func (m *GameVideo) GetGameVideo(ctx context.Context, writer io.Writer, video *domain.GameVideo) error {
	ret0 := m.getGameVideo(ctx, video)

	_, err := io.Copy(writer, m.buf)
	if err != nil {
		m.ctrl.T.Fatalf("unexpected error copying from buffer: %v", err)
	}

	return ret0
}

// GetGameVideo mocks base method.
func (m *GameVideo) getGameVideo(ctx context.Context, video *domain.GameVideo) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetGameVideo", ctx, video)
	ret0, _ := ret[0].(error)
	return ret0
}

// GetGameVideo indicates an expected call of GetGameVideo.
func (mr *GameVideoMockRecorder) GetGameVideo(ctx, video interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetGameVideo", reflect.TypeOf((*GameVideo)(nil).GetGameVideo), ctx, video)
}

// SaveGameVideo mocks base method.
func (m *GameVideo) SaveGameVideo(ctx context.Context, reader io.Reader, video *domain.GameVideo) error {
	ret0 := m.saveGameVideo(ctx, video)

	_, err := io.Copy(m.buf, reader)
	if err != nil {
		m.ctrl.T.Fatalf("unexpected error copying to buffer: %v", err)
	}

	return ret0
}

func (m *GameVideo) saveGameVideo(ctx context.Context, video *domain.GameVideo) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SaveGameVideo", ctx, video)
	ret0, _ := ret[0].(error)

	return ret0
}

// SaveGameVideo indicates an expected call of SaveGameVideo.
func (mr *GameVideoMockRecorder) SaveGameVideo(ctx, video interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SaveGameVideo", reflect.TypeOf((*GameVideo)(nil).saveGameVideo), ctx, video)
}
