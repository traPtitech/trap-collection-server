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

// GetTempURL mocks base method.
func (m *GameVideo) GetTempURL(ctx context.Context, video *domain.GameVideo, expires time.Duration) (values.GameVideoTmpURL, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetTempURL", ctx, video, expires)
	ret0, _ := ret[0].(values.GameVideoTmpURL)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetTempURL indicates an expected call of GetTempURL.
func (mr *GameVideoMockRecorder) GetTempURL(ctx, video, expires interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetTempURL", reflect.TypeOf((*GameVideo)(nil).GetTempURL), ctx, video, expires)
}

// SaveGameVideo mocks base method.
func (m *GameVideo) SaveGameVideo(ctx context.Context, reader io.Reader, videoID values.GameVideoID) error {
	ret0 := m.saveGameVideo(ctx, videoID)

	_, err := io.Copy(m.buf, reader)
	if err != nil {
		m.ctrl.T.Fatalf("unexpected error copying to buffer: %v", err)
	}

	return ret0
}

func (m *GameVideo) saveGameVideo(ctx context.Context, videoID values.GameVideoID) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SaveGameVideo", ctx, videoID)
	ret0, _ := ret[0].(error)

	return ret0
}

// SaveGameVideo indicates an expected call of SaveGameVideo.
func (mr *GameVideoMockRecorder) SaveGameVideo(ctx, video interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SaveGameVideo", reflect.TypeOf((*GameVideo)(nil).saveGameVideo), ctx, video)
}
