// Code generated by mockery v2.20.0. DO NOT EDIT.

package pg_mocks

import (
	context "context"

	entity "github.com/OkDenAl/mbtiles_converter/internal/entity"
	mock "github.com/stretchr/testify/mock"
)

// Repository is an autogenerated mock type for the Repository type
type Repository struct {
	mock.Mock
}

// GetNElements provides a mock function with given fields: ctx, n, offset
func (_m *Repository) GetNElements(ctx context.Context, n int, offset int) ([]entity.MapPoint, error) {
	ret := _m.Called(ctx, n, offset)

	var r0 []entity.MapPoint
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, int, int) ([]entity.MapPoint, error)); ok {
		return rf(ctx, n, offset)
	}
	if rf, ok := ret.Get(0).(func(context.Context, int, int) []entity.MapPoint); ok {
		r0 = rf(ctx, n, offset)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]entity.MapPoint)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, int, int) error); ok {
		r1 = rf(ctx, n, offset)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type mockConstructorTestingTNewRepository interface {
	mock.TestingT
	Cleanup(func())
}

// NewRepository creates a new instance of Repository. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewRepository(t mockConstructorTestingTNewRepository) *Repository {
	mock := &Repository{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
