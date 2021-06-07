// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import (
	common "github.com/Yunpeng-J/fabric-protos-go/common"
	mock "github.com/stretchr/testify/mock"
)

// ChainPuller is an autogenerated mock type for the ChainPuller type
type ChainPuller struct {
	mock.Mock
}

// Close provides a mock function with given fields:
func (_m *ChainPuller) Close() {
	_m.Called()
}

// HeightsByEndpoints provides a mock function with given fields:
func (_m *ChainPuller) HeightsByEndpoints() (map[string]uint64, error) {
	ret := _m.Called()

	var r0 map[string]uint64
	if rf, ok := ret.Get(0).(func() map[string]uint64); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(map[string]uint64)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// PullBlock provides a mock function with given fields: seq
func (_m *ChainPuller) PullBlock(seq uint64) *common.Block {
	ret := _m.Called(seq)

	var r0 *common.Block
	if rf, ok := ret.Get(0).(func(uint64) *common.Block); ok {
		r0 = rf(seq)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*common.Block)
		}
	}

	return r0
}
