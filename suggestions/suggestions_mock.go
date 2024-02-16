package suggestions

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type MockSuggestionsUmbrellaAPI struct {
	mock.Mock
}

func (_mu *MockSuggestionsUmbrellaAPI) FetchSuggestions(ctx context.Context, content []byte) (suggestion []byte, err error) {
	ret := _mu.Called(ctx, content)
	r1 := ret.Get(0).([]byte)
	rErr := ret.Error(1)
	return r1, rErr
}

func (_mu *MockSuggestionsUmbrellaAPI) Endpoint() string {
	ret := _mu.Called()
	r1 := ret.Get(0).(string)
	return r1
}

func (_mu *MockSuggestionsUmbrellaAPI) IsGTG(ctx context.Context) (string, error) {
	ret := _mu.Called(ctx)
	r1 := ret.Get(0).(string)
	rErr := ret.Error(1)
	return r1, rErr
}

func (_mu *MockSuggestionsUmbrellaAPI) IsValid() error {
	ret := _mu.Called()
	rErr := ret.Error(0)
	return rErr
}
