package product

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
	}, mockClient)

	return repo, mockClient
}

func TestGetProductAPI(t *testing.T) {
	t.Run(constant.TestCaseSuccess, func(t *testing.T) {
		mockData := model.AifcoreAPIResponse[*productResponseData]{
			Success: true,
			Data: &productResponseData{
				ProductId:   constant.DummyIdInt,
				ProductSlug: constant.DummyProduct,
			},
		}
		body, err := json.Marshal(mockData)
		require.NoError(t, err)

		resp := &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader(body)),
		}

		repo, mockClient := setupMockRepo(t, resp, nil)

		result, err := repo.GetProductAPI(constant.DummyProduct)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, uint(constant.DummyIdInt), result.ProductId)
		assert.Equal(t, constant.DummyProduct, result.ProductSlug)
		mockClient.AssertExpectations(t)
	})

	t.Run(constant.TestCaseNewRequestError, func(t *testing.T) {
		mockClient := new(MockClient)
		repo := NewRepository(&application.Config{
			Env: &application.Environment{AifcoreHost: constant.MockInvalidHost},
		}, mockClient)

		result, err := repo.GetProductAPI(constant.DummyProduct)

		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run(constant.TestCaseHTTPRequestError, func(t *testing.T) {
		expectedErr := errors.New(constant.ErrHTTPReqFailed)

		repo, mockClient := setupMockRepo(t, nil, expectedErr)

		result, err := repo.GetProductAPI(constant.DummyProduct)

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

		result, err := repo.GetProductAPI(constant.DummyProduct)

		assert.Error(t, err)
		assert.Nil(t, result)
		mockClient.AssertExpectations(t)
	})
}
