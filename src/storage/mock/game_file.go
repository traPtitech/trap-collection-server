package mock

import (
	"bytes"
	context "context"
	io "io"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	domain "github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
)

// GameFile is a mock of GameFile interface.
type GameFile struct {
	ctrl     *gomock.Controller
	recorder *GameFileMockRecorder
	buf      *bytes.Buffer
}

// GameFileMockRecorder is the mock recorder for MockGameFile.
type GameFileMockRecorder struct {
	mock *GameFile
}

// NewGameFile creates a new mock instance.
func NewGameFile(ctrl *gomock.Controller, buf *bytes.Buffer) *GameFile {
	mock := &GameFile{ctrl: ctrl}
	mock.recorder = &GameFileMockRecorder{mock}
	mock.buf = buf
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *GameFile) EXPECT() *GameFileMockRecorder {
	return m.recorder
}

// GetGameFile mocks base method.
func (m *GameFile) GetGameFile(ctx context.Context, writer io.Writer, file *domain.GameFile) error {
	ret0 := m.getGameFile(ctx, file)

	_, err := io.Copy(writer, m.buf)
	if err != nil {
		m.ctrl.T.Fatalf("unexpected error copying from buffer: %v", err)
	}

	return ret0
}

// GetGameFile mocks base method.
func (m *GameFile) getGameFile(ctx context.Context, file *domain.GameFile) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetGameFile", ctx, file)
	ret0, _ := ret[0].(error)
	return ret0
}

// GetGameFile indicates an expected call of GetGameFile.
func (mr *GameFileMockRecorder) GetGameFile(ctx, file interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetGameFile", reflect.TypeOf((*GameFile)(nil).GetGameFile), ctx, file)
}

// SaveGameFile mocks base method.
func (m *GameFile) SaveGameFile(ctx context.Context, reader io.Reader, fileID values.GameFileID) error {
	ret0 := m.saveGameFile(ctx, fileID)

	_, err := io.Copy(m.buf, reader)
	if err != nil {
		m.ctrl.T.Fatalf("unexpected error copying to buffer: %v", err)
	}

	return ret0
}

func (m *GameFile) saveGameFile(ctx context.Context, fileID values.GameFileID) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SaveGameFile", ctx, fileID)
	ret0, _ := ret[0].(error)

	return ret0
}

// SaveGameFile indicates an expected call of SaveGameFile.
func (mr *GameFileMockRecorder) SaveGameFile(ctx, file interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SaveGameFile", reflect.TypeOf((*GameFile)(nil).saveGameFile), ctx, file)
}
