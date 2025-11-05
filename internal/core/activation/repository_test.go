package activation

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
		Env: &application.Environment{AifcoreHost: constant.MockHost},
	}, mockClient, nil)

	return repo, mockClient
}

func TestCallGetActivationTokenAPI(t *testing.T) {
	t.Run(constant.TestCaseSuccess, func(t *testing.T) {
		mockData := model.AifcoreAPIResponse[*MstActivationToken]{
			Success: true,
			Data: &MstActivationToken{
				Id:    constant.DummyId,
				Token: constant.DummyToken,
			},
		}
		body, err := json.Marshal(mockData)
		require.NoError(t, err)

		resp := &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader(body)),
		}

		repo, mockClient := setupMockRepo(t, resp, nil)

		result, err := repo.GetActivationTokenAPI(constant.DummyToken)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		mockClient.AssertExpectations(t)
	})

	t.Run(constant.TestCaseNewRequestError, func(t *testing.T) {
		mockClient := new(MockClient)
		repo := NewRepository(&application.Config{
			Env: &application.Environment{AifcoreHost: constant.MockInvalidHost},
		}, mockClient, nil)

		_, err := repo.GetActivationTokenAPI(constant.DummyToken)

		assert.Error(t, err)
	})

	t.Run(constant.TestCaseHTTPRequestError, func(t *testing.T) {
		expectedErr := errors.New(constant.ErrUpstreamUnavailable)

		repo, mockClient := setupMockRepo(t, nil, expectedErr)

		_, err := repo.GetActivationTokenAPI(constant.DummyToken)

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

		result, err := repo.GetActivationTokenAPI(constant.DummyToken)

		assert.Nil(t, result)
		assert.Error(t, err)
		mockClient.AssertExpectations(t)
	})
}

func TestCreateActivationTokenAPI(t *testing.T) {
	createTokenReq := &CreateActivationTokenRequest{
		Token: constant.DummyToken,
	}

	t.Run(constant.TestCaseSuccess, func(t *testing.T) {
		mockData := model.AifcoreAPIResponse[any]{
			Success: true,
		}
		body, err := json.Marshal(mockData)
		require.NoError(t, err)

		resp := &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader(body)),
		}

		repo, mockClient := setupMockRepo(t, resp, nil)

		err = repo.CreateActivationTokenAPI(constant.DummyMemberId, createTokenReq)

		assert.NoError(t, err)
		mockClient.AssertExpectations(t)
	})

	t.Run(constant.TestCaseMarshalError, func(t *testing.T) {
		fakeMarshal := func(v any) ([]byte, error) {
			return nil, errors.New(constant.ErrInvalidRequestPayload)
		}

		repo := NewRepository(&application.Config{
			Env: &application.Environment{AifcoreHost: constant.MockHost},
		}, &MockClient{}, fakeMarshal)

		err := repo.CreateActivationTokenAPI(constant.DummyMemberId, createTokenReq)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), constant.ErrInvalidRequestPayload)
	})

	t.Run(constant.TestCaseNewRequestError, func(t *testing.T) {
		mockClient := new(MockClient)
		repo := NewRepository(&application.Config{
			Env: &application.Environment{AifcoreHost: constant.MockInvalidHost},
		}, mockClient, nil)

		err := repo.CreateActivationTokenAPI(constant.DummyMemberId, createTokenReq)
		assert.Error(t, err)
	})

	t.Run(constant.TestCaseHTTPRequestError, func(t *testing.T) {
		expectedErr := errors.New(constant.ErrUpstreamUnavailable)

		repo, mockClient := setupMockRepo(t, nil, expectedErr)

		err := repo.CreateActivationTokenAPI(constant.DummyMemberId, createTokenReq)

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

		err := repo.CreateActivationTokenAPI(constant.DummyMemberId, createTokenReq)
		assert.Error(t, err)
		mockClient.AssertExpectations(t)
	})
}
