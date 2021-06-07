// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import (
	orderer "github.com/Yunpeng-J/fabric-protos-go/orderer"
	mock "github.com/stretchr/testify/mock"
)

// MessageReceiver is an autogenerated mock type for the MessageReceiver type
type MessageReceiver struct {
	mock.Mock
}

// Consensus provides a mock function with given fields: req, sender
func (_m *MessageReceiver) Consensus(req *orderer.ConsensusRequest, sender uint64) error {
	ret := _m.Called(req, sender)

	var r0 error
	if rf, ok := ret.Get(0).(func(*orderer.ConsensusRequest, uint64) error); ok {
		r0 = rf(req, sender)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Submit provides a mock function with given fields: req, sender
func (_m *MessageReceiver) Submit(req *orderer.SubmitRequest, sender uint64) error {
	ret := _m.Called(req, sender)

	var r0 error
	if rf, ok := ret.Get(0).(func(*orderer.SubmitRequest, uint64) error); ok {
		r0 = rf(req, sender)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
