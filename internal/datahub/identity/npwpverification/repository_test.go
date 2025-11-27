package npwpverification

import (
	"bytes"
	"encoding/json"
	"errors"
	"front-office/configs/application"
	"front-office/pkg/common/constant"
	"front-office/pkg/common/model"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockClient struct {
	mock.Mock
}

func (m *MockClient) Do(req *http.Request) (*http.Response, error) {
	args := m.Called(req)
	return args.Get(0).(*http.Response), args.Error(1)
}

func setupMockRepo(t *testing.T, response *http.Response, err error) (Repository, *MockClient) {
	t.Helper()

	mockClient := new(MockClient)
	mockClient.On("Do", mock.Anything).Return(response, err)

	repo := NewRepository(&application.Config{
		App: &application.Environment{ProductCatalogHost: constant.MockHost},
	}, mockClient, nil)

	return repo, mockClient
}

func TestCallNPWPVerificationAPI(t *testing.T) {
	t.Run(constant.TestCaseSuccess, func(t *testing.T) {
		mockData := model.ProCatAPIResponse[npwpVerificationRespData]{
			Success: true,
			Message: "success",
			Data:    npwpVerificationRespData{Name: constant.DummyName},
		}
		body, err := json.Marshal(mockData)
		require.NoError(t, err)

		resp := &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader(body)),
		}

		repo, mockClient := setupMockRepo(t, resp, nil)

		result, err := repo.NPWPVerificationAPI(constant.DummyAPIKey, constant.DummyJobId, &npwpVerificationRequest{})

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.True(t, result.Success)
		assert.Equal(t, constant.DummyName, result.Data.Name)
		mockClient.AssertExpectations(t)
	})

	t.Run(constant.TestCaseMarshalError, func(t *testing.T) {
		fakeMarshal := func(v any) ([]byte, error) {
			return nil, errors.New(constant.ErrInvalidRequestPayload)
		}

		repo := NewRepository(&application.Config{
			App: &application.Environment{ProductCatalogHost: constant.MockHost},
		}, &MockClient{}, fakeMarshal)

		result, err := repo.NPWPVerificationAPI(constant.DummyAPIKey, constant.DummyJobId, &npwpVerificationRequest{})
		assert.Nil(t, result)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), constant.ErrInvalidRequestPayload)
	})

	t.Run(constant.TestCaseNewRequestError, func(t *testing.T) {
		mockClient := new(MockClient)
		repo := NewRepository(&application.Config{
			App: &application.Environment{ProductCatalogHost: constant.MockInvalidHost},
		}, mockClient, nil)

		_, err := repo.NPWPVerificationAPI(constant.DummyAPIKey, constant.DummyJobId, &npwpVerificationRequest{})
		assert.Error(t, err)
	})

	t.Run(constant.TestCaseHTTPRequestError, func(t *testing.T) {
		expectedErr := errors.New(constant.ErrUpstreamUnavailable)

		repo, mockClient := setupMockRepo(t, nil, expectedErr)

		_, err := repo.NPWPVerificationAPI(constant.DummyAPIKey, constant.DummyJobId, &npwpVerificationRequest{})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), constant.ErrUpstreamUnavailable)
		mockClient.AssertExpectations(t)
	})

	t.Run(constant.TestCaseParseError, func(t *testing.T) {
		resp := &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(constant.InvalidJSON)),
		}

		repo, mockClient := setupMockRepo(t, resp, nil)

		result, err := repo.NPWPVerificationAPI(constant.DummyAPIKey, constant.DummyJobId, &npwpVerificationRequest{})
		assert.Nil(t, result)
		assert.Error(t, err)
		mockClient.AssertExpectations(t)
	})
}
