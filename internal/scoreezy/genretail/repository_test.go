package genretail

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
		Env: &application.Environment{
			ProductCatalogHost: constant.MockHost,
			ScoreezyHost:       constant.MockHost,
		},
	}, mockClient, nil)

	return repo, mockClient
}

func TestGenRetailV3API(t *testing.T) {
	t.Run(constant.TestCaseSuccess, func(t *testing.T) {
		mockData := model.ScoreezyAPIResponse[dataGenRetailV3]{
			Success: true,
			Data:    &dataGenRetailV3{Grade: "A"},
		}
		body, err := json.Marshal(mockData)
		require.NoError(t, err)

		resp := &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader(body)),
		}

		repo, mockClient := setupMockRepo(t, resp, nil)

		result, err := repo.GenRetailV3API(
			constant.DummyMemberId,
			&genRetailRequest{
				Name:     constant.DummyName,
				IdCardNo: constant.DummyNIK,
				PhoneNo:  constant.DummyPhoneNumber,
				LoanNo:   constant.DummyId,
			},
		)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.True(t, result.Success)
		assert.Equal(t, "A", result.Data.Grade)
		mockClient.AssertExpectations(t)
	})

	t.Run(constant.TestCaseMarshalError, func(t *testing.T) {
		fakeMarshal := func(v any) ([]byte, error) {
			return nil, errors.New(constant.ErrFailedMarshalReq)
		}

		repo := NewRepository(&application.Config{
			Env: &application.Environment{
				ScoreezyHost: constant.MockHost,
			},
		}, &MockClient{}, fakeMarshal)

		result, err := repo.GenRetailV3API(
			constant.DummyMemberId,
			&genRetailRequest{
				Name:     constant.DummyName,
				IdCardNo: constant.DummyNIK,
				PhoneNo:  constant.DummyPhoneNumber,
				LoanNo:   constant.DummyId,
			},
		)

		assert.Nil(t, result)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), constant.ErrFailedMarshalReq)
	})

	t.Run(constant.TestCaseNewRequestError, func(t *testing.T) {
		mockClient := new(MockClient)
		repo := NewRepository(
			&application.Config{
				Env: &application.Environment{
					ScoreezyHost: constant.MockInvalidHost,
				},
			}, mockClient, nil)

		_, err := repo.GenRetailV3API(
			constant.DummyMemberId,
			&genRetailRequest{
				Name:     constant.DummyName,
				IdCardNo: constant.DummyNIK,
				PhoneNo:  constant.DummyPhoneNumber,
				LoanNo:   constant.DummyId,
			},
		)

		assert.Error(t, err)
	})

	t.Run(constant.TestCaseHTTPRequestError, func(t *testing.T) {
		expectedErr := errors.New(constant.ErrHTTPReqFailed)

		repo, mockClient := setupMockRepo(t, nil, expectedErr)

		_, err := repo.GenRetailV3API(
			constant.DummyMemberId,
			&genRetailRequest{},
		)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), constant.ErrHTTPReqFailed)
		mockClient.AssertExpectations(t)
	})

	t.Run(constant.TestCaseParseError, func(t *testing.T) {
		resp := &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(constant.InvalidJSON)),
		}

		repo, mockClient := setupMockRepo(t, resp, nil)

		result, err := repo.GenRetailV3API(
			constant.DummyMemberId,
			&genRetailRequest{},
		)

		assert.Nil(t, result)
		assert.Error(t, err)
		mockClient.AssertExpectations(t)
	})
}

func TestGetLogsScoreezyAPI(t *testing.T) {
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

		result, err := repo.GetLogsScoreezyAPI(&filterLogs{})

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, http.StatusOK, result.StatusCode)
		mockClient.AssertExpectations(t)
	})

	t.Run(constant.TestCaseNewRequestError, func(t *testing.T) {
		mockClient := new(MockClient)
		repo := NewRepository(&application.Config{
			Env: &application.Environment{AifcoreHost: constant.MockInvalidHost},
		}, mockClient, nil)

		_, err := repo.GetLogsScoreezyAPI(&filterLogs{})

		assert.Error(t, err)
	})

	t.Run(constant.TestCaseHTTPRequestError, func(t *testing.T) {
		expectedErr := errors.New(constant.ErrHTTPReqFailed)

		repo, mockClient := setupMockRepo(t, nil, expectedErr)

		_, err := repo.GetLogsScoreezyAPI(&filterLogs{})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), constant.ErrHTTPReqFailed)
		mockClient.AssertExpectations(t)
	})

	t.Run(constant.TestCaseParseError, func(t *testing.T) {
		resp := &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(constant.InvalidJSON)),
		}

		repo, mockClient := setupMockRepo(t, resp, nil)

		result, err := repo.GetLogsScoreezyAPI(&filterLogs{})

		assert.Nil(t, result)
		assert.Error(t, err)
		mockClient.AssertExpectations(t)
	})
}

func TestGetLogsByRangeDateAPI(t *testing.T) {
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

		result, err := repo.GetLogsByRangeDateAPI(&filterLogs{})

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, http.StatusOK, result.StatusCode)
		mockClient.AssertExpectations(t)
	})

	t.Run(constant.TestCaseNewRequestError, func(t *testing.T) {
		mockClient := new(MockClient)
		repo := NewRepository(&application.Config{
			Env: &application.Environment{AifcoreHost: constant.MockInvalidHost},
		}, mockClient, nil)

		_, err := repo.GetLogsByRangeDateAPI(&filterLogs{})

		assert.Error(t, err)
	})

	t.Run(constant.TestCaseHTTPRequestError, func(t *testing.T) {
		expectedErr := errors.New(constant.ErrHTTPReqFailed)

		repo, mockClient := setupMockRepo(t, nil, expectedErr)

		_, err := repo.GetLogsByRangeDateAPI(&filterLogs{})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), constant.ErrHTTPReqFailed)
		mockClient.AssertExpectations(t)
	})

	t.Run(constant.TestCaseParseError, func(t *testing.T) {
		resp := &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(constant.InvalidJSON)),
		}

		repo, mockClient := setupMockRepo(t, resp, nil)

		result, err := repo.GetLogsByRangeDateAPI(&filterLogs{})

		assert.Nil(t, result)
		assert.Error(t, err)
		mockClient.AssertExpectations(t)
	})
}

func TestGetLogByTrxIdAPI(t *testing.T) {
	t.Run(constant.TestCaseSuccess, func(t *testing.T) {
		mockData := model.AifcoreAPIResponse[any]{
			Success: true,
			Data: &logTransScoreezy{
				TrxId: constant.DummyId,
			},
		}
		body, err := json.Marshal(mockData)
		require.NoError(t, err)

		resp := &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader(body)),
		}

		repo, mockClient := setupMockRepo(t, resp, nil)

		result, err := repo.GetLogByTrxIdAPI(&filterLogs{})

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, constant.DummyId, result.TrxId)
		mockClient.AssertExpectations(t)
	})

	t.Run(constant.TestCaseNewRequestError, func(t *testing.T) {
		mockClient := new(MockClient)
		repo := NewRepository(&application.Config{
			Env: &application.Environment{AifcoreHost: constant.MockInvalidHost},
		}, mockClient, nil)

		_, err := repo.GetLogByTrxIdAPI(&filterLogs{})

		assert.Error(t, err)
	})

	t.Run(constant.TestCaseHTTPRequestError, func(t *testing.T) {
		expectedErr := errors.New(constant.ErrHTTPReqFailed)

		repo, mockClient := setupMockRepo(t, nil, expectedErr)

		_, err := repo.GetLogByTrxIdAPI(&filterLogs{})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), constant.ErrHTTPReqFailed)
		mockClient.AssertExpectations(t)
	})

	t.Run(constant.TestCaseParseError, func(t *testing.T) {
		resp := &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(constant.InvalidJSON)),
		}

		repo, mockClient := setupMockRepo(t, resp, nil)

		result, err := repo.GetLogByTrxIdAPI(&filterLogs{})

		assert.Nil(t, result)
		assert.Error(t, err)
		mockClient.AssertExpectations(t)
	})
}
