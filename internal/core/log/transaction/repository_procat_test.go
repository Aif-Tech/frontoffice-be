package transaction

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
	"github.com/stretchr/testify/require"
)

func TestGetLogTransByJobIdAPI(t *testing.T) {
	t.Run(constant.TestCaseSuccess, func(t *testing.T) {
		mockData := model.AifcoreAPIResponse[[]*LogTransProductCatalog]{
			Success: true,
			Data:    []*LogTransProductCatalog{},
		}
		body, err := json.Marshal(mockData)
		require.NoError(t, err)

		resp := &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader(body)),
		}

		repo, mockClient := setupMockRepo(t, resp, nil)

		result, err := repo.GetLogTransByJobIdAPI(constant.DummyJobId, constant.DummyCompanyId)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		mockClient.AssertExpectations(t)
	})

	t.Run(constant.TestCaseNewRequestError, func(t *testing.T) {
		mockClient := new(MockClient)
		repo := NewRepository(&application.Config{
			App: &application.Environment{AifcoreHost: constant.MockInvalidHost},
		}, mockClient, nil)

		result, err := repo.GetLogTransByJobIdAPI(constant.DummyJobId, constant.DummyCompanyId)

		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run(constant.TestCaseHTTPRequestError, func(t *testing.T) {
		expectedErr := errors.New(constant.ErrUpstreamUnavailable)

		repo, mockClient := setupMockRepo(t, nil, expectedErr)

		result, err := repo.GetLogTransByJobIdAPI(constant.DummyJobId, constant.DummyCompanyId)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), constant.ErrUpstreamUnavailable)
		mockClient.AssertExpectations(t)
	})

	t.Run(constant.TestCaseParseError, func(t *testing.T) {
		resp := &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(constant.InvalidJSON)),
		}

		repo, mockClient := setupMockRepo(t, resp, nil)

		result, err := repo.GetLogTransByJobIdAPI(constant.DummyJobId, constant.DummyCompanyId)

		assert.Nil(t, result)
		assert.Error(t, err)
		mockClient.AssertExpectations(t)
	})
}

func TestProcessedLogCountAPI(t *testing.T) {
	t.Run(constant.TestCaseSuccess, func(t *testing.T) {
		mockData := model.AifcoreAPIResponse[*getProcessedCountResp]{
			Success: true,
			Data: &getProcessedCountResp{
				ProcessedCount: uint(constant.DummyCount),
			},
		}
		body, err := json.Marshal(mockData)
		require.NoError(t, err)

		resp := &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader(body)),
		}

		repo, mockClient := setupMockRepo(t, resp, nil)

		result, err := repo.ProcessedLogCountAPI(constant.DummyJobId)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, uint(constant.DummyCount), uint(result.ProcessedCount))
		mockClient.AssertExpectations(t)
	})

	t.Run(constant.TestCaseNewRequestError, func(t *testing.T) {
		mockClient := new(MockClient)
		repo := NewRepository(&application.Config{
			App: &application.Environment{AifcoreHost: constant.MockInvalidHost},
		}, mockClient, nil)

		result, err := repo.ProcessedLogCountAPI(constant.DummyJobId)

		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run(constant.TestCaseHTTPRequestError, func(t *testing.T) {
		expectedErr := errors.New(constant.ErrUpstreamUnavailable)

		repo, mockClient := setupMockRepo(t, nil, expectedErr)

		result, err := repo.ProcessedLogCountAPI(constant.DummyJobId)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), constant.ErrUpstreamUnavailable)
		mockClient.AssertExpectations(t)
	})

	t.Run(constant.TestCaseParseError, func(t *testing.T) {
		resp := &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(constant.InvalidJSON)),
		}

		repo, mockClient := setupMockRepo(t, resp, nil)

		result, err := repo.ProcessedLogCountAPI(constant.DummyJobId)

		assert.Nil(t, result)
		assert.Error(t, err)
		mockClient.AssertExpectations(t)
	})
}

func TestCreateLogTransAPI(t *testing.T) {
	addLogReq := &LogTransProCatRequest{}

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

		err = repo.CreateLogTransAPI(addLogReq)

		assert.NoError(t, err)
		mockClient.AssertExpectations(t)
	})

	t.Run(constant.TestCaseMarshalError, func(t *testing.T) {
		fakeMarshal := func(v any) ([]byte, error) {
			return nil, errors.New(constant.ErrInvalidRequestPayload)
		}

		repo := NewRepository(&application.Config{
			App: &application.Environment{AifcoreHost: constant.MockHost},
		}, &MockClient{}, fakeMarshal)

		err := repo.CreateLogTransAPI(addLogReq)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), constant.ErrInvalidRequestPayload)
	})

	t.Run(constant.TestCaseNewRequestError, func(t *testing.T) {
		mockClient := new(MockClient)
		repo := NewRepository(&application.Config{
			App: &application.Environment{AifcoreHost: constant.MockInvalidHost},
		}, mockClient, nil)

		err := repo.CreateLogTransAPI(addLogReq)

		assert.Error(t, err)
	})

	t.Run(constant.TestCaseHTTPRequestError, func(t *testing.T) {
		expectedErr := errors.New(constant.ErrUpstreamUnavailable)

		repo, mockClient := setupMockRepo(t, nil, expectedErr)

		err := repo.CreateLogTransAPI(addLogReq)

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

		err := repo.CreateLogTransAPI(addLogReq)

		assert.Error(t, err)
		mockClient.AssertExpectations(t)
	})
}

func TestUpdateLogTransAPI(t *testing.T) {
	updateLogReq := map[string]interface{}{}

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

		err = repo.UpdateLogTransAPI(constant.DummyTransactionId, updateLogReq)

		assert.NoError(t, err)
		mockClient.AssertExpectations(t)
	})

	t.Run(constant.TestCaseMarshalError, func(t *testing.T) {
		fakeMarshal := func(v any) ([]byte, error) {
			return nil, errors.New(constant.ErrInvalidRequestPayload)
		}

		repo := NewRepository(&application.Config{
			App: &application.Environment{AifcoreHost: constant.MockHost},
		}, &MockClient{}, fakeMarshal)

		err := repo.UpdateLogTransAPI(constant.DummyTransactionId, updateLogReq)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), constant.ErrInvalidRequestPayload)
	})

	t.Run(constant.TestCaseNewRequestError, func(t *testing.T) {
		mockClient := new(MockClient)
		repo := NewRepository(&application.Config{
			App: &application.Environment{AifcoreHost: constant.MockInvalidHost},
		}, mockClient, nil)

		err := repo.UpdateLogTransAPI(constant.DummyTransactionId, updateLogReq)

		assert.Error(t, err)
	})

	t.Run(constant.TestCaseHTTPRequestError, func(t *testing.T) {
		expectedErr := errors.New(constant.ErrUpstreamUnavailable)

		repo, mockClient := setupMockRepo(t, nil, expectedErr)

		err := repo.UpdateLogTransAPI(constant.DummyTransactionId, updateLogReq)

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

		err := repo.UpdateLogTransAPI(constant.DummyTransactionId, updateLogReq)

		assert.Error(t, err)
		mockClient.AssertExpectations(t)
	})
}
