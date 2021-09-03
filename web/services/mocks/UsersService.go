// Code generated by mockery v0.0.0-dev. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"

// UsersService is an autogenerated mock type for the UsersService type
type UsersService struct {
	mock.Mock
}

// AuthenticateByEmailPassword provides a mock function with given fields: email, password
func (_m *UsersService) AuthenticateByEmailPassword(email string, password string) bool {
	ret := _m.Called(email, password)

	var r0 bool
	if rf, ok := ret.Get(0).(func(string, string) bool); ok {
		r0 = rf(email, password)
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// CreateUserByEmailPassword provides a mock function with given fields: email, password
func (_m *UsersService) CreateUserByEmailPassword(email string, password string) {
	_m.Called(email, password)
}
