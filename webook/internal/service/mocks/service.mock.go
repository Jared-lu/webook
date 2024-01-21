// Code generated by MockGen. DO NOT EDIT.
// Source: ./types.go

// Package svcmocks is a generated GoMock package.
package svcmocks

import (
	context "context"
	reflect "reflect"
	domain "webook/webook/internal/domain"

	gomock "go.uber.org/mock/gomock"
)

// MockUserService is a mock of UserService interface.
type MockUserService struct {
	ctrl     *gomock.Controller
	recorder *MockUserServiceMockRecorder
}

// MockUserServiceMockRecorder is the mock recorder for MockUserService.
type MockUserServiceMockRecorder struct {
	mock *MockUserService
}

// NewMockUserService creates a new mock instance.
func NewMockUserService(ctrl *gomock.Controller) *MockUserService {
	mock := &MockUserService{ctrl: ctrl}
	mock.recorder = &MockUserServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockUserService) EXPECT() *MockUserServiceMockRecorder {
	return m.recorder
}

// Edit mocks base method.
func (m *MockUserService) Edit(ctx context.Context, user domain.User) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Edit", ctx, user)
	ret0, _ := ret[0].(error)
	return ret0
}

// Edit indicates an expected call of Edit.
func (mr *MockUserServiceMockRecorder) Edit(ctx, user interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Edit", reflect.TypeOf((*MockUserService)(nil).Edit), ctx, user)
}

// FindOrCreateByPhone mocks base method.
func (m *MockUserService) FindOrCreateByPhone(ctx context.Context, phone string) (domain.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FindOrCreateByPhone", ctx, phone)
	ret0, _ := ret[0].(domain.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FindOrCreateByPhone indicates an expected call of FindOrCreateByPhone.
func (mr *MockUserServiceMockRecorder) FindOrCreateByPhone(ctx, phone interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FindOrCreateByPhone", reflect.TypeOf((*MockUserService)(nil).FindOrCreateByPhone), ctx, phone)
}

// FindOrCreateByWechat mocks base method.
func (m *MockUserService) FindOrCreateByWechat(ctx context.Context, info domain.WechatInfo) (domain.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FindOrCreateByWechat", ctx, info)
	ret0, _ := ret[0].(domain.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FindOrCreateByWechat indicates an expected call of FindOrCreateByWechat.
func (mr *MockUserServiceMockRecorder) FindOrCreateByWechat(ctx, info interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FindOrCreateByWechat", reflect.TypeOf((*MockUserService)(nil).FindOrCreateByWechat), ctx, info)
}

// Login mocks base method.
func (m *MockUserService) Login(ctx context.Context, user domain.User) (domain.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Login", ctx, user)
	ret0, _ := ret[0].(domain.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Login indicates an expected call of Login.
func (mr *MockUserServiceMockRecorder) Login(ctx, user interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Login", reflect.TypeOf((*MockUserService)(nil).Login), ctx, user)
}

// Profile mocks base method.
func (m *MockUserService) Profile(ctx context.Context, user domain.User) (domain.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Profile", ctx, user)
	ret0, _ := ret[0].(domain.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Profile indicates an expected call of Profile.
func (mr *MockUserServiceMockRecorder) Profile(ctx, user interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Profile", reflect.TypeOf((*MockUserService)(nil).Profile), ctx, user)
}

// SignUp mocks base method.
func (m *MockUserService) SignUp(ctx context.Context, u domain.User) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SignUp", ctx, u)
	ret0, _ := ret[0].(error)
	return ret0
}

// SignUp indicates an expected call of SignUp.
func (mr *MockUserServiceMockRecorder) SignUp(ctx, u interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SignUp", reflect.TypeOf((*MockUserService)(nil).SignUp), ctx, u)
}

// MockCodeService is a mock of CodeService interface.
type MockCodeService struct {
	ctrl     *gomock.Controller
	recorder *MockCodeServiceMockRecorder
}

// MockCodeServiceMockRecorder is the mock recorder for MockCodeService.
type MockCodeServiceMockRecorder struct {
	mock *MockCodeService
}

// NewMockCodeService creates a new mock instance.
func NewMockCodeService(ctrl *gomock.Controller) *MockCodeService {
	mock := &MockCodeService{ctrl: ctrl}
	mock.recorder = &MockCodeServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockCodeService) EXPECT() *MockCodeServiceMockRecorder {
	return m.recorder
}

// Send mocks base method.
func (m *MockCodeService) Send(ctx context.Context, biz, phone string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Send", ctx, biz, phone)
	ret0, _ := ret[0].(error)
	return ret0
}

// Send indicates an expected call of Send.
func (mr *MockCodeServiceMockRecorder) Send(ctx, biz, phone interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Send", reflect.TypeOf((*MockCodeService)(nil).Send), ctx, biz, phone)
}

// Verify mocks base method.
func (m *MockCodeService) Verify(ctx context.Context, biz, phone, inputCode string) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Verify", ctx, biz, phone, inputCode)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Verify indicates an expected call of Verify.
func (mr *MockCodeServiceMockRecorder) Verify(ctx, biz, phone, inputCode interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Verify", reflect.TypeOf((*MockCodeService)(nil).Verify), ctx, biz, phone, inputCode)
}
