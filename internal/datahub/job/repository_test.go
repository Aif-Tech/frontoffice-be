package job

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
		App: &application.Environment{AifcoreHost: constant.MockHost},
	}, mockClient, nil)

	return repo, mockClient
}

func TestCallCreateProCatJob(t *testing.T) {
	t.Run(constant.TestCaseSuccess, func(t *testing.T) {
		mockData := model.AifcoreAPIResponse[any]{
			Success: true,
			Message: "Succeed to Request Data.",
			Data: createJobRespData{
				JobId: 1,
			},
		}
		body, err := json.Marshal(mockData)
		require.NoError(t, err)

		resp := &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader(body)),
		}

		repo, mockClient := setupMockRepo(t, resp, nil)

		result, err := repo.CreateJobAPI(&CreateJobRequest{})

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, result.JobId, uint(1))
		mockClient.AssertExpectations(t)
	})

	t.Run(constant.TestCaseMarshalError, func(t *testing.T) {
		fakeMarshal := func(v any) ([]byte, error) {
			return nil, errors.New(constant.ErrInvalidRequestPayload)
		}

		repo := NewRepository(&application.Config{
			App: &application.Environment{AifcoreHost: constant.MockHost},
		}, &MockClient{}, fakeMarshal)

		result, err := repo.CreateJobAPI(&CreateJobRequest{})

		assert.Nil(t, result)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), constant.ErrInvalidRequestPayload)
	})

	t.Run(constant.TestCaseNewRequestError, func(t *testing.T) {
		mockClient := new(MockClient)
		repo := NewRepository(&application.Config{
			App: &application.Environment{AifcoreHost: constant.MockInvalidHost},
		}, mockClient, nil)

		_, err := repo.CreateJobAPI(&CreateJobRequest{})
		assert.Error(t, err)
	})

	t.Run(constant.TestCaseHTTPRequestError, func(t *testing.T) {
		expectedErr := errors.New(constant.ErrUpstreamUnavailable)

		repo, mockClient := setupMockRepo(t, nil, expectedErr)

		req := &CreateJobRequest{}
		_, err := repo.CreateJobAPI(req)

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

		result, err := repo.CreateJobAPI(&CreateJobRequest{})
		assert.Nil(t, result)
		assert.Error(t, err)
		mockClient.AssertExpectations(t)
	})
}

func TestCallUpdateJob(t *testing.T) {
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

		err = repo.UpdateJobAPI(constant.DummyJobId, map[string]interface{}{})

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

		err := repo.UpdateJobAPI(constant.DummyJobId, map[string]interface{}{})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), constant.ErrInvalidRequestPayload)
	})

	t.Run(constant.TestCaseNewRequestError, func(t *testing.T) {
		mockClient := new(MockClient)
		repo := NewRepository(&application.Config{
			App: &application.Environment{AifcoreHost: constant.MockInvalidHost},
		}, mockClient, nil)

		err := repo.UpdateJobAPI(constant.DummyJobId, map[string]interface{}{})

		assert.Error(t, err)
	})

	t.Run(constant.TestCaseHTTPRequestError, func(t *testing.T) {
		expectedErr := errors.New(constant.ErrUpstreamUnavailable)

		repo, mockClient := setupMockRepo(t, nil, expectedErr)

		err := repo.UpdateJobAPI(constant.DummyJobId, map[string]interface{}{})

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

		err := repo.UpdateJobAPI(constant.DummyJobId, map[string]interface{}{})

		assert.Error(t, err)
		mockClient.AssertExpectations(t)
	})
}

func TestCallGetProCatJobAPI(t *testing.T) {
	filter := &logFilter{
		ProductSlug: constant.DummyProduct,
		AuthCtx: &model.AuthContext{
			UserId:    constant.DummyIdInt,
			CompanyId: constant.DummyIdInt,
			RoleId:    constant.DummyIdInt,
		},
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

		result, err := repo.GetJobsAPI(filter)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		mockClient.AssertExpectations(t)
	})

	t.Run(constant.TestCaseNewRequestError, func(t *testing.T) {
		mockClient := new(MockClient)
		repo := NewRepository(&application.Config{
			App: &application.Environment{AifcoreHost: constant.MockInvalidHost},
		}, mockClient, nil)

		_, err := repo.GetJobsAPI(filter)

		assert.Error(t, err)
	})

	t.Run(constant.TestCaseHTTPRequestError, func(t *testing.T) {
		expectedErr := errors.New(constant.ErrUpstreamUnavailable)

		repo, mockClient := setupMockRepo(t, nil, expectedErr)

		_, err := repo.GetJobsAPI(filter)

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

		result, err := repo.GetJobsAPI(filter)

		assert.Nil(t, result)
		assert.Error(t, err)
		mockClient.AssertExpectations(t)
	})
}

func TestCallGetProCatJobDetailAPI(t *testing.T) {
	filter := &logFilter{
		ProductSlug: constant.DummyProduct,
		AuthCtx: &model.AuthContext{
			UserId:    constant.DummyIdInt,
			CompanyId: constant.DummyIdInt,
			RoleId:    constant.DummyIdInt,
		},
	}

	t.Run(constant.TestCaseSuccess, func(t *testing.T) {
		mockData := model.AifcoreAPIResponse[any]{
			Success: true,
			Data: &jobDetailResponse{
				TotalData: 3,
			},
		}
		body, err := json.Marshal(mockData)
		require.NoError(t, err)

		resp := &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader(body)),
		}

		repo, mockClient := setupMockRepo(t, resp, nil)

		result, err := repo.GetJobDetailAPI(filter)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, result.Data.TotalData, int64(3))
		mockClient.AssertExpectations(t)
	})

	t.Run(constant.TestCaseNewRequestError, func(t *testing.T) {
		mockClient := new(MockClient)
		repo := NewRepository(&application.Config{
			App: &application.Environment{AifcoreHost: constant.MockInvalidHost},
		}, mockClient, nil)

		_, err := repo.GetJobDetailAPI(filter)

		assert.Error(t, err)
	})

	t.Run(constant.TestCaseHTTPRequestError, func(t *testing.T) {
		expectedErr := errors.New(constant.ErrUpstreamUnavailable)

		repo, mockClient := setupMockRepo(t, nil, expectedErr)

		_, err := repo.GetJobDetailAPI(filter)

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

		result, err := repo.GetJobDetailAPI(filter)

		assert.Nil(t, result)
		assert.Error(t, err)
		mockClient.AssertExpectations(t)
	})
}

func TestCallGetProCatJobDetailsAPI(t *testing.T) {
	filter := &logFilter{
		ProductSlug: constant.DummyProduct,
		AuthCtx: &model.AuthContext{
			UserId:    constant.DummyIdInt,
			CompanyId: constant.DummyIdInt,
			RoleId:    constant.DummyIdInt,
		},
	}

	t.Run(constant.TestCaseSuccess, func(t *testing.T) {
		mockData := model.AifcoreAPIResponse[any]{
			Success: true,
			Data: &jobDetailResponse{
				TotalData: 3,
			},
		}
		body, err := json.Marshal(mockData)
		require.NoError(t, err)

		resp := &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader(body)),
		}

		repo, mockClient := setupMockRepo(t, resp, nil)

		result, err := repo.GetJobsSummaryAPI(filter)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, result.Data.TotalData, int64(3))
		mockClient.AssertExpectations(t)
	})

	t.Run(constant.TestCaseNewRequestError, func(t *testing.T) {
		mockClient := new(MockClient)
		repo := NewRepository(&application.Config{
			App: &application.Environment{AifcoreHost: constant.MockInvalidHost},
		}, mockClient, nil)

		_, err := repo.GetJobsSummaryAPI(filter)

		assert.Error(t, err)
	})

	t.Run(constant.TestCaseHTTPRequestError, func(t *testing.T) {
		expectedErr := errors.New(constant.ErrUpstreamUnavailable)

		repo, mockClient := setupMockRepo(t, nil, expectedErr)

		_, err := repo.GetJobsSummaryAPI(filter)

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

		result, err := repo.GetJobsSummaryAPI(filter)

		assert.Nil(t, result)
		assert.Error(t, err)
		mockClient.AssertExpectations(t)
	})
}
