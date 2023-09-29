// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/gardener/gardener/extensions/pkg/controller (interfaces: ChartRendererFactory)

// Package mock is a generated GoMock package.
package mock

import (
	reflect "reflect"

	chartrenderer "github.com/gardener/gardener/pkg/chartrenderer"
	gomock "go.uber.org/mock/gomock"
)

// MockChartRendererFactory is a mock of ChartRendererFactory interface.
type MockChartRendererFactory struct {
	ctrl     *gomock.Controller
	recorder *MockChartRendererFactoryMockRecorder
}

// MockChartRendererFactoryMockRecorder is the mock recorder for MockChartRendererFactory.
type MockChartRendererFactoryMockRecorder struct {
	mock *MockChartRendererFactory
}

// NewMockChartRendererFactory creates a new mock instance.
func NewMockChartRendererFactory(ctrl *gomock.Controller) *MockChartRendererFactory {
	mock := &MockChartRendererFactory{ctrl: ctrl}
	mock.recorder = &MockChartRendererFactoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockChartRendererFactory) EXPECT() *MockChartRendererFactoryMockRecorder {
	return m.recorder
}

// NewChartRendererForShoot mocks base method.
func (m *MockChartRendererFactory) NewChartRendererForShoot(arg0 string) (chartrenderer.Interface, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "NewChartRendererForShoot", arg0)
	ret0, _ := ret[0].(chartrenderer.Interface)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// NewChartRendererForShoot indicates an expected call of NewChartRendererForShoot.
func (mr *MockChartRendererFactoryMockRecorder) NewChartRendererForShoot(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "NewChartRendererForShoot", reflect.TypeOf((*MockChartRendererFactory)(nil).NewChartRendererForShoot), arg0)
}
