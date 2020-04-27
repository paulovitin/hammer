// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import (
	hammer "github.com/allisson/hammer"
	mock "github.com/stretchr/testify/mock"
)

// SubscriptionRepository is an autogenerated mock type for the SubscriptionRepository type
type SubscriptionRepository struct {
	mock.Mock
}

// Find provides a mock function with given fields: id
func (_m *SubscriptionRepository) Find(id string) (hammer.Subscription, error) {
	ret := _m.Called(id)

	var r0 hammer.Subscription
	if rf, ok := ret.Get(0).(func(string) hammer.Subscription); ok {
		r0 = rf(id)
	} else {
		r0 = ret.Get(0).(hammer.Subscription)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// FindAll provides a mock function with given fields: limit, offset
func (_m *SubscriptionRepository) FindAll(limit int, offset int) ([]hammer.Subscription, error) {
	ret := _m.Called(limit, offset)

	var r0 []hammer.Subscription
	if rf, ok := ret.Get(0).(func(int, int) []hammer.Subscription); ok {
		r0 = rf(limit, offset)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]hammer.Subscription)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(int, int) error); ok {
		r1 = rf(limit, offset)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// FindByTopic provides a mock function with given fields: topicID
func (_m *SubscriptionRepository) FindByTopic(topicID string) ([]hammer.Subscription, error) {
	ret := _m.Called(topicID)

	var r0 []hammer.Subscription
	if rf, ok := ret.Get(0).(func(string) []hammer.Subscription); ok {
		r0 = rf(topicID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]hammer.Subscription)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(topicID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Store provides a mock function with given fields: subscription
func (_m *SubscriptionRepository) Store(subscription *hammer.Subscription) error {
	ret := _m.Called(subscription)

	var r0 error
	if rf, ok := ret.Get(0).(func(*hammer.Subscription) error); ok {
		r0 = rf(subscription)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
