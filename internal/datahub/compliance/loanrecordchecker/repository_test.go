package loanrecordchecker

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
		Env: &application.Environment{ProductCatalogHost: constant.MockHost},
	}, mockClient, nil)

	return repo, mockClient
}

func TestCallLoanRecordCheckerAPI(t *testing.T) {
	t.Run(constant.TestCaseSuccess, func(t *testing.T) {
		mockData := model.ProCatAPIResponse[dataLoanRecord]{
			Success: true,
			Message: "success",
			Data:    dataLoanRecord{Remarks: "-", Status: "-"},
		}
		body, err := json.Marshal(mockData)
		require.NoError(t, err)

		resp := &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader(body)),
		}

		repo, mockClient := setupMockRepo(t, resp, nil)

		result, err := repo.LoanRecordCheckerAPI(constant.DummyAPIKey, constant.DummyJobId, constant.DummyMemberId, constant.DummyCompanyId, &loanRecordCheckerRequest{})

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.True(t, result.Success)
		assert.Equal(t, "-", result.Data.Status)
		assert.Equal(t, "-", result.Data.Remarks)
		mockClient.AssertExpectations(t)
	})

	t.Run(constant.TestCaseMarshalError, func(t *testing.T) {
		fakeMarshal := func(v any) ([]byte, error) {
			return nil, errors.New(constant.ErrFailedMarshalReq)
		}

		repo := NewRepository(&application.Config{
			Env: &application.Environment{ProductCatalogHost: constant.MockHost},
		}, &MockClient{}, fakeMarshal)

		result, err := repo.LoanRecordCheckerAPI(constant.DummyAPIKey, constant.DummyJobId, constant.DummyMemberId, constant.DummyCompanyId, &loanRecordCheckerRequest{})
		assert.Nil(t, result)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), constant.ErrFailedMarshalReq)
	})

	t.Run(constant.TestCaseNewRequestError, func(t *testing.T) {
		mockClient := new(MockClient)
		repo := NewRepository(&application.Config{
			Env: &application.Environment{ProductCatalogHost: constant.MockInvalidHost},
		}, mockClient, nil)

		_, err := repo.LoanRecordCheckerAPI(constant.DummyAPIKey, constant.DummyJobId, constant.DummyMemberId, constant.DummyCompanyId, &loanRecordCheckerRequest{})
		assert.Error(t, err)
	})

	t.Run(constant.TestCaseHTTPRequestError, func(t *testing.T) {
		expectedErr := errors.New(constant.ErrHTTPReqFailed)

		repo, mockClient := setupMockRepo(t, nil, expectedErr)

		req := &loanRecordCheckerRequest{}
		_, err := repo.LoanRecordCheckerAPI(constant.DummyAPIKey, constant.DummyJobId, constant.DummyMemberId, constant.DummyCompanyId, req)

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

		result, err := repo.LoanRecordCheckerAPI(constant.DummyAPIKey, constant.DummyJobId, constant.DummyMemberId, constant.DummyCompanyId, &loanRecordCheckerRequest{})
		assert.Nil(t, result)
		assert.Error(t, err)
		mockClient.AssertExpectations(t)
	})
}
