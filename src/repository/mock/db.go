package repository

import (
	context "context"
	sql "database/sql"
	"fmt"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockDB is a mock of DB interface.
type MockDB struct {
	ctrl     *gomock.Controller
	recorder *MockDBMockRecorder
}

// MockDBMockRecorder is the mock recorder for MockDB.
type MockDBMockRecorder struct {
	mock *MockDB
}

// NewMockDB creates a new mock instance.
func NewMockDB(ctrl *gomock.Controller) *MockDB {
	mock := &MockDB{ctrl: ctrl}
	mock.recorder = &MockDBMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockDB) EXPECT() *MockDBMockRecorder {
	return m.recorder
}

// Get mocks base method.
func (m *MockDB) Get() *sql.DB {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get")
	ret0, _ := ret[0].(*sql.DB)
	return ret0
}

// Get indicates an expected call of Get.
func (mr *MockDBMockRecorder) Get() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockDB)(nil).Get))
}

// Transaction mocks base method.
func (m *MockDB) Transaction(ctx context.Context, txOpt *sql.TxOptions, fn func(context.Context) error) error {
	err := fn(ctx)
	if err != nil {
		return fmt.Errorf("failed in transaction: %w", err)
	}

	return nil
}
