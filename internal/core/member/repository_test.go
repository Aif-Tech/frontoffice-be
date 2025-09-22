package member

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

func TestAddMemberAPI(t *testing.T) {
	addMemberReq := &RegisterMemberRequest{}

	t.Run(constant.TestCaseSuccess, func(t *testing.T) {
		mockData := model.AifcoreAPIResponse[*registerResponseData]{
			Success: true,
			Data: &registerResponseData{
				MemberId: constant.DummyIdInt,
			},
		}
		body, err := json.Marshal(mockData)
		require.NoError(t, err)

		resp := &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader(body)),
		}

		repo, mockClient := setupMockRepo(t, resp, nil)

		result, err := repo.AddMemberAPI(addMemberReq)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, uint(constant.DummyIdInt), result.MemberId)
		mockClient.AssertExpectations(t)
	})

	t.Run(constant.TestCaseNewRequestError, func(t *testing.T) {
		mockClient := new(MockClient)
		repo := NewRepository(&application.Config{
			Env: &application.Environment{AifcoreHost: constant.MockInvalidHost},
		}, mockClient, nil)

		result, err := repo.AddMemberAPI(addMemberReq)

		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run(constant.TestCaseHTTPRequestError, func(t *testing.T) {
		expectedErr := errors.New(constant.ErrHTTPReqFailed)

		repo, mockClient := setupMockRepo(t, nil, expectedErr)

		result, err := repo.AddMemberAPI(addMemberReq)

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

		result, err := repo.AddMemberAPI(addMemberReq)

		assert.Error(t, err)
		assert.Nil(t, result)
		mockClient.AssertExpectations(t)
	})
}

func TestGetMemberAPI(t *testing.T) {
	t.Run(constant.TestCaseSuccess, func(t *testing.T) {
		mockData := model.AifcoreAPIResponse[*MstMember]{
			Success: true,
			Data:    &MstMember{},
		}
		body, err := json.Marshal(mockData)
		require.NoError(t, err)

		resp := &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader(body)),
		}

		repo, mockClient := setupMockRepo(t, resp, nil)

		result, err := repo.GetMemberAPI(&MemberParams{})

		assert.NoError(t, err)
		assert.NotNil(t, result)
		mockClient.AssertExpectations(t)
	})

	t.Run(constant.TestCaseNewRequestError, func(t *testing.T) {
		mockClient := new(MockClient)
		repo := NewRepository(&application.Config{
			Env: &application.Environment{AifcoreHost: constant.MockInvalidHost},
		}, mockClient, nil)

		_, err := repo.GetMemberAPI(&MemberParams{})

		assert.Error(t, err)
	})

	t.Run(constant.TestCaseHTTPRequestError, func(t *testing.T) {
		expectedErr := errors.New(constant.ErrHTTPReqFailed)

		repo, mockClient := setupMockRepo(t, nil, expectedErr)

		_, err := repo.GetMemberAPI(&MemberParams{})

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

		result, err := repo.GetMemberAPI(&MemberParams{})

		assert.Nil(t, result)
		assert.Error(t, err)
		mockClient.AssertExpectations(t)
	})
}

func TestGetMemberListAPI(t *testing.T) {
	filter := &MemberParams{
		CompanyId: constant.CompanyId,
	}

	t.Run(constant.TestCaseSuccess, func(t *testing.T) {
		mockData := model.AifcoreAPIResponse[[]*MstMember]{
			Success: true,
			Data:    []*MstMember{},
			Meta:    &model.Meta{},
		}
		body, err := json.Marshal(mockData)
		require.NoError(t, err)

		resp := &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader(body)),
		}

		repo, mockClient := setupMockRepo(t, resp, nil)

		result, meta, err := repo.GetMemberListAPI(filter)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, meta)
		mockClient.AssertExpectations(t)
	})

	t.Run(constant.TestCaseNewRequestError, func(t *testing.T) {
		mockClient := new(MockClient)
		repo := NewRepository(&application.Config{
			Env: &application.Environment{AifcoreHost: constant.MockInvalidHost},
		}, mockClient, nil)

		result, meta, err := repo.GetMemberListAPI(filter)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Nil(t, meta)
	})

	t.Run(constant.TestCaseHTTPRequestError, func(t *testing.T) {
		expectedErr := errors.New(constant.ErrHTTPReqFailed)

		repo, mockClient := setupMockRepo(t, nil, expectedErr)

		result, meta, err := repo.GetMemberListAPI(filter)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Nil(t, meta)
		assert.Contains(t, err.Error(), constant.ErrHTTPReqFailed)
		mockClient.AssertExpectations(t)
	})

	t.Run(constant.TestCaseParseError, func(t *testing.T) {
		resp := &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(constant.InvalidJSON)),
		}

		repo, mockClient := setupMockRepo(t, resp, nil)

		result, meta, err := repo.GetMemberListAPI(filter)

		assert.Nil(t, result)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Nil(t, meta)
		mockClient.AssertExpectations(t)
	})
}

func TestUpdateMemberAPI(t *testing.T) {
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

		err = repo.UpdateMemberAPI(constant.DummyMemberId, map[string]interface{}{})

		assert.NoError(t, err)
		mockClient.AssertExpectations(t)
	})

	t.Run(constant.TestCaseMarshalError, func(t *testing.T) {
		fakeMarshal := func(v any) ([]byte, error) {
			return nil, errors.New(constant.ErrFailedMarshalReq)
		}

		repo := NewRepository(&application.Config{
			Env: &application.Environment{AifcoreHost: constant.MockHost},
		}, &MockClient{}, fakeMarshal)

		err := repo.UpdateMemberAPI(constant.DummyMemberId, map[string]interface{}{})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), constant.ErrFailedMarshalReq)
	})

	t.Run(constant.TestCaseNewRequestError, func(t *testing.T) {
		mockClient := new(MockClient)
		repo := NewRepository(&application.Config{
			Env: &application.Environment{AifcoreHost: constant.MockInvalidHost},
		}, mockClient, nil)

		err := repo.UpdateMemberAPI(constant.DummyMemberId, map[string]interface{}{})

		assert.Error(t, err)
	})

	t.Run(constant.TestCaseHTTPRequestError, func(t *testing.T) {
		expectedErr := errors.New(constant.ErrHTTPReqFailed)

		repo, mockClient := setupMockRepo(t, nil, expectedErr)

		err := repo.UpdateMemberAPI(constant.DummyMemberId, map[string]interface{}{})

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

		err := repo.UpdateMemberAPI(constant.DummyMemberId, map[string]interface{}{})

		assert.Error(t, err)
		mockClient.AssertExpectations(t)
	})
}

func TestDeleteMemberAPI(t *testing.T) {
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

		err = repo.DeleteMemberAPI(constant.DummyMemberId)

		assert.NoError(t, err)
		mockClient.AssertExpectations(t)
	})

	t.Run(constant.TestCaseNewRequestError, func(t *testing.T) {
		mockClient := new(MockClient)
		repo := NewRepository(&application.Config{
			Env: &application.Environment{AifcoreHost: constant.MockInvalidHost},
		}, mockClient, nil)

		err := repo.DeleteMemberAPI(constant.DummyMemberId)

		assert.Error(t, err)
	})

	t.Run(constant.TestCaseHTTPRequestError, func(t *testing.T) {
		expectedErr := errors.New(constant.ErrHTTPReqFailed)

		repo, mockClient := setupMockRepo(t, nil, expectedErr)

		err := repo.DeleteMemberAPI(constant.DummyMemberId)

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

		err := repo.DeleteMemberAPI(constant.DummyMemberId)

		assert.Error(t, err)
		mockClient.AssertExpectations(t)
	})
}
