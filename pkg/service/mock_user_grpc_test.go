// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/instill-ai/protogen-go/vdp/mgmt/v1alpha (interfaces: MgmtPrivateServiceClient)

// Package service_test is a generated GoMock package.
package service_test

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	mgmtv1alpha "github.com/instill-ai/protogen-go/vdp/mgmt/v1alpha"
	grpc "google.golang.org/grpc"
)

// MockMgmtPrivateServiceClient is a mock of MgmtPrivateServiceClient interface.
type MockMgmtPrivateServiceClient struct {
	ctrl     *gomock.Controller
	recorder *MockMgmtPrivateServiceClientMockRecorder
}

// MockMgmtPrivateServiceClientMockRecorder is the mock recorder for MockMgmtPrivateServiceClient.
type MockMgmtPrivateServiceClientMockRecorder struct {
	mock *MockMgmtPrivateServiceClient
}

// NewMockMgmtPrivateServiceClient creates a new mock instance.
func NewMockMgmtPrivateServiceClient(ctrl *gomock.Controller) *MockMgmtPrivateServiceClient {
	mock := &MockMgmtPrivateServiceClient{ctrl: ctrl}
	mock.recorder = &MockMgmtPrivateServiceClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockMgmtPrivateServiceClient) EXPECT() *MockMgmtPrivateServiceClientMockRecorder {
	return m.recorder
}

// CreateUserAdmin mocks base method.
func (m *MockMgmtPrivateServiceClient) CreateUserAdmin(arg0 context.Context, arg1 *mgmtv1alpha.CreateUserAdminRequest, arg2 ...grpc.CallOption) (*mgmtv1alpha.CreateUserAdminResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "CreateUserAdmin", varargs...)
	ret0, _ := ret[0].(*mgmtv1alpha.CreateUserAdminResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateUserAdmin indicates an expected call of CreateUserAdmin.
func (mr *MockMgmtPrivateServiceClientMockRecorder) CreateUserAdmin(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateUserAdmin", reflect.TypeOf((*MockMgmtPrivateServiceClient)(nil).CreateUserAdmin), varargs...)
}

// DeleteUserAdmin mocks base method.
func (m *MockMgmtPrivateServiceClient) DeleteUserAdmin(arg0 context.Context, arg1 *mgmtv1alpha.DeleteUserAdminRequest, arg2 ...grpc.CallOption) (*mgmtv1alpha.DeleteUserAdminResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "DeleteUserAdmin", varargs...)
	ret0, _ := ret[0].(*mgmtv1alpha.DeleteUserAdminResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DeleteUserAdmin indicates an expected call of DeleteUserAdmin.
func (mr *MockMgmtPrivateServiceClientMockRecorder) DeleteUserAdmin(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteUserAdmin", reflect.TypeOf((*MockMgmtPrivateServiceClient)(nil).DeleteUserAdmin), varargs...)
}

// GetUserAdmin mocks base method.
func (m *MockMgmtPrivateServiceClient) GetUserAdmin(arg0 context.Context, arg1 *mgmtv1alpha.GetUserAdminRequest, arg2 ...grpc.CallOption) (*mgmtv1alpha.GetUserAdminResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "GetUserAdmin", varargs...)
	ret0, _ := ret[0].(*mgmtv1alpha.GetUserAdminResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetUserAdmin indicates an expected call of GetUserAdmin.
func (mr *MockMgmtPrivateServiceClientMockRecorder) GetUserAdmin(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUserAdmin", reflect.TypeOf((*MockMgmtPrivateServiceClient)(nil).GetUserAdmin), varargs...)
}

// ListUsersAdmin mocks base method.
func (m *MockMgmtPrivateServiceClient) ListUsersAdmin(arg0 context.Context, arg1 *mgmtv1alpha.ListUsersAdminRequest, arg2 ...grpc.CallOption) (*mgmtv1alpha.ListUsersAdminResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "ListUsersAdmin", varargs...)
	ret0, _ := ret[0].(*mgmtv1alpha.ListUsersAdminResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListUsersAdmin indicates an expected call of ListUsersAdmin.
func (mr *MockMgmtPrivateServiceClientMockRecorder) ListUsersAdmin(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListUsersAdmin", reflect.TypeOf((*MockMgmtPrivateServiceClient)(nil).ListUsersAdmin), varargs...)
}

// LookUpUserAdmin mocks base method.
func (m *MockMgmtPrivateServiceClient) LookUpUserAdmin(arg0 context.Context, arg1 *mgmtv1alpha.LookUpUserAdminRequest, arg2 ...grpc.CallOption) (*mgmtv1alpha.LookUpUserAdminResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "LookUpUserAdmin", varargs...)
	ret0, _ := ret[0].(*mgmtv1alpha.LookUpUserAdminResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// LookUpUserAdmin indicates an expected call of LookUpUserAdmin.
func (mr *MockMgmtPrivateServiceClientMockRecorder) LookUpUserAdmin(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "LookUpUserAdmin", reflect.TypeOf((*MockMgmtPrivateServiceClient)(nil).LookUpUserAdmin), varargs...)
}

// UpdateUserAdmin mocks base method.
func (m *MockMgmtPrivateServiceClient) UpdateUserAdmin(arg0 context.Context, arg1 *mgmtv1alpha.UpdateUserAdminRequest, arg2 ...grpc.CallOption) (*mgmtv1alpha.UpdateUserAdminResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "UpdateUserAdmin", varargs...)
	ret0, _ := ret[0].(*mgmtv1alpha.UpdateUserAdminResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UpdateUserAdmin indicates an expected call of UpdateUserAdmin.
func (mr *MockMgmtPrivateServiceClientMockRecorder) UpdateUserAdmin(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateUserAdmin", reflect.TypeOf((*MockMgmtPrivateServiceClient)(nil).UpdateUserAdmin), varargs...)
}

// ValidateToken mocks base method.
func (m *MockMgmtPrivateServiceClient) ValidateToken(arg0 context.Context, arg1 *mgmtv1alpha.ValidateTokenRequest, arg2 ...grpc.CallOption) (*mgmtv1alpha.ValidateTokenResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "ValidateToken", varargs...)
	ret0, _ := ret[0].(*mgmtv1alpha.ValidateTokenResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ValidateToken indicates an expected call of ValidateToken.
func (mr *MockMgmtPrivateServiceClientMockRecorder) ValidateToken(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ValidateToken", reflect.TypeOf((*MockMgmtPrivateServiceClient)(nil).ValidateToken), varargs...)
}
