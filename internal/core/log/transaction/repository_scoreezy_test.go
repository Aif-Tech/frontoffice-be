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

func TestGetLogsScoreezyByDateAPI(t *testing.T) {
	t.Run(constant.TestCaseSuccess, func(t *testing.T) {
		mockData := model.AifcoreAPIResponse[[]*LogTransScoreezy]{
			Success: true,
			Data:    []*LogTransScoreezy{},
		}
		body, err := json.Marshal(mockData)
		require.NoError(t, err)

		resp := &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader(body)),
		}

		repo, mockClient := setupMockRepo(t, resp, nil)

		result, err := repo.GetLogsScoreezyByDateAPI(constant.DummyJobId, constant.DummyCompanyId)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		mockClient.AssertExpectations(t)
	})

	t.Run(constant.TestCaseNewRequestError, func(t *testing.T) {
		mockClient := new(MockClient)
		repo := NewRepository(&application.Config{
			Env: &application.Environment{AifcoreHost: constant.MockInvalidHost},
		}, mockClient, nil)

		result, err := repo.GetLogsScoreezyByDateAPI(constant.DummyJobId, constant.DummyCompanyId)

		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run(constant.TestCaseHTTPRequestError, func(t *testing.T) {
		expectedErr := errors.New(constant.ErrHTTPReqFailed)

		repo, mockClient := setupMockRepo(t, nil, expectedErr)

		result, err := repo.GetLogsScoreezyByDateAPI(constant.DummyJobId, constant.DummyCompanyId)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), constant.ErrHTTPReqFailed)
		mockClient.AssertExpectations(t)
	})

	t.Run(constant.TestCaseParseError, func(t *testing.T) {
		resp := &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(constant.InvalidJSON)),
		}

		repo, mockClient := setupMockRepo(t, resp, nil)

		result, err := repo.GetLogsScoreezyByDateAPI(constant.DummyJobId, constant.DummyCompanyId)

		assert.Nil(t, result)
		assert.Error(t, err)
		mockClient.AssertExpectations(t)
	})
}

func TestGetLogsScoreezyByMonthAPI(t *testing.T) {
	t.Run(constant.TestCaseSuccess, func(t *testing.T) {
		mockData := model.AifcoreAPIResponse[[]*LogTransScoreezy]{
			Success: true,
			Data:    []*LogTransScoreezy{},
		}
		body, err := json.Marshal(mockData)
		require.NoError(t, err)

		resp := &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader(body)),
		}

		repo, mockClient := setupMockRepo(t, resp, nil)

		result, err := repo.GetLogsScoreezyByMonthAPI(constant.DummyCompanyId, constant.DummyMonth)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		mockClient.AssertExpectations(t)
	})

	t.Run(constant.TestCaseNewRequestError, func(t *testing.T) {
		mockClient := new(MockClient)
		repo := NewRepository(&application.Config{
			Env: &application.Environment{AifcoreHost: constant.MockInvalidHost},
		}, mockClient, nil)

		result, err := repo.GetLogsScoreezyByMonthAPI(constant.DummyCompanyId, constant.DummyMonth)

		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run(constant.TestCaseHTTPRequestError, func(t *testing.T) {
		expectedErr := errors.New(constant.ErrHTTPReqFailed)

		repo, mockClient := setupMockRepo(t, nil, expectedErr)

		result, err := repo.GetLogsScoreezyByMonthAPI(constant.DummyCompanyId, constant.DummyMonth)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), constant.ErrHTTPReqFailed)
		mockClient.AssertExpectations(t)
	})

	t.Run(constant.TestCaseParseError, func(t *testing.T) {
		resp := &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(constant.InvalidJSON)),
		}

		repo, mockClient := setupMockRepo(t, resp, nil)

		result, err := repo.GetLogsScoreezyByMonthAPI(constant.DummyCompanyId, constant.DummyMonth)

		assert.Nil(t, result)
		assert.Error(t, err)
		mockClient.AssertExpectations(t)
	})
}

func TestGetLogsScoreezyByDateRangeAPI(t *testing.T) {
	t.Run(constant.TestCaseSuccess, func(t *testing.T) {
		mockData := model.AifcoreAPIResponse[[]*LogTransScoreezy]{
			Success: true,
			Data:    []*LogTransScoreezy{},
		}
		body, err := json.Marshal(mockData)
		require.NoError(t, err)

		resp := &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader(body)),
		}

		repo, mockClient := setupMockRepo(t, resp, nil)

		result, err := repo.GetLogsScoreezyByDateRangeAPI(constant.DummyCompanyId, constant.DummyDate, constant.DummyDate)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		mockClient.AssertExpectations(t)
	})

	t.Run(constant.TestCaseNewRequestError, func(t *testing.T) {
		mockClient := new(MockClient)
		repo := NewRepository(&application.Config{
			Env: &application.Environment{AifcoreHost: constant.MockInvalidHost},
		}, mockClient, nil)

		result, err := repo.GetLogsScoreezyByDateRangeAPI(constant.DummyCompanyId, constant.DummyDate, constant.DummyDate)

		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run(constant.TestCaseHTTPRequestError, func(t *testing.T) {
		expectedErr := errors.New(constant.ErrHTTPReqFailed)

		repo, mockClient := setupMockRepo(t, nil, expectedErr)

		result, err := repo.GetLogsScoreezyByDateRangeAPI(constant.DummyCompanyId, constant.DummyDate, constant.DummyDate)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), constant.ErrHTTPReqFailed)
		mockClient.AssertExpectations(t)
	})

	t.Run(constant.TestCaseParseError, func(t *testing.T) {
		resp := &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(constant.InvalidJSON)),
		}

		repo, mockClient := setupMockRepo(t, resp, nil)

		result, err := repo.GetLogsScoreezyByDateRangeAPI(constant.DummyCompanyId, constant.DummyDate, constant.DummyDate)

		assert.Nil(t, result)
		assert.Error(t, err)
		mockClient.AssertExpectations(t)
	})
}

func TestGetLogsScoreezyAPI(t *testing.T) {
	t.Run(constant.TestCaseSuccess, func(t *testing.T) {
		mockData := model.AifcoreAPIResponse[[]*LogTransScoreezy]{
			Success: true,
			Data:    []*LogTransScoreezy{},
		}
		body, err := json.Marshal(mockData)
		require.NoError(t, err)

		resp := &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader(body)),
		}

		repo, mockClient := setupMockRepo(t, resp, nil)

		result, err := repo.GetLogsScoreezyAPI()

		assert.NoError(t, err)
		assert.NotNil(t, result)
		mockClient.AssertExpectations(t)
	})

	t.Run(constant.TestCaseNewRequestError, func(t *testing.T) {
		mockClient := new(MockClient)
		repo := NewRepository(&application.Config{
			Env: &application.Environment{AifcoreHost: constant.MockInvalidHost},
		}, mockClient, nil)

		result, err := repo.GetLogsScoreezyAPI()

		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run(constant.TestCaseHTTPRequestError, func(t *testing.T) {
		expectedErr := errors.New(constant.ErrHTTPReqFailed)

		repo, mockClient := setupMockRepo(t, nil, expectedErr)

		result, err := repo.GetLogsScoreezyAPI()

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), constant.ErrHTTPReqFailed)
		mockClient.AssertExpectations(t)
	})

	t.Run(constant.TestCaseParseError, func(t *testing.T) {
		resp := &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(constant.InvalidJSON)),
		}

		repo, mockClient := setupMockRepo(t, resp, nil)

		result, err := repo.GetLogsScoreezyAPI()

		assert.Nil(t, result)
		assert.Error(t, err)
		mockClient.AssertExpectations(t)
	})
}
