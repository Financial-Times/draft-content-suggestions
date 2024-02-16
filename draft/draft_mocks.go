package draft

import (
	"context"
	"io"

	"github.com/Financial-Times/go-logger/v2"
	"github.com/stretchr/testify/mock"
)

type MockDraftContentAPI struct {
	mock.Mock
}

func (_md *MockDraftContentAPI) FetchDraftContent(ctx context.Context, uuid string) (content []byte, err error) {
	ret := _md.Called(ctx, uuid)
	r1 := ret.Get(0).([]byte)
	rErr := ret.Error(1)
	return r1, rErr
}

func (_md *MockDraftContentAPI) FetchValidatedContent(ctx context.Context, body io.Reader, contentUUID string, contentType string, log *logger.UPPLogger) ([]byte, error) {
	ret := _md.Called(ctx, body, contentUUID, contentType, log)
	r1 := ret.Get(0).([]byte)
	rErr := ret.Error(1)
	return r1, rErr
}

func (_md *MockDraftContentAPI) Endpoint() string {
	ret := _md.Called()
	r1 := ret.Get(0).(string)
	return r1
}

func (_md *MockDraftContentAPI) IsGTG(ctx context.Context) (string, error) {
	ret := _md.Called(ctx)
	r1 := ret.Get(0).(string)
	rErr := ret.Error(1)
	return r1, rErr
}

func (_md *MockDraftContentAPI) IsValid() error {
	ret := _md.Called()
	rErr := ret.Error(0)
	return rErr
}

type MockValidatorResolver struct {
	mock.Mock
}

func (_mr *MockValidatorResolver) ValidatorForContentType(contentType string) (DraftContentValidator, error) {
	ret := _mr.Called(contentType)
	r0 := ret.Get(0).(DraftContentValidator)
	rErr := ret.Error(1)
	return r0, rErr
}

type MockValidator struct {
	mock.Mock
}

func (_mv *MockValidator) Validate(ctx context.Context, contentUUID string, nativeBody io.Reader, contentType string, log *logger.UPPLogger) (io.ReadCloser, error) {
	ret := _mv.Called(ctx, contentUUID, nativeBody, contentType, log)
	r0 := ret.Get(0).(io.Reader)
	rErr := ret.Error(1)
	return io.NopCloser(r0), rErr
}

func (_mv *MockValidator) GTG() error {
	ret := _mv.Called()
	rErr := ret.Error(0)
	return rErr
}

func (_mv *MockValidator) Endpoint() string {
	ret := _mv.Called()
	r1 := ret.Get(0).(string)
	return r1
}
